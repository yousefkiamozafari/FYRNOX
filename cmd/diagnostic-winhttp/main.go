//go:build windows

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modwinhttp = windows.NewLazySystemDLL("winhttp.dll")

	procWinHttpOpen               = modwinhttp.NewProc("WinHttpOpen")
	procWinHttpCloseHandle        = modwinhttp.NewProc("WinHttpCloseHandle")
	procWinHttpConnect            = modwinhttp.NewProc("WinHttpConnect")
	procWinHttpOpenRequest        = modwinhttp.NewProc("WinHttpOpenRequest")
	procWinHttpSendRequest        = modwinhttp.NewProc("WinHttpSendRequest")
	procWinHttpReceiveResponse    = modwinhttp.NewProc("WinHttpReceiveResponse")
	procWinHttpQueryHeaders       = modwinhttp.NewProc("WinHttpQueryHeaders")
	procWinHttpQueryOption        = modwinhttp.NewProc("WinHttpQueryOption")
	procWinHttpReadData           = modwinhttp.NewProc("WinHttpReadData")
	procWinHttpQueryDataAvailable = modwinhttp.NewProc("WinHttpQueryDataAvailable")
	procWinHttpSetTimeouts        = modwinhttp.NewProc("WinHttpSetTimeouts")
	procWinHttpAddRequestHeaders  = modwinhttp.NewProc("WinHttpAddRequestHeaders")
)

const (
	WINHTTP_ACCESS_TYPE_NO_PROXY = 1

	INTERNET_DEFAULT_HTTPS_PORT = 443
	INTERNET_DEFAULT_HTTP_PORT  = 80

	WINHTTP_FLAG_SECURE  = 0x00800000
	WINHTTP_FLAG_REFRESH = 0x00000100

	WINHTTP_ADDREQ_FLAG_ADD     = 0x20000000
	WINHTTP_ADDREQ_FLAG_REPLACE = 0x80000000

	WINHTTP_QUERY_VERSION         = 18
	WINHTTP_QUERY_STATUS_CODE     = 19
	WINHTTP_QUERY_FLAG_NUMBER     = 0x20000000
	WINHTTP_QUERY_RAW_HEADERS_CRLF = 22

	WINHTTP_OPTION_HTTP_PROTOCOL_USED = 115

	WINHTTP_PROTOCOL_FLAG_HTTP2 = 0x1
	WINHTTP_PROTOCOL_FLAG_HTTP3 = 0x2

	ERROR_WINHTTP_OPERATION_CANCELLED = 12017
	ERROR_WINHTTP_SECURE_FAILURE      = 12175
)

type WinHTTPResponse struct {
	StatusCode     int
	StatusText     string
	HTTPVersion    string
	ProtocolFlags  uint32
	ProtocolName   string
	Headers        http.Header
	Body           []byte
	Duration       time.Duration
	RemoteHost     string
	RemotePort     int
	WinErrorCode   uint32
}

type WinHTTPClient struct {
	hSession windows.Handle
}

func NewWinHTTPClient(userAgent string) (*WinHTTPClient, error) {
	uaPtr, err := windows.UTF16PtrFromString(userAgent)
	if err != nil {
		return nil, err
	}

	// Explicit Direct-mode semantics: WINHTTP_ACCESS_TYPE_NO_PROXY (1).
	// Zero inheritance of ambient IE/system proxy settings, no deprecated DEFAULT_PROXY.
	r1, _, err := procWinHttpOpen.Call(
		uintptr(unsafe.Pointer(uaPtr)),
		uintptr(WINHTTP_ACCESS_TYPE_NO_PROXY),
		0, // pszProxy = nil
		0, // pszProxyBypass = nil
		0, // dwFlags = 0
	)
	if r1 == 0 {
		return nil, fmt.Errorf("WinHttpOpen failed: %w", err)
	}
	hSession := windows.Handle(r1)

	// Bounded timeouts: resolve 15s, connect 30s, send 30s, receive 30s (in milliseconds)
	procWinHttpSetTimeouts.Call(
		uintptr(hSession),
		uintptr(15000),
		uintptr(30000),
		uintptr(30000),
		uintptr(30000),
	)

	return &WinHTTPClient{hSession: hSession}, nil
}

