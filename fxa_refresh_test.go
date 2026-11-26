package vpnclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type mockTransport func(*http.Request) (*http.Response, error)

func (m mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m(req)
}

func jsonResponse(status int, body interface{}) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(b)),
	}
}

// 1. Primary endpoint returns HTTP 200, valid token parsed.
func TestOAuthRefresh_PrimarySuccess_200(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			return jsonResponse(http.StatusOK, map[string]interface{}{
				"access_token":  "primary-at",
				"refresh_token": "primary-rt",
				"expires_in":    3600,
				"token_type":    "bearer",
			}), nil
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	tok, err := fxaRefreshToken(context.Background(), "rt-initial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "primary-at" {
		t.Fatalf("expected access token primary-at, got %s", tok.AccessToken)
	}
	if tok.RefreshToken != "primary-rt" {
		t.Fatalf("expected refresh token primary-rt, got %s", tok.RefreshToken)
	}
}

// 2. Primary returns 406, alternate returns 200, token updated.
func TestOAuthRefresh_Primary406_AlternateSuccess_200(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	primaryAttempts := 0
	alternateAttempts := 0

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			primaryAttempts++
			return jsonResponse(http.StatusNotAcceptable, map[string]interface{}{
				"error": "Not Acceptable",
			}), nil
		}
		if req.URL.String() == fxaOAuthTokenURL {
			alternateAttempts++
			return jsonResponse(http.StatusOK, map[string]interface{}{
				"access_token":  "alternate-at",
				"refresh_token": "alternate-rt",
				"expires_in":    3600,
				"token_type":    "bearer",
			}), nil
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	tok, err := fxaRefreshToken(context.Background(), "rt-initial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primaryAttempts != 1 {
		t.Fatalf("expected 1 primary attempt, got %d", primaryAttempts)
	}
	if alternateAttempts != 1 {
		t.Fatalf("expected 1 alternate attempt, got %d", alternateAttempts)
	}
	if tok.AccessToken != "alternate-at" {
		t.Fatalf("expected access token alternate-at, got %s", tok.AccessToken)
	}
}

// 3. Primary returns 400 invalid_grant, immediately returns ErrReauthRequired without alternate retry.
func TestOAuthRefresh_Primary400_NoAlternateRetry(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	alternateCalled := false

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			return jsonResponse(http.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"errno":   fxaErrnoInvalidToken,
				"message": "invalid_grant",
			}), nil
		}
		if req.URL.String() == fxaOAuthTokenURL {
			alternateCalled = true
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	_, err := fxaRefreshToken(context.Background(), "rt-revoked")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrReauthRequired) {
		t.Fatalf("expected ErrReauthRequired, got %v", err)
	}
	if alternateCalled {
		t.Fatalf("alternate OAuth endpoint should NOT be called on 400 invalid_grant")
	}
}

// 4. Alternate returns 400 invalid_grant, returns ErrReauthRequired.
func TestOAuthRefresh_Alternate400_ReauthRequired(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			return jsonResponse(http.StatusNotAcceptable, map[string]interface{}{"error": "Fastly challenge"}), nil
		}
		if req.URL.String() == fxaOAuthTokenURL {
			return jsonResponse(http.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"errno":   fxaErrnoInvalidToken,
				"message": "invalid refresh token",
			}), nil
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	_, err := fxaRefreshToken(context.Background(), "rt-expired")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrReauthRequired) {
		t.Fatalf("expected ErrReauthRequired, got %v", err)
	}
}

// 5. Both return 406, stops cleanly with meaningful error.
func TestOAuthRefresh_Both406_BoundedFailure(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	calls := 0
	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(http.StatusNotAcceptable, map[string]interface{}{
			"error": "both blocked",
		}), nil
	}))

	_, err := fxaRefreshToken(context.Background(), "rt-test")
	if err == nil {
		t.Fatalf("expected error when both endpoints return 406, got nil")
	}
	if !strings.Contains(err.Error(), "406") {
		t.Fatalf("expected error mentioning 406, got %v", err)
	}
}

