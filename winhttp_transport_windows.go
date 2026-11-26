//go:build windows

package vpnclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ─── Windows API proc bindings ─────────────────────────────────────────────

var (
	modwinhttp  = windows.NewLazySystemDLL("winhttp.dll")
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procWinHttpOpen               = modwinhttp.NewProc("WinHttpOpen")
	procWinHttpCloseHandle        = modwinhttp.NewProc("WinHttpCloseHandle")
	procWinHttpConnect            = modwinhttp.NewProc("WinHttpConnect")
	procWinHttpOpenRequest        = modwinhttp.NewProc("WinHttpOpenRequest")
	procWinHttpSendRequest        = modwinhttp.NewProc("WinHttpSendRequest")
	procWinHttpReceiveResponse    = modwinhttp.NewProc("WinHttpReceiveResponse")
	procWinHttpQueryHeaders       = modwinhttp.NewProc("WinHttpQueryHeaders")
	procWinHttpQueryOption        = modwinhttp.NewProc("WinHttpQueryOption")
	procWinHttpSetOption          = modwinhttp.NewProc("WinHttpSetOption")
	procWinHttpQueryDataAvailable = modwinhttp.NewProc("WinHttpQueryDataAvailable")
	procWinHttpReadData           = modwinhttp.NewProc("WinHttpReadData")
	procWinHttpSetTimeouts        = modwinhttp.NewProc("WinHttpSetTimeouts")
	procWinHttpAddRequestHeaders  = modwinhttp.NewProc("WinHttpAddRequestHeaders")
	procWinHttpSetStatusCallback  = modwinhttp.NewProc("WinHttpSetStatusCallback")

	procRtlMoveMemory = modkernel32.NewProc("RtlMoveMemory")
)

const (
	winhttpAccessTypeNoProxy    = 1 // WINHTTP_ACCESS_TYPE_NO_PROXY  — no ambient/IE proxy inheritance
	winhttpAccessTypeNamedProxy = 3 // WINHTTP_ACCESS_TYPE_NAMED_PROXY

	internetDefaultHTTPSPort = 443
	internetDefaultHTTPPort  = 80

	winhttpFlagAsync   = 0x10000000
	winhttpFlagSecure  = 0x00800000
	winhttpFlagRefresh = 0x00000100

	winhttpAddReqFlagAdd     = 0x20000000
	winhttpAddReqFlagReplace = 0x80000000

	winhttpOptionDisableFeature   = 63
	winhttpDisableCookies         = 0x00000001
	winhttpDisableRedirects       = 0x00000002
	winhttpOptionContextValue     = 45
	winhttpOptionProxyUsername    = 0x1002
	winhttpOptionProxyPassword    = 0x1003
	winhttpOptionHTTPProtocolUsed = 115

	winhttpQueryVersion        = 18
	winhttpQueryStatusCode     = 19
	winhttpQueryFlagNumber     = 0x20000000
	winhttpQueryRawHeadersCRLF = 22

	winhttpProtocolFlagHTTP2 = 0x1
	winhttpProtocolFlagHTTP3 = 0x2

	winhttpCallbackStatusHandleClosing       = 0x00000800
	winhttpCallbackStatusHeadersAvailable    = 0x00020000
	winhttpCallbackStatusDataAvailable       = 0x00040000
	winhttpCallbackStatusReadComplete        = 0x00080000
	winhttpCallbackStatusRequestError        = 0x00200000
	winhttpCallbackStatusSendRequestComplete = 0x00400000

	// WINHTTP_CALLBACK_FLAG_HANDLES includes HANDLE_CREATED (0x400) | HANDLE_CLOSING (0x800).
	winhttpCallbackMask = 0x00000400 |
		winhttpCallbackStatusHandleClosing |
		winhttpCallbackStatusHeadersAvailable |
		winhttpCallbackStatusDataAvailable |
		winhttpCallbackStatusReadComplete |
		winhttpCallbackStatusRequestError |
		winhttpCallbackStatusSendRequestComplete

	winhttpInvalidStatusCallback = ^uintptr(0)

	errorWinHTTPTimeout            = 12002
	errorWinHTTPOperationCancelled = 12017

	defaultResolveTimeoutMs = 15000
	defaultConnectTimeoutMs = 30000
	defaultSendTimeoutMs    = 30000
	defaultReceiveTimeoutMs = 30000
	winhttpReadChunkSize    = 32 * 1024
)

// winhttpAsyncResult mirrors WinHTTP's WINHTTP_ASYNC_RESULT on both amd64 and 386.
type winhttpAsyncResult struct {
	dwResult uintptr
	dwError  uint32
}

// ─── Global Async Callback Registry ────────────────────────────────────────

var (
	winhttpCallbackOnce sync.Once
	winhttpCallbackPtr  uintptr

	asyncReqSeq   atomic.Uintptr
	asyncReqTable sync.Map // uintptr -> *winHTTPAsyncReq
)

func initWinHTTPCallback() uintptr {
	winhttpCallbackOnce.Do(func() {
		winhttpCallbackPtr = windows.NewCallback(winhttpStatusCallback)
	})
	return winhttpCallbackPtr
}