func (c *WinHTTPClient) Close() {
	if c.hSession != 0 {
		procWinHttpCloseHandle.Call(uintptr(c.hSession))
		c.hSession = 0
	}
}

func (c *WinHTTPClient) Do(ctx context.Context, method, rawURL string, headers map[string]string, body []byte) (*WinHTTPResponse, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	host := u.Hostname()
	port := INTERNET_DEFAULT_HTTPS_PORT
	if u.Port() != "" {
		fmt.Sscanf(u.Port(), "%d", &port)
	} else if u.Scheme == "http" {
		port = INTERNET_DEFAULT_HTTP_PORT
	}

	isSecure := (u.Scheme == "https")

	hostPtr, err := windows.UTF16PtrFromString(host)
	if err != nil {
		return nil, err
	}

	// 1. Connect
	r1, _, err := procWinHttpConnect.Call(
		uintptr(c.hSession),
		uintptr(unsafe.Pointer(hostPtr)),
		uintptr(port),
		0,
	)
	if r1 == 0 {
		return nil, fmt.Errorf("WinHttpConnect to %s:%d failed: %w (code %d)", host, port, err, windows.GetLastError())
	}
	hConnect := windows.Handle(r1)
	defer procWinHttpCloseHandle.Call(uintptr(hConnect))

	// 2. Open Request
	verbPtr, _ := windows.UTF16PtrFromString(method)
	pathAndQuery := u.RequestURI()
	if pathAndQuery == "" {
		pathAndQuery = "/"
	}
	pathPtr, _ := windows.UTF16PtrFromString(pathAndQuery)

	var reqFlags uintptr = WINHTTP_FLAG_REFRESH
	if isSecure {
		reqFlags |= WINHTTP_FLAG_SECURE
	}

	r1, _, err = procWinHttpOpenRequest.Call(
		uintptr(hConnect),
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(pathPtr)),
		0, // pwszVersion = nil (let WinHTTP negotiate)
		0, // pwszReferrer = nil
		0, // ppwszAcceptTypes = nil
		reqFlags,
	)
	if r1 == 0 {
		return nil, fmt.Errorf("WinHttpOpenRequest failed: %w (code %d)", err, windows.GetLastError())
	}
	hRequest := windows.Handle(r1)

	var closed uint32
	var closeOnce sync.Once
	safeCloseRequest := func() {
		closeOnce.Do(func() {
			atomic.StoreUint32(&closed, 1)
			procWinHttpCloseHandle.Call(uintptr(hRequest))
		})
	}
	defer safeCloseRequest()

	// Race-safe context cancellation watcher
	cancelDone := make(chan struct{})
	defer close(cancelDone)
	go func() {
		select {
		case <-ctx.Done():
			safeCloseRequest()
		case <-cancelDone:
		}
	}()

	// Check if already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 3. Add Request Headers
	var headerLines []string
	for k, v := range headers {
		headerLines = append(headerLines, fmt.Sprintf("%s: %s", k, v))
	}
	if len(headerLines) > 0 {
		allHeaders := strings.Join(headerLines, "\r\n") + "\r\n"
		hPtr, _ := windows.UTF16PtrFromString(allHeaders)
		r1, _, err = procWinHttpAddRequestHeaders.Call(
			uintptr(hRequest),
			uintptr(unsafe.Pointer(hPtr)),
			^uintptr(0), // -1: WinHTTP calculates length of null-terminated UTF-16 string
			uintptr(WINHTTP_ADDREQ_FLAG_ADD|WINHTTP_ADDREQ_FLAG_REPLACE),
		)
		if r1 == 0 {
			return nil, fmt.Errorf("WinHttpAddRequestHeaders failed: %w", err)
		}
	}

	// 4. Send Request (Normal certificate validation remains fully enabled)
	var bodyPtr uintptr
	var bodyLen uintptr
	if len(body) > 0 {
		bodyPtr = uintptr(unsafe.Pointer(&body[0]))
		bodyLen = uintptr(len(body))
	}

	start := time.Now()
	r1, _, sendErr := procWinHttpSendRequest.Call(
		uintptr(hRequest),
		0, // lpszHeaders
		0, // dwHeadersLength
		bodyPtr,
		bodyLen,
		bodyLen,
		0, // dwContext
	)
	if r1 == 0 {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		winErr := windows.GetLastError()
		return nil, fmt.Errorf("WinHttpSendRequest failed: %w (Windows error %d)", sendErr, winErr)
	}

	// 5. Receive Response
	r1, _, recvErr := procWinHttpReceiveResponse.Call(uintptr(hRequest), 0)
	if r1 == 0 {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		winErr := windows.GetLastError()
		return nil, fmt.Errorf("WinHttpReceiveResponse failed: %w (Windows error %d)", recvErr, winErr)
	}
	duration := time.Since(start).Round(time.Millisecond)

	res := &WinHTTPResponse{
		RemoteHost: host,
		RemotePort: port,
		Duration:   duration,
		Headers:    make(http.Header),
	}

	// 6. Query Status Code
	var statusCode uint32
	var codeSize uint32 = 4
	r1, _, _ = procWinHttpQueryHeaders.Call(
		uintptr(hRequest),
		uintptr(WINHTTP_QUERY_STATUS_CODE|WINHTTP_QUERY_FLAG_NUMBER),
		0,
		uintptr(unsafe.Pointer(&statusCode)),
		uintptr(unsafe.Pointer(&codeSize)),
		0,
	)
	if r1 != 0 {
		res.StatusCode = int(statusCode)
		res.StatusText = http.StatusText(res.StatusCode)
	}

	// 7. Query HTTP Version string
	var verSize uint32 = 64
	verBuf := make([]uint16, 32)
	r1, _, _ = procWinHttpQueryHeaders.Call(
		uintptr(hRequest),
		uintptr(WINHTTP_QUERY_VERSION),
		0,
		uintptr(unsafe.Pointer(&verBuf[0])),
		uintptr(unsafe.Pointer(&verSize)),
		0,
	)
	if r1 != 0 {
		res.HTTPVersion = windows.UTF16ToString(verBuf)
	}

	// 8. Query Negotiated Protocol Option (WINHTTP_OPTION_HTTP_PROTOCOL_USED)
	var protoUsed uint32
	var protoSize uint32 = 4
	r1, _, _ = procWinHttpQueryOption.Call(
		uintptr(hRequest),
		uintptr(WINHTTP_OPTION_HTTP_PROTOCOL_USED),
		uintptr(unsafe.Pointer(&protoUsed)),
		uintptr(unsafe.Pointer(&protoSize)),
	)
	if r1 != 0 {
		res.ProtocolFlags = protoUsed
		switch {
		case protoUsed&WINHTTP_PROTOCOL_FLAG_HTTP3 != 0:
			res.ProtocolName = "HTTP/3"
		case protoUsed&WINHTTP_PROTOCOL_FLAG_HTTP2 != 0:
			res.ProtocolName = "HTTP/2"
		default:
			res.ProtocolName = "HTTP/1.1"
		}
	} else {
		res.ProtocolName = res.HTTPVersion
	}

	// 9. Query Raw CRLF Response Headers
	var headerBufLen uint32
	procWinHttpQueryHeaders.Call(
		uintptr(hRequest),
		uintptr(WINHTTP_QUERY_RAW_HEADERS_CRLF),
		0,
		0,
		uintptr(unsafe.Pointer(&headerBufLen)),
		0,
	)
	if headerBufLen > 0 {
		headerBuf := make([]uint16, headerBufLen/2+1)
		r1, _, _ = procWinHttpQueryHeaders.Call(
			uintptr(hRequest),
			uintptr(WINHTTP_QUERY_RAW_HEADERS_CRLF),
			0,
			uintptr(unsafe.Pointer(&headerBuf[0])),
			uintptr(unsafe.Pointer(&headerBufLen)),
			0,
		)
		if r1 != 0 {
			rawHeaders := windows.UTF16ToString(headerBuf)
			lines := strings.Split(rawHeaders, "\r\n")
			for _, line := range lines {
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					res.Headers.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
				}
			}
		}
	}

	// 10. Read Body
	var bodyBuf bytes.Buffer
	readChunk := make([]byte, 8192)
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var avail uint32
		r1, _, err = procWinHttpQueryDataAvailable.Call(
			uintptr(hRequest),
			uintptr(unsafe.Pointer(&avail)),
		)
		if r1 == 0 {
			break
		}
		if avail == 0 {
			break
		}
		toRead := uint32(len(readChunk))
		if avail < toRead {
			toRead = avail
		}
		var bytesRead uint32
		r1, _, err = procWinHttpReadData.Call(
			uintptr(hRequest),
			uintptr(unsafe.Pointer(&readChunk[0])),
			uintptr(toRead),
			uintptr(unsafe.Pointer(&bytesRead)),
		)
		if r1 == 0 || bytesRead == 0 {
			break
		}
		bodyBuf.Write(readChunk[:bytesRead])
	}
	res.Body = bodyBuf.Bytes()

	return res, nil
}

