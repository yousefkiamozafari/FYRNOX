package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	vpnclient "firefox-vpn-client"
)

const (
	primaryOAuthURL   = "https://api.accounts.firefox.com/v1/oauth/token"
	alternateOAuthURL = "https://oauth.accounts.firefox.com/v1/token"
	firefoxClientID   = "5882386c6d801776"
	oauthScope        = "profile https://identity.mozilla.com/apps/vpn"
	chromeUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

type MatrixCellResult struct {
	Name            string
	TLSProfile      string
	HeadersProfile  string
	ConnectTime     time.Duration
	HandshakeTime   time.Duration
	InitialHTTP     int
	HasChallengeTag bool
	IsJSONResponse  bool
	Error           string
}

func sha256Prefix(s string) string {
	if s == "" {
		return "none"
	}
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:8]
}

func runMatrixCell(ctx context.Context, name, targetURL string, standardTLS bool, firefoxHeaders bool, refreshToken string) MatrixCellResult {
	res := MatrixCellResult{
		Name: name,
	}
	if standardTLS {
		res.TLSProfile = "crypto/tls (Go standard)"
	} else {
		res.TLSProfile = "uTLS (Firefox Desktop)"
	}
	if firefoxHeaders {
		res.HeadersProfile = "Firefox (" + vpnclient.FirefoxUserAgent() + ")"
	} else {
		res.HeadersProfile = "Legacy Chrome 126"
	}

	tr, err := vpnclient.NewControlPlaneTransport(vpnclient.ControlPlaneTransportOptions{
		ForceStandardTLS: standardTLS,
		HandshakeTimeout: 15 * time.Second,
	})
	if err != nil {
		res.Error = fmt.Sprintf("creating transport: %v", err)
		return res
	}
	defer func() {
		type idleCloser interface{ CloseIdleConnections() }
		if ic, ok := tr.(idleCloser); ok {
			ic.CloseIdleConnections()
		}
	}()

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}

	body := map[string]interface{}{
		"client_id":     firefoxClientID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"scope":         oauthScope,
	}
	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(bodyJSON))
	if err != nil {
		res.Error = fmt.Sprintf("creating request: %v", err)
		return res
	}

	req.Header.Set("Content-Type", "application/json")
	if firefoxHeaders {
		req.Header.Set("User-Agent", vpnclient.FirefoxUserAgent())
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Encoding", "identity")
	} else {
		req.Header.Set("User-Agent", chromeUserAgent)
		req.Header.Set("Accept", "application/json")
	}

	start := time.Now()
	resp, err := client.Do(req)
	res.HandshakeTime = time.Since(start).Round(time.Millisecond)

	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.InitialHTTP = resp.StatusCode
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	bodyStr := string(bodyBytes)

	res.HasChallengeTag = strings.Contains(bodyStr, "_fs_ch_script") || strings.Contains(bodyStr, "/_fs-ch-")
	res.IsJSONResponse = strings.Contains(resp.Header.Get("Content-Type"), "application/json") ||
		(strings.HasPrefix(strings.TrimSpace(bodyStr), "{") && strings.HasSuffix(strings.TrimSpace(bodyStr), "}"))

	return res
}