func winhttpStatusCallback(
	hInternet,
	dwContext,
	dwInternetStatus,
	lpvStatusInfo,
	dwStatusInfoLen uintptr,
) uintptr {
	if dwContext == 0 {
		return 0
	}
	v, ok := asyncReqTable.Load(dwContext)
	if !ok {
		return 0
	}
	req := v.(*winHTTPAsyncReq)

	switch uint32(dwInternetStatus) {
	case winhttpCallbackStatusSendRequestComplete:
		req.deliverOpResult(winhttpCallbackStatusSendRequestComplete, winhttpOpResult{})

	case winhttpCallbackStatusHeadersAvailable:
		req.deliverOpResult(winhttpCallbackStatusHeadersAvailable, winhttpOpResult{})

	case winhttpCallbackStatusDataAvailable:
		var avail uint32
		if lpvStatusInfo != 0 && dwStatusInfoLen >= 4 {
			procRtlMoveMemory.Call(
				uintptr(unsafe.Pointer(&avail)),
				lpvStatusInfo,
				4,
			)
		}
		req.deliverOpResult(winhttpCallbackStatusDataAvailable, winhttpOpResult{
			availBytes: avail,
		})

	case winhttpCallbackStatusReadComplete:
		req.deliverOpResult(winhttpCallbackStatusReadComplete, winhttpOpResult{
			bytesRead: uint32(dwStatusInfoLen),
		})

	case winhttpCallbackStatusRequestError:
		var res winhttpAsyncResult
		if lpvStatusInfo != 0 && dwStatusInfoLen >= unsafe.Sizeof(res) {
			procRtlMoveMemory.Call(
				uintptr(unsafe.Pointer(&res)),
				lpvStatusInfo,
				unsafe.Sizeof(res),
			)
		}
		req.deliverError(windows.Errno(res.dwError))

	case winhttpCallbackStatusHandleClosing:
		// Guaranteed final callback for hReq: unpin async buffers, delete registry
		// entry, and wake anyone waiting on teardown.
		req.finalizeCleanup()
	}

	return 0
}

// ─── Routing Selection ─────────────────────────────────────────────────────

// shouldUseWinHTTP returns true when the Windows-native WinHTTP transport
// should handle the given ControlPlaneTransportOptions.
//
// Eligible modes:
//   - Direct connections (ProxyURL == nil, no CustomDialer, no TLSFragment, no Roots, no ForceStandardTLS).
//   - HTTP forward proxies (ProxyURL.Scheme == "http", no CustomDialer/TLSFragment/Roots/ForceStandardTLS).
//
// Preserved on the Go uTLS/crypto-tls path:
//   - SOCKS5 / SOCKS5h proxies (unsupported natively by WinHTTP).
//   - HTTPS TLS forward proxies (kept on Go httpConnectHandshake).
//   - Custom dialers, TLS fragmentation, custom Roots CertPool, and ForceStandardTLS.
func shouldUseWinHTTP(opts ControlPlaneTransportOptions) bool {
	if opts.CustomDialer != nil || opts.TLSFragment || opts.ForceStandardTLS || opts.Roots != nil {
		return false
	}
	if opts.ProxyURL != nil {
		return strings.EqualFold(opts.ProxyURL.Scheme, "http")
	}
	return true
}

// IsWinHTTPTransport reports whether rt is a Windows-native WinHTTP RoundTripper.
func IsWinHTTPTransport(rt http.RoundTripper) bool {
	_, ok := rt.(*winHTTPTransport)
	return ok
}

// ─── Session Ownership & Refcounting ───────────────────────────────────────

type winHTTPSession struct {
	mu      sync.Mutex
	handle  windows.Handle
	refs    int
	retired bool
}

func (s *winHTTPSession) acquire() (windows.Handle, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.retired || s.handle == 0 {
		return 0, false
	}
	s.refs++
	return s.handle, true
}

func (s *winHTTPSession) release() {
	var toClose windows.Handle
	s.mu.Lock()
	if s.refs > 0 {
		s.refs--
	}
	if s.refs == 0 && s.retired && s.handle != 0 {
		toClose = s.handle
		s.handle = 0
	}
	s.mu.Unlock()

	if toClose != 0 {
		procWinHttpCloseHandle.Call(uintptr(toClose))
	}
}

func (s *winHTTPSession) retire() {
	var toClose windows.Handle
	s.mu.Lock()
	s.retired = true
	if s.refs == 0 && s.handle != 0 {
		toClose = s.handle
		s.handle = 0
	}
	s.mu.Unlock()

	if toClose != 0 {
		procWinHttpCloseHandle.Call(uintptr(toClose))
	}
}

// ─── Error Type ────────────────────────────────────────────────────────────

type winHTTPError struct {
	op      string
	err     error
	timeout bool
}

func (e *winHTTPError) Error() string {
	return fmt.Sprintf("%s: %v", e.op, e.err)
}

func (e *winHTTPError) Unwrap() error {
	return e.err
}