// 6. Verified payload stream is intact across retries.
func TestOAuthRefresh_BodyRecreated(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	var primaryBody, alternateBody []byte

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(req.Body)
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			primaryBody = b
			return jsonResponse(http.StatusNotAcceptable, map[string]interface{}{"error": "blocked"}), nil
		}
		if req.URL.String() == fxaOAuthTokenURL {
			alternateBody = b
			return jsonResponse(http.StatusOK, map[string]interface{}{
				"access_token": "recreated-at",
				"expires_in":   3600,
			}), nil
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	tok, err := fxaRefreshToken(context.Background(), "rt-payload-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "recreated-at" {
		t.Fatalf("unexpected token: %s", tok.AccessToken)
	}

	if len(primaryBody) == 0 {
		t.Fatalf("primary body was empty")
	}
	if len(alternateBody) == 0 {
		t.Fatalf("alternate body was empty (stream not recreated)")
	}
	if !strings.Contains(string(primaryBody), "rt-payload-test") {
		t.Fatalf("primary body did not contain refresh token")
	}
	if !strings.Contains(string(alternateBody), "rt-payload-test") {
		t.Fatalf("alternate body did not contain refresh token")
	}
}

// 7. Verified User-Agent: Firefox/..., Accept: */*, Accept-Encoding: identity.
func TestOAuthRefresh_FirefoxHeadersApplied(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	var capturedReq *http.Request

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		capturedReq = req.Clone(req.Context())
		return jsonResponse(http.StatusOK, map[string]interface{}{
			"access_token": "at-headers",
			"expires_in":   3600,
		}), nil
	}))

	_, err := fxaRefreshToken(context.Background(), "rt-headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ua := capturedReq.Header.Get("User-Agent")
	if !strings.HasPrefix(ua, "Mozilla/5.0") || !strings.Contains(ua, "Firefox/") {
		t.Fatalf("expected Firefox User-Agent, got %q", ua)
	}
	if got := capturedReq.Header.Get("Accept"); got != "*/*" {
		t.Fatalf("expected Accept: */*, got %q", got)
	}
	if got := capturedReq.Header.Get("Accept-Encoding"); got != "identity" {
		t.Fatalf("expected Accept-Encoding: identity, got %q", got)
	}
	if got := capturedReq.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type: application/json, got %q", got)
	}
}

// 8. Verified Guardian headers: Cache-Control: max-age=0, Accept-Encoding: identity, Bearer token.
func TestGuardianHeaders_CacheControlAndIdentity(t *testing.T) {
	req, err := http.NewRequest("GET", "https://guardian.mozilla.org/api/v1/fpn/token", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	applyGuardianHeaders(req, "test-access-token-xyz")

	if got := req.Header.Get("Authorization"); got != "Bearer test-access-token-xyz" {
		t.Fatalf("expected Authorization: Bearer ..., got %q", got)
	}
	if got := req.Header.Get("Cache-Control"); got != "max-age=0" {
		t.Fatalf("expected Cache-Control: max-age=0, got %q", got)
	}
	if got := req.Header.Get("Accept-Encoding"); got != "identity" {
		t.Fatalf("expected Accept-Encoding: identity, got %q", got)
	}
	if got := req.Header.Get("Accept"); got != "*/*" {
		t.Fatalf("expected Accept: */*, got %q", got)
	}
	ua := req.Header.Get("User-Agent")
	if !strings.Contains(ua, "Firefox/") {
		t.Fatalf("expected Firefox User-Agent in Guardian request, got %q", ua)
	}
}

// 9. Omitted refresh_token in response preserves existing token.
func TestOAuthRefresh_PreservesRefreshToken(t *testing.T) {
	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, map[string]interface{}{
			"access_token": "new-at-only",
			"expires_in":   7200,
			"token_type":   "bearer",
		}), nil
	}))

	const existingRT = "rt-original-keep-me"
	tok, err := fxaRefreshToken(context.Background(), existingRT)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "new-at-only" {
		t.Fatalf("expected new-at-only, got %s", tok.AccessToken)
	}
	if tok.RefreshToken != existingRT {
		t.Fatalf("expected preserved refresh token %s, got %s", existingRT, tok.RefreshToken)
	}
}

