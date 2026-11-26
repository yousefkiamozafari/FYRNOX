package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	vpnclient "firefox-vpn-client"
)

func sha256Prefix(s string) string {
	if s == "" {
		return "none"
	}
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:8]
}

func main() {
	fmt.Println("==================================================================")
	fmt.Println("  FYRNOX - Post-v1.0.6 OAuth Refresh & Guardian Diagnostic Tool   ")
	fmt.Println("==================================================================")
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339))

	// Step 1: Inspect Stored Credentials in Read-Only Mode
	fmt.Println("[Step 1] Inspecting Stored Credentials...")
	tokens, err := vpnclient.LoadTokens()
	if err != nil {
		fmt.Printf("[-] Failed to load stored tokens: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Access Token Present:      %t\n", tokens.AccessToken != "")
	fmt.Printf("[+] Access Token SHA-256 (8c): %s\n", sha256Prefix(tokens.AccessToken))
	fmt.Printf("[+] Access Token Valid:        %t\n", tokens.AccessTokenValid())
	fmt.Printf("[+] Token Obtained At:         %s\n", tokens.ObtainedAt.Format(time.RFC3339))
	fmt.Printf("[+] Token Age:                 %v\n", time.Since(tokens.ObtainedAt).Round(time.Minute))
	fmt.Printf("[+] Refresh Token Present:     %t\n", tokens.RefreshToken != "")
	fmt.Printf("[+] Refresh Token SHA-256(8c): %s\n", sha256Prefix(tokens.RefreshToken))
	fmt.Printf("[+] Configured Scope:          %s\n\n", tokens.Scope)

	// Step 2: Attempt Real OAuth Refresh (Primary with Alternate Endpoint Fallback)
	fmt.Println("[Step 2] Testing Live OAuth Token Refresh via Updated Protocol...")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	startRefresh := time.Now()
	newTok, err := vpnclient.ForceRefreshAndSaveTokens(ctx)
	refreshDuration := time.Since(startRefresh).Round(time.Millisecond)

	if err != nil {
		fmt.Printf("[-] Live OAuth Refresh Failed (%v): %v\n", refreshDuration, err)
		fmt.Println("    (Check if Mozilla account credentials or refresh token have been permanently revoked)")
		os.Exit(2)
	}

	fmt.Printf("[+] Live OAuth Refresh Succeeded in %v!\n", refreshDuration)
	fmt.Printf("[+] New Access Token SHA-256 (8c):  %s\n", sha256Prefix(newTok.AccessToken))
	fmt.Printf("[+] New Refresh Token SHA-256 (8c): %s\n", sha256Prefix(newTok.RefreshToken))
	fmt.Printf("[+] Token Expires In:               %d seconds\n\n", newTok.ExpiresIn)

	// Step 3: Test Guardian Proxy Pass Acquisition with Refreshed Access Token
	fmt.Println("[Step 3] Testing Live Guardian Proxy Pass Acquisition...")
	guardianCtx, guardianCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer guardianCancel()

	startGuardian := time.Now()
	pass, err := vpnclient.FetchProxyPass(guardianCtx, vpnclient.GuardianEndpointDefault, newTok.AccessToken)
	guardianDuration := time.Since(startGuardian).Round(time.Millisecond)

	if err != nil {
		fmt.Printf("[-] Guardian Proxy Pass Acquisition Failed (%v): %v\n", guardianDuration, err)
		os.Exit(3)
	}

	fmt.Printf("[+] Guardian Proxy Pass Succeeded in %v!\n", guardianDuration)
	fmt.Printf("[+] Pass Issued At:        %s\n", time.Unix(pass.Claims.Iat, 0).Format(time.RFC3339))
	fmt.Printf("[+] Pass Expires At:       %s\n", pass.ExpiresAt().Format(time.RFC3339))
	fmt.Printf("[+] Quota Remaining:       %s / %s\n\n", pass.QuotaLeft, pass.QuotaMax)

	fmt.Println("==================================================================")
	fmt.Println("  [SUCCESS] All Remediation Verification Checks Passed!           ")
	fmt.Println("==================================================================")
}