func (e *winHTTPError) Timeout() bool {
	return e.timeout
}

func (e *winHTTPError) Temporary() bool {
	return e.timeout
}

func toWinErrno(err error) windows.Errno {
	if err == nil {
		return 0
	}
	var we windows.Errno
	if errors.As(err, &we) {
		return we
	}
	var se syscall.Errno
	if errors.As(err, &se) {
		return windows.Errno(se)
	}
	return 0
}

func classifyWinHTTPError(op string, winErr windows.Errno, ctx context.Context) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if winErr == errorWinHTTPTimeout {
		return &winHTTPError{
			op:      op,
			err:     context.DeadlineExceeded,
			timeout: true,
		}
	}
	if winErr == errorWinHTTPOperationCancelled {
		return &winHTTPError{
			op:  op,
			err: net.ErrClosed,
		}
	}
	return &winHTTPError{
		op:  op,
		err: winErr,
	}
}

// ─── winHTTPTransport ──────────────────────────────────────────────────────

// winHTTPTransport is an http.RoundTripper backed by asynchronous Windows
// WinHTTP (WINHTTP_FLAG_ASYNC + Schannel). It bypasses Fastly WAF HTTP 406
// while keeping full crypto certificate validation enabled.
type winHTTPTransport struct {
	proxyName string // empty = WINHTTP_ACCESS_TYPE_NO_PROXY
	proxyUser string
	proxyPass string

	mu      sync.Mutex
	session *winHTTPSession
}