func main() {
	fmt.Println("==========================================================================")
	fmt.Println("   FYRNOX — Firefox-Compatible TLS Control Plane Diagnostic Suite        ")
	fmt.Println("==========================================================================")
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339))

	// Step 1: Inspect Stored Credentials (Read-Only)
	fmt.Println("[Step 1] Inspecting Stored Credentials (Read-Only)...")
	tokens, err := vpnclient.LoadTokens()
	if err != nil {
		fmt.Printf("[-] Failed to load stored tokens: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Access Token Present:       %t\n", tokens.AccessToken != "")
	fmt.Printf("[+] Access Token SHA-256 (8c):  %s\n", sha256Prefix(tokens.AccessToken))
	fmt.Printf("[+] Access Token Valid:         %t\n", tokens.AccessTokenValid())
	fmt.Printf("[+] Token Obtained At:          %s\n", tokens.ObtainedAt.Format(time.RFC3339))
	fmt.Printf("[+] Token Age:                  %v\n", time.Since(tokens.ObtainedAt).Round(time.Minute))
	fmt.Printf("[+] Refresh Token Present:      %t\n", tokens.RefreshToken != "")
	fmt.Printf("[+] Refresh Token SHA-256(8c):  %s\n", sha256Prefix(tokens.RefreshToken))
	fmt.Printf("[+] Configured Scope:           %s\n", tokens.Scope)
	fmt.Printf("[+] Runtime Reported Firefox:   %s\n\n", vpnclient.FirefoxUserAgent())

	// Step 2: Controlled 4-Cell A/B Matrix Diagnostic
	fmt.Println("[Step 2] Executing Controlled 4-Cell A/B Diagnostic Matrix...")
	fmt.Println("         Target Endpoint: " + primaryOAuthURL)
	fmt.Println("--------------------------------------------------------------------------")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cells := []struct {
		name           string
		standardTLS    bool
		firefoxHeaders bool
	}{
		{"Cell 1: crypto/tls + Chrome Headers", true, false},
		{"Cell 2: crypto/tls + Firefox Headers", true, true},
		{"Cell 3: uTLS       + Chrome Headers", false, false},
		{"Cell 4: uTLS       + Firefox Headers", false, true},
	}

	for idx, c := range cells {
		fmt.Printf("\n--> Running [%d/4] %s...\n", idx+1, c.name)
		result := runMatrixCell(ctx, c.name, primaryOAuthURL, c.standardTLS, c.firefoxHeaders, tokens.RefreshToken)

		fmt.Printf("    * TLS Transport:      %s\n", result.TLSProfile)
		fmt.Printf("    * Headers Sent:       %s\n", result.HeadersProfile)
		fmt.Printf("    * Roundtrip Elapsed:  %v\n", result.HandshakeTime)
		if result.Error != "" {
			fmt.Printf("    * Error Encountered:  %s\n", result.Error)
		} else {
			fmt.Printf("    * Initial HTTP Code:  HTTP %d\n", result.InitialHTTP)
			fmt.Printf("    * Fastly Challenge:   %t\n", result.HasChallengeTag)
			fmt.Printf("    * JSON API Response:  %t\n", result.IsJSONResponse)
		}
	}

	fmt.Println("\n--------------------------------------------------------------------------")
	fmt.Println("[Step 3] Testing Production Unified Transport & Token Renewal...")
	fmt.Println("--------------------------------------------------------------------------")

	refreshCtx, refreshCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer refreshCancel()

	startRefresh := time.Now()
	newTok, err := vpnclient.ForceRefreshAndSaveTokens(refreshCtx)
	refreshDuration := time.Since(startRefresh).Round(time.Millisecond)

	if err != nil {
		fmt.Printf("[-] Production Token Refresh Failed (%v): %v\n", refreshDuration, err)
		fmt.Println("    (Check if Mozilla account credentials or refresh token have been permanently revoked)")
	} else {
		fmt.Printf("[+] Production Token Refresh Succeeded in %v!\n", refreshDuration)
		fmt.Printf("[+] New Access Token SHA-256 (8c):   %s\n", sha256Prefix(newTok.AccessToken))
		fmt.Printf("[+] New Refresh Token SHA-256 (8c):  %s\n", sha256Prefix(newTok.RefreshToken))
		fmt.Printf("[+] Token Lifetime (expires_in):     %d seconds\n\n", newTok.ExpiresIn)

		// Step 4: Guardian Proxy Pass Verification
		fmt.Println("[Step 4] Testing Live Guardian Proxy Pass Acquisition...")
		guardianCtx, guardianCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer guardianCancel()

		startGuardian := time.Now()
		pass, err := vpnclient.FetchProxyPass(guardianCtx, vpnclient.GuardianEndpointDefault, newTok.AccessToken)
		guardianDuration := time.Since(startGuardian).Round(time.Millisecond)

		if err != nil {
			fmt.Printf("[-] Guardian Proxy Pass Acquisition Failed (%v): %v\n", guardianDuration, err)
		} else {
			fmt.Printf("[+] Guardian Proxy Pass Succeeded in %v!\n", guardianDuration)
			fmt.Printf("[+] Pass Issued At:        %s\n", time.Unix(pass.Claims.Iat, 0).Format(time.RFC3339))
			fmt.Printf("[+] Pass Expires At:       %s\n", pass.ExpiresAt().Format(time.RFC3339))
			fmt.Printf("[+] Quota Remaining:       %s / %s\n", pass.QuotaLeft, pass.QuotaMax)
		}
	}

	fmt.Println("\n==========================================================================")
	fmt.Println("   Diagnostic Run Completed                                               ")
	fmt.Println("==========================================================================")
}