// 10. Verified atomic save via tokens.go when RefreshAndSaveTokens runs.
func TestOAuthRefresh_PersistsNewAccessToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	// Initial expired token
	initialTok := &TokenResponse{
		AccessToken:  "initial-expired-at",
		RefreshToken: "rt-persist-test",
		ExpiresIn:    3600,
		Scope:        "test-scope",
	}
	if err := SaveTokens(initialTok); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			return jsonResponse(http.StatusNotAcceptable, map[string]interface{}{"error": "406"}), nil
		}
		if req.URL.String() == fxaOAuthTokenURL {
			return jsonResponse(http.StatusOK, map[string]interface{}{
				"access_token": "persisted-fresh-at",
				"expires_in":   3600,
				"token_type":   "bearer",
			}), nil
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	newTok, err := ForceRefreshAndSaveTokens(context.Background())
	if err != nil {
		t.Fatalf("ForceRefreshAndSaveTokens failed: %v", err)
	}
	if newTok.AccessToken != "persisted-fresh-at" {
		t.Fatalf("expected fresh token, got %s", newTok.AccessToken)
	}

	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}
	if loaded.AccessToken != "persisted-fresh-at" {
		t.Fatalf("expected persisted-fresh-at on disk, got %s", loaded.AccessToken)
	}
	if loaded.RefreshToken != "rt-persist-test" {
		t.Fatalf("expected preserved refresh token on disk, got %s", loaded.RefreshToken)
	}
}

// 11. Expired token never passed to Guardian on refresh error.
func TestOAuthRefresh_FailedRefresh_NoExpiredTokenToGuardian(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	// Save expired token
	if err := SaveTokens(&TokenResponse{
		AccessToken:  "expired-token-do-not-use",
		RefreshToken: "rt-failed-refresh",
		ExpiresIn:    1,
	}); err != nil {
		t.Fatalf("save token failed: %v", err)
	}

	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	guardianCalled := false
	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.String(), "guardian") {
			guardianCalled = true
			return jsonResponse(http.StatusOK, nil), nil
		}
		// Refresh returns 400 invalid grant
		return jsonResponse(http.StatusBadRequest, map[string]interface{}{
			"code":  400,
			"errno": fxaErrnoInvalidToken,
		}), nil
	}))

	_, err := ForceRefreshAndSaveTokens(context.Background())
	if err == nil {
		t.Fatalf("expected error from failed refresh, got nil")
	}
	if guardianCalled {
		t.Fatalf("guardian was unexpectedly called during failed refresh")
	}
}

// 12. Verified sensitive tokens do not appear in logger output.
func TestOAuthRefresh_NoSecretsInLogs(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)

	origTransport := controlPlaneHTTPClient.Transport
	defer func() { SetControlPlaneTransport(origTransport) }()

	const sensitiveRefreshToken = "SUPER-SECRET-REFRESH-TOKEN-9988776655"
	const sensitiveAccessToken = "SUPER-SECRET-ACCESS-TOKEN-1122334455"

	SetControlPlaneTransport(mockTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == fxaAuthServer+"/oauth/token" {
			return jsonResponse(http.StatusNotAcceptable, map[string]interface{}{"error": "blocked"}), nil
		}
		if req.URL.String() == fxaOAuthTokenURL {
			return jsonResponse(http.StatusOK, map[string]interface{}{
				"access_token": sensitiveAccessToken,
				"expires_in":   3600,
			}), nil
		}
		return jsonResponse(http.StatusNotFound, nil), nil
	}))

	_, err := fxaRefreshToken(context.Background(), sensitiveRefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logOutput := logBuf.String()
	if strings.Contains(logOutput, sensitiveRefreshToken) {
		t.Fatalf("security violation: sensitive refresh token found in log output: %s", logOutput)
	}
	if strings.Contains(logOutput, sensitiveAccessToken) {
		t.Fatalf("security violation: sensitive access token found in log output: %s", logOutput)
	}
}