// newWinHTTPTransport builds a winHTTPTransport using explicit Direct-mode or
// HTTP forward-proxy semantics with zero ambient proxy inheritance.
func newWinHTTPTransport(opts ControlPlaneTransportOptions) (http.RoundTripper, error) {
	t := &winHTTPTransport{}
	if opts.ProxyURL != nil {
		if opts.ProxyURL.Host == "" {
			return nil, errors.New("winhttp: proxy URL missing host")
		}
		t.proxyName = opts.ProxyURL.Host
		if opts.ProxyAuth != nil {
			t.proxyUser = opts.ProxyAuth.User
			t.proxyPass = opts.ProxyAuth.Password
		} else if opts.ProxyURL.User != nil {
			t.proxyUser = opts.ProxyURL.User.Username()
			t.proxyPass, _ = opts.ProxyURL.User.Password()
		}
	}

	if err := t.ensureSessionLockedless(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *winHTTPTransport) ensureSessionLockedless() error {
	sess, _, err := t.acquireSession()
	if err != nil {
		return err
	}
	sess.release()
	return nil
}

func (t *winHTTPTransport) acquireSession() (*winHTTPSession, windows.Handle, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.session != nil {
		if h, ok := t.session.acquire(); ok {
			return t.session, h, nil
		}
		t.session = nil
	}

	sess, err := t.openSession()
	if err != nil {
		return nil, 0, err
	}
	h, ok := sess.acquire()
	if !ok {
		return nil, 0, errors.New("winhttp: newly opened session inactive")
	}
	t.session = sess
	return sess, h, nil
}

func (t *winHTTPTransport) openSession() (*winHTTPSession, error) {
	var accessType uintptr = winhttpAccessTypeNoProxy
	var proxyStrPtr *uint16
	var bypassStrPtr *uint16
	if t.proxyName != "" {
		accessType = winhttpAccessTypeNamedProxy
		ps, err := windows.UTF16PtrFromString(t.proxyName)
		if err != nil {
			return nil, fmt.Errorf("winhttp: encoding proxy host: %w", err)
		}
		proxyStrPtr = ps
		// Prevent WinHTTP from silently bypassing configured proxies for loopback/local test or tunnel targets.
		bs, err := windows.UTF16PtrFromString("<-loopback>")
		if err != nil {
			return nil, fmt.Errorf("winhttp: encoding proxy bypass: %w", err)
		}
		bypassStrPtr = bs
	}

	uaPtr, err := windows.UTF16PtrFromString(FirefoxUserAgent())
	if err != nil {
		return nil, fmt.Errorf("winhttp: encoding user-agent: %w", err)
	}

	r1, _, winErr := procWinHttpOpen.Call(
		uintptr(unsafe.Pointer(uaPtr)),
		accessType,
		uintptr(unsafe.Pointer(proxyStrPtr)),
		uintptr(unsafe.Pointer(bypassStrPtr)),
		winhttpFlagAsync,
	)
	if r1 == 0 {
		return nil, fmt.Errorf("WinHttpOpen failed: %w", winErr)
	}
	hSession := windows.Handle(r1)

	procWinHttpSetTimeouts.Call(
		uintptr(hSession),
		defaultResolveTimeoutMs,
		defaultConnectTimeoutMs,
		defaultSendTimeoutMs,
		defaultReceiveTimeoutMs,
	)

	cbPtr := initWinHTTPCallback()
	cbRet, _, cbErr := procWinHttpSetStatusCallback.Call(
		uintptr(hSession),
		cbPtr,
		winhttpCallbackMask,
		0,
	)
	if cbRet == winhttpInvalidStatusCallback {
		procWinHttpCloseHandle.Call(uintptr(hSession))
		return nil, fmt.Errorf("WinHttpSetStatusCallback failed: %w", cbErr)
	}

	return &winHTTPSession{handle: hSession}, nil
}

// CloseIdleConnections retires the current idle WinHTTP session pool without
// disrupting active requests or preventing future requests on the transport.
func (t *winHTTPTransport) CloseIdleConnections() {
	t.mu.Lock()
	sess := t.session
	t.session = nil
	t.mu.Unlock()

	if sess != nil {
		sess.retire()
	}
}

// ─── Per-Request Handle & Async State ──────────────────────────────────────

type winhttpOpResult struct {
	availBytes uint32
	bytesRead  uint32
	errErrno   windows.Errno
}

type winHTTPAsyncReq struct {
	id   uintptr
	sess *winHTTPSession

	// handleMu serializes every WinHttp* call against closeAsync so no WinHTTP
	// API is ever invoked on hReq after WinHttpCloseHandle(hReq) begins.
	// Lock ordering DAG: handleMu -> stateMu (never reversed; callback takes only stateMu).
	handleMu     sync.Mutex
	handleClosed bool
	hConn        windows.Handle
	hReq         windows.Handle

	stateMu        sync.Mutex
	expectedStatus uint32
	opWait         chan winhttpOpResult
	pinner         runtime.Pinner
	reqBody        []byte
	readBuf        []byte
	readBufPinned  bool
	cleanupOnce    sync.Once
	closingCh      chan struct{}

	teardownOnce sync.Once
}

func newWinHTTPAsyncReq(sess *winHTTPSession, hConn, hReq windows.Handle) (*winHTTPAsyncReq, error) {
	id := asyncReqSeq.Add(1)
	if id == 0 {
		id = asyncReqSeq.Add(1)
	}
	r := &winHTTPAsyncReq{
		id:        id,
		sess:      sess,
		hConn:     hConn,
		hReq:      hReq,
		closingCh: make(chan struct{}),
	}
	asyncReqTable.Store(id, r)

	ctxVal := id
	r1, _, winErr := procWinHttpSetOption.Call(
		uintptr(hReq),
		uintptr(winhttpOptionContextValue),
		uintptr(unsafe.Pointer(&ctxVal)),
		unsafe.Sizeof(ctxVal),
	)
	if r1 == 0 {
		asyncReqTable.Delete(id)
		return nil, fmt.Errorf("WinHttpSetOption(CONTEXT_VALUE) failed: %w", winErr)
	}
	return r, nil
}

func (r *winHTTPAsyncReq) deliverOpResult(status uint32, res winhttpOpResult) {
	r.stateMu.Lock()
	var ch chan winhttpOpResult
	if r.expectedStatus == status && r.opWait != nil {
		ch = r.opWait
		r.opWait = nil
		r.expectedStatus = 0
	}
	r.stateMu.Unlock()

	if ch != nil {
		select {
		case ch <- res:
		default:
		}
	}
}

func (r *winHTTPAsyncReq) deliverError(errno windows.Errno) {
	r.stateMu.Lock()
	ch := r.opWait
	r.opWait = nil
	r.expectedStatus = 0
	r.stateMu.Unlock()

	if ch != nil {
		select {
		case ch <- winhttpOpResult{errErrno: errno}:
		default:
		}
	}
}

func (r *winHTTPAsyncReq) finalizeCleanup() {
	r.cleanupOnce.Do(func() {
		r.stateMu.Lock()
		ch := r.opWait
		r.opWait = nil
		r.expectedStatus = 0
		r.pinner.Unpin()
		r.reqBody = nil
		r.readBuf = nil
		r.readBufPinned = false
		r.stateMu.Unlock()

		asyncReqTable.Delete(r.id)
		if ch != nil {
			select {
			case ch <- winhttpOpResult{errErrno: errorWinHTTPOperationCancelled}:
			default:
			}
		}
		close(r.closingCh)
	})
}

func (r *winHTTPAsyncReq) closeAsync() {
	r.handleMu.Lock()
	if r.handleClosed {
		r.handleMu.Unlock()
		return
	}
	r.handleClosed = true
	hReq := r.hReq
	var closedOK bool
	if hReq != 0 {
		r1, _, _ := procWinHttpCloseHandle.Call(uintptr(hReq))
		closedOK = (r1 != 0)
	}
	r.handleMu.Unlock()

	if !closedOK {
		r.finalizeCleanup()
	}
}

// startTeardown closes hRequest asynchronously and releases hConn + session ref
// ONLY after WINHTTP_CALLBACK_STATUS_HANDLE_CLOSING confirms WinHTTP is done.
func (r *winHTTPAsyncReq) startTeardown(wait bool) {
	r.closeAsync()
	r.teardownOnce.Do(func() {
		releaseParent := func() {
			<-r.closingCh
			if r.hConn != 0 {
				procWinHttpCloseHandle.Call(uintptr(r.hConn))
				r.hConn = 0
			}
			if r.sess != nil {
				r.sess.release()
				r.sess = nil
			}
		}
		if wait {
			releaseParent()
		} else {
			go releaseParent()
		}
	})
}

func (r *winHTTPAsyncReq) runAsyncStep(
	ctx context.Context,
	opName string,
	expectedStatus uint32,
	prepare func() (uintptr, windows.Errno),
) (winhttpOpResult, error) {
	r.handleMu.Lock()
	if r.handleClosed {
		r.handleMu.Unlock()
		if ctx != nil && ctx.Err() != nil {
			return winhttpOpResult{}, ctx.Err()
		}
		return winhttpOpResult{}, net.ErrClosed
	}

	waitCh := make(chan winhttpOpResult, 1)
	r.stateMu.Lock()
	r.expectedStatus = expectedStatus
	r.opWait = waitCh
	r.stateMu.Unlock()

	r1, winErr := prepare()
	if r1 == 0 {
		r.stateMu.Lock()
		if r.opWait == waitCh {
			r.opWait = nil
			r.expectedStatus = 0
		}
		r.stateMu.Unlock()
		r.handleMu.Unlock()
		return winhttpOpResult{}, classifyWinHTTPError(opName, winErr, ctx)
	}
	r.handleMu.Unlock()

	var ctxDone <-chan struct{}
	if ctx != nil {
		ctxDone = ctx.Done()
	}

	select {
	case res := <-waitCh:
		if res.errErrno != 0 {
			return winhttpOpResult{}, classifyWinHTTPError(opName, res.errErrno, ctx)
		}
		return res, nil
	case <-r.closingCh:
		if ctx != nil && ctx.Err() != nil {
			return winhttpOpResult{}, ctx.Err()
		}
		return winhttpOpResult{}, net.ErrClosed
	case <-ctxDone:
		r.startTeardown(false)
		return winhttpOpResult{}, ctx.Err()
	}
}

// ─── RoundTrip ─────────────────────────────────────────────────────────────

func (t *winHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := ctx.Err(); err != nil {
		if req.Body != nil {
			_ = req.Body.Close()
		}
		return nil, err
	}

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("winhttp: reading request body: %w", err)
		}
	}

	resolveMs, connectMs, sendMs, recvMs, err := boundedTimeoutsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sess, sessHandle, err := t.acquireSession()
	if err != nil {
		return nil, err
	}

	u := req.URL
	host := u.Hostname()
	port := internetDefaultHTTPSPort
	if p := u.Port(); p != "" {
		if parsedPort, perr := strconv.Atoi(p); perr == nil && parsedPort > 0 && parsedPort <= 65535 {
			port = parsedPort
		}
	} else if strings.EqualFold(u.Scheme, "http") {
		port = internetDefaultHTTPPort
	}
	isSecure := strings.EqualFold(u.Scheme, "https")

	hostPtr, err := windows.UTF16PtrFromString(host)
	if err != nil {
		sess.release()
		return nil, fmt.Errorf("winhttp: encoding host: %w", err)
	}

	r1, _, winErr := procWinHttpConnect.Call(
		uintptr(sessHandle),
		uintptr(unsafe.Pointer(hostPtr)),
		uintptr(port),
		0,
	)
	if r1 == 0 {
		sess.release()
		return nil, classifyWinHTTPError("WinHttpConnect", toWinErrno(winErr), ctx)
	}
	hConnect := windows.Handle(r1)

	verbPtr, err := windows.UTF16PtrFromString(req.Method)
	if err != nil {
		procWinHttpCloseHandle.Call(uintptr(hConnect))
		sess.release()
		return nil, fmt.Errorf("winhttp: encoding method: %w", err)
	}
	pathAndQuery := u.RequestURI()
	if pathAndQuery == "" {
		pathAndQuery = "/"
	}
	pathPtr, err := windows.UTF16PtrFromString(pathAndQuery)
	if err != nil {
		procWinHttpCloseHandle.Call(uintptr(hConnect))
		sess.release()
		return nil, fmt.Errorf("winhttp: encoding path: %w", err)
	}

	var reqFlags uintptr = winhttpFlagRefresh
	if isSecure {
		reqFlags |= winhttpFlagSecure
	}

	r1, _, winErr = procWinHttpOpenRequest.Call(
		uintptr(hConnect),
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(pathPtr)),
		0, // pwszVersion = nil
		0, // pwszReferrer = nil
		0, // ppwszAcceptTypes = nil
		reqFlags,
	)
	if r1 == 0 {
		procWinHttpCloseHandle.Call(uintptr(hConnect))
		sess.release()
		return nil, classifyWinHTTPError("WinHttpOpenRequest", toWinErrno(winErr), ctx)
	}
	hRequest := windows.Handle(r1)

	asyncReq, err := newWinHTTPAsyncReq(sess, hConnect, hRequest)
	if err != nil {
		procWinHttpCloseHandle.Call(uintptr(hRequest))
		procWinHttpCloseHandle.Call(uintptr(hConnect))
		sess.release()
		return nil, err
	}

	// Disable automatic WinHTTP cookies & redirects so http.Client owns Jar and CheckRedirect.
	disableFlags := uint32(winhttpDisableCookies | winhttpDisableRedirects)
	if r1, _, optErr := procWinHttpSetOption.Call(
		uintptr(hRequest),
		uintptr(winhttpOptionDisableFeature),
		uintptr(unsafe.Pointer(&disableFlags)),
		unsafe.Sizeof(disableFlags),
	); r1 == 0 {
		asyncReq.startTeardown(false)
		return nil, fmt.Errorf("WinHttpSetOption(DISABLE_FEATURE) failed: %w", optErr)
	}

	// Apply per-request bounded timeouts derived from ctx.Deadline().
	procWinHttpSetTimeouts.Call(
		uintptr(hRequest),
		uintptr(resolveMs),
		uintptr(connectMs),
		uintptr(sendMs),
		uintptr(recvMs),
	)

	// Set proxy credentials on hRequest when configured (length in UTF-16 characters per WinHTTP spec).
	if t.proxyName != "" && t.proxyUser != "" {
		if err := setWinHTTPCharOption(hRequest, winhttpOptionProxyUsername, t.proxyUser); err != nil {
			asyncReq.startTeardown(false)
			return nil, fmt.Errorf("WinHttpSetOption(PROXY_USERNAME) failed: %w", err)
		}
		if t.proxyPass != "" {
			if err := setWinHTTPCharOption(hRequest, winhttpOptionProxyPassword, t.proxyPass); err != nil {
				asyncReq.startTeardown(false)
				return nil, fmt.Errorf("WinHttpSetOption(PROXY_PASSWORD) failed: %w", err)
			}
		}
	}

	// Apply explicit request headers.
	if err := applyRequestHeadersWinHTTP(hRequest, req); err != nil {
		asyncReq.startTeardown(false)
		return nil, err
	}

	// Early cancellation check before starting async send.
	if err := ctx.Err(); err != nil {
		asyncReq.startTeardown(false)
		return nil, err
	}

	// Pin request body until HANDLE_CLOSING releases asyncReq.
	var bodyPtr uintptr
	var bodyLen uintptr
	if len(bodyBytes) > 0 {
		asyncReq.stateMu.Lock()
		asyncReq.reqBody = bodyBytes
		asyncReq.pinner.Pin(&bodyBytes[0])
		asyncReq.stateMu.Unlock()
		bodyPtr = uintptr(unsafe.Pointer(&bodyBytes[0]))
		bodyLen = uintptr(len(bodyBytes))
	}

	if _, err := asyncReq.runAsyncStep(ctx, "WinHttpSendRequest", winhttpCallbackStatusSendRequestComplete, func() (uintptr, windows.Errno) {
		ok, _, callErr := procWinHttpSendRequest.Call(
			uintptr(hRequest),
			0, // lpszHeaders
			0, // dwHeadersLength
			bodyPtr,
			bodyLen,
			bodyLen,
			asyncReq.id,
		)
		if ok == 0 {
			return 0, toWinErrno(callErr)
		}
		return ok, 0
	}); err != nil {
		asyncReq.startTeardown(false)
		return nil, err
	}

	if _, err := asyncReq.runAsyncStep(ctx, "WinHttpReceiveResponse", winhttpCallbackStatusHeadersAvailable, func() (uintptr, windows.Errno) {
		ok, _, callErr := procWinHttpReceiveResponse.Call(uintptr(hRequest), 0)
		if ok == 0 {
			return 0, toWinErrno(callErr)
		}
		return ok, 0
	}); err != nil {
		asyncReq.startTeardown(false)
		return nil, err
	}

	resp, err := buildHTTPResponseFromWinHTTP(asyncReq, req)
	if err != nil {
		asyncReq.startTeardown(false)
		return nil, err
	}
	return resp, nil
}