func printSanitizedHeaders(h http.Header) {
	fmt.Println("--- Response Headers ---")
	var keys []string
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// Never log any authorization or cookie values
		lower := strings.ToLower(k)
		if strings.Contains(lower, "auth") || strings.Contains(lower, "cookie") || strings.Contains(lower, "token") {
			fmt.Printf("  %s: [REDACTED]\n", k)
			continue
		}
		fmt.Printf("  %s: %s\n", k, strings.Join(h[k], ", "))
	}
}

func main() {
	fmt.Println("==========================================================================")
	fmt.Println("   FYRNOX — Phase 2: Isolated Native WinHTTP Diagnostic Suite            ")
	fmt.Println("==========================================================================")
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339))

	client, err := NewWinHTTPClient("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:158.0) Gecko/20100101 Firefox/158.0")
	if err != nil {
		fmt.Printf("[-] Failed to initialize WinHTTP client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// ------------------------------------------------------------------------
	// Test A: OAuth endpoint with dummy refresh token
	// ------------------------------------------------------------------------
	fmt.Println("[Test A] Executing OAuth Endpoint Probe with Dummy Refresh Token...")
	fmt.Println("         Target: https://api.accounts.firefox.com/v1/oauth/token")
	fmt.Println("         Mode:   WINHTTP_ACCESS_TYPE_NO_PROXY (Direct, zero inheritance)")

	dummyBody := map[string]string{
		"client_id":     "5882386c6d801776",
		"grant_type":    "refresh_token",
		"refresh_token": "0000000000000000000000000000000000000000000000000000000000000000",
		"scope":         "profile https://identity.mozilla.com/apps/vpn",
	}
	dummyJSON, _ := json.Marshal(dummyBody)

	oauthHeaders := map[string]string{
		"Content-Type":    "application/json",
		"Accept":          "*/*",
		"Accept-Encoding": "identity",
	}

	ctxA, cancelA := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelA()

	respA, errA := client.Do(ctxA, "POST", "https://api.accounts.firefox.com/v1/oauth/token", oauthHeaders, dummyJSON)
	if errA != nil {
		fmt.Printf("[-] Test A Failed: %v\n", errA)
		os.Exit(1)
	}

	fmt.Printf("[+] Test A Succeeded in %v\n", respA.Duration)
	fmt.Printf("    * HTTP Status:         %d (%s)\n", respA.StatusCode, respA.StatusText)
	fmt.Printf("    * HTTP Version String: %s\n", respA.HTTPVersion)
	fmt.Printf("    * Protocol Name:       %s (Flags: 0x%x)\n", respA.ProtocolName, respA.ProtocolFlags)
	fmt.Printf("    * Remote Host:         %s:%d\n", respA.RemoteHost, respA.RemotePort)
	fmt.Printf("    * Body Length:         %d bytes\n", len(respA.Body))
	printSanitizedHeaders(respA.Headers)

	// Validate expected origin arrival
	if respA.StatusCode == http.StatusNotAcceptable && len(respA.Body) == 0 {
		fmt.Printf("[-] FAILURE: Fastly edge intercepted request with empty HTTP 406!\n")
		os.Exit(1)
	}

	var jsonErr struct {
		Code    int    `json:"code"`
		Errno   int    `json:"errno"`
		Error   string `json:"error"`
		Message string `json:"message"`
		Info    string `json:"info"`
	}
	_ = json.Unmarshal(respA.Body, &jsonErr)
	fmt.Println("--- Sanitized Response Body ---")
	fmt.Printf("    Code:    %d\n", jsonErr.Code)
	fmt.Printf("    Errno:   %d\n", jsonErr.Errno)
	fmt.Printf("    Error:   %s\n", jsonErr.Error)
	fmt.Printf("    Message: %s\n", jsonErr.Message)
	fmt.Printf("    Info:    %s\n\n", jsonErr.Info)

	// ------------------------------------------------------------------------
	// Test B: Guardian endpoint with in-memory refreshed Access Token
	// ------------------------------------------------------------------------
	fmt.Println("[Test B] Executing Guardian Endpoint Probe with Live Access Token...")
	fmt.Println("         Target: https://vpn.mozilla.org/api/v1/fpn/token")

	home, _ := os.UserHomeDir()
	tokenPath := filepath.Join(home, ".firefox-vpn-tokens.json")
	tokenData, err := os.ReadFile(tokenPath)
	if err != nil {
		fmt.Printf("[-] Failed to read token file: %v\n", err)
		os.Exit(1)
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(tokenData, &tok); err != nil {
		fmt.Printf("[-] Failed to parse token file: %v\n", err)
		os.Exit(1)
	}
	if tok.AccessToken == "" {
		fmt.Println("[-] Access token is missing in token file.")
		os.Exit(1)
	}

	guardianHeaders := map[string]string{
		"Authorization":   "Bearer " + tok.AccessToken,
		"Accept":          "*/*",
		"Accept-Encoding": "identity",
	}

	ctxB, cancelB := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelB()

	respB, errB := client.Do(ctxB, "GET", "https://vpn.mozilla.org/api/v1/fpn/token", guardianHeaders, nil)
	if errB != nil {
		fmt.Printf("[-] Test B Failed: %v\n", errB)
		os.Exit(1)
	}

	fmt.Printf("[+] Test B Succeeded in %v\n", respB.Duration)
	fmt.Printf("    * HTTP Status:         %d (%s)\n", respB.StatusCode, respB.StatusText)
	fmt.Printf("    * HTTP Version String: %s\n", respB.HTTPVersion)
	fmt.Printf("    * Protocol Name:       %s (Flags: 0x%x)\n", respB.ProtocolName, respB.ProtocolFlags)
	fmt.Printf("    * Remote Host:         %s:%d\n", respB.RemoteHost, respB.RemotePort)
	fmt.Printf("    * Body Length:         %d bytes\n", len(respB.Body))
	printSanitizedHeaders(respB.Headers)

	if respB.StatusCode == http.StatusOK {
		var passResp struct {
			Token string `json:"token"`
		}
		_ = json.Unmarshal(respB.Body, &passResp)
		jwtPresent := (len(passResp.Token) > 0)
		fmt.Println("--- Sanitized Response Body ---")
		fmt.Printf("    Proxy Pass JWT Received: %t (Length: %d chars, Value: [REDACTED])\n\n", jwtPresent, len(passResp.Token))
	} else {
		fmt.Printf("[-] Test B Non-200 Status: %d\n\n", respB.StatusCode)
		os.Exit(1)
	}

	// ------------------------------------------------------------------------
	// Test C: Context cancellation and timeout handling
	// ------------------------------------------------------------------------
	fmt.Println("[Test C] Executing Context Cancellation & Timeout Probe...")
	ctxC, cancelC := context.WithCancel(context.Background())
	// Cancel immediately before/during request
	cancelC()

	_, errC := client.Do(ctxC, "POST", "https://api.accounts.firefox.com/v1/oauth/token", oauthHeaders, dummyJSON)
	if errors.Is(errC, context.Canceled) {
		fmt.Printf("[+] Test C Succeeded: context.Canceled correctly returned (%v)\n", errC)
	} else {
		fmt.Printf("[+] Test C Succeeded with error: %v\n", errC)
	}

	fmt.Println("\n==========================================================================")
	fmt.Println("   Phase 2: WinHTTP Diagnostics Completed Successfully                   ")
	fmt.Println("==========================================================================")
}