func boundedTimeoutsFromContext(ctx context.Context) (resolveMs, connectMs, sendMs, recvMs int, err error) {
	resolveMs = defaultResolveTimeoutMs
	connectMs = defaultConnectTimeoutMs
	sendMs = defaultSendTimeoutMs
	recvMs = defaultReceiveTimeoutMs

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return resolveMs, connectMs, sendMs, recvMs, nil
	}
	rem := time.Until(deadline)
	if rem <= 0 {
		return 0, 0, 0, 0, context.DeadlineExceeded
	}
	remMs := int(rem.Milliseconds())
	if remMs < 1 {
		// WinHttpSetTimeouts interprets 0 as infinite; clamp strictly >= 1 ms.
		remMs = 1
	}
	if resolveMs > remMs {
		resolveMs = remMs
	}
	if connectMs > remMs {
		connectMs = remMs
	}
	if sendMs > remMs {
		sendMs = remMs
	}
	if recvMs > remMs {
		recvMs = remMs
	}
	return resolveMs, connectMs, sendMs, recvMs, nil
}

func setWinHTTPCharOption(h windows.Handle, option uintptr, val string) error {
	encoded := utf16.Encode([]rune(val + "\x00"))
	charLen := uintptr(len(encoded) - 1) // exclude null terminator per WinHttpSetOption docs
	r1, _, winErr := procWinHttpSetOption.Call(
		uintptr(h),
		option,
		uintptr(unsafe.Pointer(&encoded[0])),
		charLen,
	)
	if r1 == 0 {
		return winErr
	}
	return nil
}

func applyRequestHeadersWinHTTP(hRequest windows.Handle, req *http.Request) error {
	var headerLines []string
	if req.Host != "" && !strings.EqualFold(req.Host, req.URL.Host) {
		headerLines = append(headerLines, "Host: "+req.Host)
	}

	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		vs := req.Header[k]
		if len(vs) == 0 {
			continue
		}
		// Coalesce multi-value headers per RFC 9110 so ADD|REPLACE never drops earlier values.
		sep := ", "
		if strings.EqualFold(k, "Cookie") {
			sep = "; "
		}
		headerLines = append(headerLines, k+": "+strings.Join(vs, sep))
	}

	if len(headerLines) == 0 {
		return nil
	}
	allHeaders := strings.Join(headerLines, "\r\n") + "\r\n"
	hPtr, err := windows.UTF16PtrFromString(allHeaders)
	if err != nil {
		return fmt.Errorf("winhttp: encoding headers: %w", err)
	}
	r1, _, winErr := procWinHttpAddRequestHeaders.Call(
		uintptr(hRequest),
		uintptr(unsafe.Pointer(hPtr)),
		^uintptr(0), // -1: null-terminated UTF-16 string
		uintptr(winhttpAddReqFlagAdd|winhttpAddReqFlagReplace),
	)
	if r1 == 0 {
		return fmt.Errorf("WinHttpAddRequestHeaders failed: %w", winErr)
	}
	return nil
}

func buildHTTPResponseFromWinHTTP(asyncReq *winHTTPAsyncReq, req *http.Request) (*http.Response, error) {
	asyncReq.handleMu.Lock()
	if asyncReq.handleClosed {
		asyncReq.handleMu.Unlock()
		return nil, net.ErrClosed
	}
	hRequest := asyncReq.hReq

	var statusCode uint32
	var codeSize uint32 = 4
	r1, _, winErr := procWinHttpQueryHeaders.Call(
		uintptr(hRequest),
		uintptr(winhttpQueryStatusCode|winhttpQueryFlagNumber),
		0,
		uintptr(unsafe.Pointer(&statusCode)),
		uintptr(unsafe.Pointer(&codeSize)),
		0,
	)
	if r1 == 0 {
		asyncReq.handleMu.Unlock()
		return nil, fmt.Errorf("WinHttpQueryHeaders(STATUS_CODE) failed: %w", winErr)
	}

	proto := "HTTP/1.1"
	var verSize uint32 = 64
	verBuf := make([]uint16, 33)
	r1, _, _ = procWinHttpQueryHeaders.Call(
		uintptr(hRequest),
		uintptr(winhttpQueryVersion),
		0,
		uintptr(unsafe.Pointer(&verBuf[0])),
		uintptr(unsafe.Pointer(&verSize)),
		0,
	)
	if r1 != 0 {
		if parsed := windows.UTF16ToString(verBuf); parsed != "" {
			proto = parsed
		}
	}

	var protoUsed uint32
	var protoSize uint32 = 4
	r1, _, _ = procWinHttpQueryOption.Call(
		uintptr(hRequest),
		uintptr(winhttpOptionHTTPProtocolUsed),
		uintptr(unsafe.Pointer(&protoUsed)),
		uintptr(unsafe.Pointer(&protoSize)),
	)
	if r1 != 0 {
		switch {
		case protoUsed&winhttpProtocolFlagHTTP3 != 0:
			proto = "HTTP/3.0"
		case protoUsed&winhttpProtocolFlagHTTP2 != 0:
			proto = "HTTP/2.0"
		}
	}

	var hdrBufLen uint32
	procWinHttpQueryHeaders.Call(
		uintptr(hRequest),
		uintptr(winhttpQueryRawHeadersCRLF),
		0,
		0,
		uintptr(unsafe.Pointer(&hdrBufLen)),
		0,
	)
	var rawHeaders string
	if hdrBufLen > 0 {
		hdrBuf := make([]uint16, hdrBufLen/2+2)
		r1, _, _ = procWinHttpQueryHeaders.Call(
			uintptr(hRequest),
			uintptr(winhttpQueryRawHeadersCRLF),
			0,
			uintptr(unsafe.Pointer(&hdrBuf[0])),
			uintptr(unsafe.Pointer(&hdrBufLen)),
			0,
		)
		if r1 != 0 {
			rawHeaders = windows.UTF16ToString(hdrBuf)
		}
	}
	asyncReq.handleMu.Unlock()

	statusText, respHeaders := parseWinHTTPRawHeaders(rawHeaders, int(statusCode))

	protoMajor, protoMinor := 1, 1
	if proto == "HTTP/2.0" || proto == "HTTP/2" {
		protoMajor, protoMinor = 2, 0
	} else if proto == "HTTP/3.0" || proto == "HTTP/3" {
		protoMajor, protoMinor = 3, 0
	}

	contentLen := int64(-1)
	if !strings.EqualFold(respHeaders.Get("Transfer-Encoding"), "chunked") {
		if clStr := respHeaders.Get("Content-Length"); clStr != "" {
			if n, perr := strconv.ParseInt(strings.TrimSpace(clStr), 10, 64); perr == nil && n >= 0 {
				contentLen = n
			}
		}
	}

	resp := &http.Response{
		Status:        statusText,
		StatusCode:    int(statusCode),
		Proto:         proto,
		ProtoMajor:    protoMajor,
		ProtoMinor:    protoMinor,
		Header:        respHeaders,
		ContentLength: contentLen,
		Request:       req,
		Body: &winHTTPResponseBody{
			req: asyncReq,
			ctx: req.Context(),
		},
	}
	return resp, nil
}

func parseWinHTTPRawHeaders(rawHeaders string, statusCode int) (string, http.Header) {
	statusText := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	respHeaders := make(http.Header)

	if rawHeaders == "" {
		return statusText, respHeaders
	}

	// Normalize header block so net/textproto parses folded and multi-value headers identically to net/http.
	normalized := strings.ReplaceAll(rawHeaders, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 {
		return statusText, respHeaders
	}

	firstLine := strings.TrimSpace(lines[0])
	if _, rest, ok := strings.Cut(firstLine, " "); ok && rest != "" {
		statusText = strings.TrimSpace(rest)
	}

	if len(lines) > 1 {
		headerBlock := strings.Join(lines[1:], "\r\n")
		if !strings.HasSuffix(headerBlock, "\r\n\r\n") {
			headerBlock = strings.TrimRight(headerBlock, "\r\n") + "\r\n\r\n"
		}
		tp := textproto.NewReader(bufio.NewReader(strings.NewReader(headerBlock)))
		if mimeHdr, err := tp.ReadMIMEHeader(); err == nil {
			respHeaders = http.Header(mimeHdr)
		}
	}

	return statusText, respHeaders
}

// ─── winHTTPResponseBody ───────────────────────────────────────────────────

type winHTTPResponseBody struct {
	req    *winHTTPAsyncReq
	ctx    context.Context
	readMu sync.Mutex
	eof    bool
	closed atomic.Bool
}

func (b *winHTTPResponseBody) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	b.readMu.Lock()
	defer b.readMu.Unlock()

	if b.eof {
		return 0, io.EOF
	}
	if b.closed.Load() {
		return 0, net.ErrClosed
	}
	if b.ctx != nil && b.ctx.Err() != nil {
		return 0, b.ctx.Err()
	}

	// Allocate and pin one reusable internal chunk buffer owned by winHTTPAsyncReq
	// until HANDLE_CLOSING so caller slices are never touched across cancellations.
	b.req.stateMu.Lock()
	if !b.req.readBufPinned {
		b.req.readBuf = make([]byte, winhttpReadChunkSize)
		b.req.pinner.Pin(&b.req.readBuf[0])
		b.req.readBufPinned = true
	}
	readBuf := b.req.readBuf
	b.req.stateMu.Unlock()

	// Query available bytes asynchronously first so WinHttpReadData never blocks waiting
	// to fill the entire caller slice when smaller streamed chunks arrive.
	availRes, err := b.req.runAsyncStep(b.ctx, "WinHttpQueryDataAvailable", winhttpCallbackStatusDataAvailable, func() (uintptr, windows.Errno) {
		ok, _, callErr := procWinHttpQueryDataAvailable.Call(
			uintptr(b.req.hReq),
			0, // lpdwNumberOfBytesAvailable MUST be NULL under WINHTTP_FLAG_ASYNC
		)
		if ok == 0 {
			return 0, toWinErrno(callErr)
		}
		return ok, 0
	})
	if err != nil {
		return 0, err
	}
	if availRes.availBytes == 0 {
		b.eof = true
		return 0, io.EOF
	}

	toRead := len(p)
	if int(availRes.availBytes) < toRead {
		toRead = int(availRes.availBytes)
	}
	if toRead > len(readBuf) {
		toRead = len(readBuf)
	}

	res, err := b.req.runAsyncStep(b.ctx, "WinHttpReadData", winhttpCallbackStatusReadComplete, func() (uintptr, windows.Errno) {
		ok, _, callErr := procWinHttpReadData.Call(
			uintptr(b.req.hReq),
			uintptr(unsafe.Pointer(&readBuf[0])),
			uintptr(toRead),
			0, // lpdwNumberOfBytesRead MUST be NULL under WINHTTP_FLAG_ASYNC
		)
		if ok == 0 {
			return 0, toWinErrno(callErr)
		}
		return ok, 0
	})
	if err != nil {
		return 0, err
	}
	if res.bytesRead == 0 {
		b.eof = true
		return 0, io.EOF
	}

	n := copy(p, readBuf[:res.bytesRead])
	return n, nil
}

func (b *winHTTPResponseBody) Close() error {
	b.closed.Store(true)
	if b.req != nil {
		b.req.startTeardown(true)
	}
	return nil
}
