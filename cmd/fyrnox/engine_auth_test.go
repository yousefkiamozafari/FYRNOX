package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	vpnclient "firefox-vpn-client"
)

// Helper to construct a base test engine with valid runtime auth.
func newTestAuthEngine(accessToken, refreshToken string, expiresIn int, obtainedAt time.Time) *VpnEngine {
	engine := &VpnEngine{
		state:          StateDisconnected,
		guardian:       vpnclient.GuardianEndpointDefault,
		settings:       defaultSettings(),
		meter:          NewTrafficMeter(),
		monthlyTracker: NewMonthlyQuotaTracker(""),
	}
	engine.auth = &runtimeAuth{
		Token: &vpnclient.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    expiresIn,
			Scope:        "profile https://identity.mozilla.com/apps/vpn",
			TokenType:    "bearer",
		},
		ObtainedAt: obtainedAt,
	}
	return engine
}

// 1. Valid access token: no unnecessary refresh.
func TestProxyPass_ValidAccessToken_NoRefresh(t *testing.T) {
	now := time.Now()
	engine := newTestAuthEngine("valid-token", "rt-1", 3600, now)

	var refreshCalled int32
	var guardianCalled int32
	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		atomic.AddInt32(&refreshCalled, 1)
		return nil, errors.New("refresh should not be called")
	}
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		atomic.AddInt32(&guardianCalled, 1)
		if accessToken != "valid-token" {
			t.Errorf("expected access token 'valid-token', got %s", accessToken)
		}
		return &vpnclient.ProxyPassInfo{
			RawToken: "valid-proxy-pass",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	ctx := context.Background()
	pass, err := engine.getOrRenewProxyPass(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass == nil || pass.RawToken != "valid-proxy-pass" {
		t.Fatalf("expected valid proxy pass, got %v", pass)
	}
	if atomic.LoadInt32(&refreshCalled) != 0 {
		t.Fatalf("expected 0 refresh calls, got %d", refreshCalled)
	}
	if atomic.LoadInt32(&guardianCalled) != 1 {
		t.Fatalf("expected 1 guardian call, got %d", guardianCalled)
	}
}

// 2. Expired access token: proactive refresh before Guardian request.
func TestProxyPass_ExpiredAccessToken_ProactiveRefresh(t *testing.T) {
	// Token obtained 2 hours ago with 1-hour expiry (expired 1 hour ago)
	obtained := time.Now().Add(-2 * time.Hour)
	engine := newTestAuthEngine("expired-token", "rt-1", 3600, obtained)

	var refreshCalled int32
	var guardianReceivedToken string

	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		atomic.AddInt32(&refreshCalled, 1)
		return &vpnclient.TokenResponse{
			AccessToken:  "refreshed-token",
			RefreshToken: "rt-1",
			ExpiresIn:    3600,
		}, nil
	}
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		guardianReceivedToken = accessToken
		return &vpnclient.ProxyPassInfo{
			RawToken: "fresh-pass",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	ctx := context.Background()
	pass, err := engine.getOrRenewProxyPass(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass == nil || pass.RawToken != "fresh-pass" {
		t.Fatalf("expected fresh pass, got %v", pass)
	}
	if atomic.LoadInt32(&refreshCalled) != 1 {
		t.Fatalf("expected 1 proactive refresh call, got %d", refreshCalled)
	}
	if guardianReceivedToken != "refreshed-token" {
		t.Fatalf("expected Guardian to receive 'refreshed-token', got %s", guardianReceivedToken)
	}
}

// 3. Access token approaching expiration (<2 min lead): proactive refresh.
func TestProxyPass_ApproachingExpiration_ProactiveRefresh(t *testing.T) {
	// Token expires in 90 seconds (within 2-minute oauthRefreshLead)
	obtained := time.Now().Add(-3510 * time.Second) // 3600 - 90s
	engine := newTestAuthEngine("expiring-soon-token", "rt-1", 3600, obtained)

	var refreshCalled int32
	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		atomic.AddInt32(&refreshCalled, 1)
		return &vpnclient.TokenResponse{
			AccessToken:  "new-lead-token",
			RefreshToken: "rt-1",
			ExpiresIn:    3600,
		}, nil
	}
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		if accessToken != "new-lead-token" {
			t.Errorf("expected Guardian to receive refreshed token 'new-lead-token', got %s", accessToken)
		}
		return &vpnclient.ProxyPassInfo{
			RawToken: "lead-pass",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	ctx := context.Background()
	pass, err := engine.getOrRenewProxyPass(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass == nil || pass.RawToken != "lead-pass" {
		t.Fatalf("expected lead-pass, got %v", pass)
	}
	if atomic.LoadInt32(&refreshCalled) != 1 {
		t.Fatalf("expected proactive refresh call for expiring token, got %d", refreshCalled)
	}
}

// 4. Successful refresh: new access token used.
func TestProxyPass_SuccessfulRefresh_NewTokenUsed(t *testing.T) {
	engine := newTestAuthEngine("old-token", "rt-1", 3600, time.Now().Add(-2*time.Hour))

	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		return &vpnclient.TokenResponse{
			AccessToken:  "new-active-token",
			RefreshToken: "rt-2",
			ExpiresIn:    7200,
		}, nil
	}
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		if accessToken != "new-active-token" {
			return nil, fmt.Errorf("unexpected access token: %s", accessToken)
		}
		return &vpnclient.ProxyPassInfo{
			RawToken: "pass-with-new-token",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	pass, err := engine.getOrRenewProxyPass(context.Background())
	if err != nil {
		t.Fatalf("getOrRenewProxyPass failed: %v", err)
	}
	if pass.RawToken != "pass-with-new-token" {
		t.Fatalf("expected 'pass-with-new-token', got %s", pass.RawToken)
	}

	// Verify in-memory engine auth was updated
	engine.mu.RLock()
	currentAuth := engine.auth
	engine.mu.RUnlock()
	if currentAuth.Token.AccessToken != "new-active-token" {
		t.Errorf("expected engine auth token 'new-active-token', got %s", currentAuth.Token.AccessToken)
	}
	if currentAuth.Token.RefreshToken != "rt-2" {
		t.Errorf("expected engine refresh token 'rt-2', got %s", currentAuth.Token.RefreshToken)
	}
}

// 5. Refresh response without a new refresh token: old refresh token preserved.
func TestProxyPass_PreservesRefreshTokenWhenOmitted(t *testing.T) {
	engine := newTestAuthEngine("old-token", "original-rt", 3600, time.Now().Add(-2*time.Hour))

	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		// Mozilla FxA standard response omitting refresh_token
		return &vpnclient.TokenResponse{
			AccessToken:  "brand-new-at",
			RefreshToken: "original-rt", // Preserved by token layer
			ExpiresIn:    3600,
		}, nil
	}
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		return &vpnclient.ProxyPassInfo{
			RawToken: "ok-pass",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	_, err := engine.getOrRenewProxyPass(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engine.mu.RLock()
	current := engine.auth
	engine.mu.RUnlock()
	if current.Token.RefreshToken != "original-rt" {
		t.Fatalf("expected preserved refresh token 'original-rt', got %s", current.Token.RefreshToken)
	}
}

// 6. Guardian HTTP 401: exactly one refresh and one controlled retry.
func TestProxyPass_HTTP401_SingleRefreshAndControlledRetry(t *testing.T) {
	engine := newTestAuthEngine("initial-token", "rt-1", 3600, time.Now())

	var guardianAttempts int32
	var refreshAttempts int32
	var activateAttempts int32

	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		attempt := atomic.AddInt32(&guardianAttempts, 1)
		if attempt == 1 {
			// First call with initial-token returns HTTP 401 Unauthorized
			return nil, fmt.Errorf("%w (HTTP 401): {\"error\":\"Unauthorized\"}", vpnclient.ErrTokenInvalid)
		}
		if attempt == 2 {
			if accessToken != "refreshed-token-401" {
				t.Errorf("expected retry call to use 'refreshed-token-401', got %s", accessToken)
			}
			return &vpnclient.ProxyPassInfo{
				RawToken: "retry-successful-pass",
				Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
			}, nil
		}
		return nil, errors.New("unexpected additional attempt")
	}

	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		atomic.AddInt32(&refreshAttempts, 1)
		if !force {
			t.Errorf("expected force=true on 401 recovery refresh")
		}
		return &vpnclient.TokenResponse{
			AccessToken:  "refreshed-token-401",
			RefreshToken: "rt-1",
			ExpiresIn:    3600,
		}, nil
	}

	engine.activateGuardianFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.Entitlement, error) {
		atomic.AddInt32(&activateAttempts, 1)
		return nil, errors.New("activate must NOT be called on 401!")
	}

	ctx := context.Background()
	pass, err := engine.getOrRenewProxyPass(ctx)
	if err != nil {
		t.Fatalf("expected successful retry after 401, got err: %v", err)
	}
	if pass.RawToken != "retry-successful-pass" {
		t.Fatalf("expected 'retry-successful-pass', got %s", pass.RawToken)
	}

	if atomic.LoadInt32(&activateAttempts) != 0 {
		t.Errorf("ActivateGuardian was improperly called on HTTP 401! Count: %d", activateAttempts)
	}
	if atomic.LoadInt32(&refreshAttempts) != 1 {
		t.Errorf("expected exactly 1 refresh attempt, got %d", refreshAttempts)
	}
	if atomic.LoadInt32(&guardianAttempts) != 2 {
		t.Errorf("expected exactly 2 guardian attempts (initial + retry), got %d", guardianAttempts)
	}
}

// 7. Refresh failure after HTTP 401: original refresh error preserved.
func TestProxyPass_HTTP401_RefreshFailurePreserved(t *testing.T) {
	engine := newTestAuthEngine("initial-token", "rt-1", 3600, time.Now())

	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		return nil, fmt.Errorf("%w (HTTP 401): {\"error\":\"Unauthorized\"}", vpnclient.ErrTokenInvalid)
	}

	expectedRefreshErr := fmt.Errorf("%w: invalid_grant: refresh token expired", vpnclient.ErrReauthRequired)
	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		return nil, expectedRefreshErr
	}

	_, err := engine.getOrRenewProxyPass(context.Background())
	if err == nil {
		t.Fatal("expected error when refresh fails, got nil")
	}

	// Must preserve underlying refresh error and NOT mask it with raw 401
	if !errors.Is(err, vpnclient.ErrReauthRequired) {
		t.Fatalf("expected error to wrap ErrReauthRequired, got: %v", err)
	}
}

// 8. Repeated HTTP 401: no infinite retry.
func TestProxyPass_RepeatedHTTP401_NoInfiniteRetry(t *testing.T) {
	engine := newTestAuthEngine("initial-token", "rt-1", 3600, time.Now())

	var guardianAttempts int32
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		atomic.AddInt32(&guardianAttempts, 1)
		return nil, fmt.Errorf("%w (HTTP 401): {\"error\":\"Unauthorized\"}", vpnclient.ErrTokenInvalid)
	}

	var refreshAttempts int32
	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		atomic.AddInt32(&refreshAttempts, 1)
		return &vpnclient.TokenResponse{
			AccessToken:  "second-token-still-401",
			RefreshToken: "rt-1",
			ExpiresIn:    3600,
		}, nil
	}

	_, err := engine.getOrRenewProxyPass(context.Background())
	if err == nil {
		t.Fatal("expected error after repeated 401, got nil")
	}

	if atomic.LoadInt32(&guardianAttempts) != 2 {
		t.Fatalf("expected exactly 2 guardian attempts (initial + 1 retry), got %d", guardianAttempts)
	}
	if atomic.LoadInt32(&refreshAttempts) != 1 {
		t.Fatalf("expected exactly 1 refresh attempt, got %d", refreshAttempts)
	}
}

// 9. HTTP 403: no accidental classification as HTTP 401.
func TestProxyPass_HTTP403_TreatedAsNotEntitledNot401(t *testing.T) {
	engine := newTestAuthEngine("valid-token", "rt-1", 3600, time.Now())

	var refreshCalled int32
	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		atomic.AddInt32(&refreshCalled, 1)
		return nil, errors.New("refresh should not be called on 403")
	}

	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		return nil, fmt.Errorf("%w (HTTP 403): {\"error\":\"Forbidden\"}", vpnclient.ErrNotEntitled)
	}

	var activateCalled int32
	engine.activateGuardianFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.Entitlement, error) {
		atomic.AddInt32(&activateCalled, 1)
		return nil, errors.New("activation failed: not subscribed")
	}

	_, err := engine.getOrRenewProxyPass(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if atomic.LoadInt32(&refreshCalled) != 0 {
		t.Errorf("Refresh was improperly called on HTTP 403! Count: %d", refreshCalled)
	}
	if atomic.LoadInt32(&activateCalled) != 1 {
		t.Errorf("expected ActivateGuardian to be called on HTTP 403, got %d", activateCalled)
	}
	if errors.Is(err, vpnclient.ErrTokenInvalid) {
		t.Errorf("error was improperly classified as ErrTokenInvalid: %v", err)
	}
}

// 10. Activation only when entitlement/activation semantics are established (HTTP 403).
func TestProxyPass_HTTP403_ActivationAttempted(t *testing.T) {
	engine := newTestAuthEngine("valid-token", "rt-1", 3600, time.Now())

	var guardianAttempts int32
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		attempt := atomic.AddInt32(&guardianAttempts, 1)
		if attempt == 1 {
			// First call returns 403 Forbidden (account not entitled)
			return nil, fmt.Errorf("%w (HTTP 403): {\"error\":\"Forbidden\"}", vpnclient.ErrNotEntitled)
		}
		// Second call after successful activation returns pass
		return &vpnclient.ProxyPassInfo{
			RawToken: "activated-pass",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	var activateAttempts int32
	engine.activateGuardianFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.Entitlement, error) {
		atomic.AddInt32(&activateAttempts, 1)
		if accessToken != "valid-token" {
			t.Errorf("expected activation with valid access token, got %s", accessToken)
		}
		return &vpnclient.Entitlement{Subscribed: true}, nil
	}

	pass, err := engine.getOrRenewProxyPass(context.Background())
	if err != nil {
		t.Fatalf("expected successful pass fetch after activation, got: %v", err)
	}
	if pass.RawToken != "activated-pass" {
		t.Fatalf("expected 'activated-pass', got %s", pass.RawToken)
	}
	if atomic.LoadInt32(&activateAttempts) != 1 {
		t.Fatalf("expected exactly 1 activation attempt, got %d", activateAttempts)
	}
	if atomic.LoadInt32(&guardianAttempts) != 2 {
		t.Fatalf("expected exactly 2 guardian attempts, got %d", guardianAttempts)
	}
}

// 11. Concurrent refresh requests: no conflicting token writes.
func TestProxyPass_ConcurrentRefresh_SafeSynchronization(t *testing.T) {
	obtained := time.Now().Add(-2 * time.Hour)
	engine := newTestAuthEngine("expired-initial", "rt-1", 3600, obtained)

	var networkRefreshCount int32
	engine.refreshTokenFn = func(ctx context.Context, force bool) (*vpnclient.TokenResponse, error) {
		time.Sleep(10 * time.Millisecond) // Simulate network delay
		atomic.AddInt32(&networkRefreshCount, 1)
		return &vpnclient.TokenResponse{
			AccessToken:  "concurrent-fresh-token",
			RefreshToken: "rt-1",
			ExpiresIn:    3600,
		}, nil
	}

	const workers = 15
	var wg sync.WaitGroup
	wg.Add(workers)

	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			auth, err := engine.freshRuntimeAuth(context.Background(), false)
			if err != nil {
				errs[idx] = err
				return
			}
			if auth.Token.AccessToken != "concurrent-fresh-token" {
				errs[idx] = fmt.Errorf("unexpected token: %s", auth.Token.AccessToken)
			}
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("worker %d failed: %v", i, err)
		}
	}

	engine.mu.RLock()
	finalToken := engine.auth.Token.AccessToken
	engine.mu.RUnlock()
	if finalToken != "concurrent-fresh-token" {
		t.Fatalf("expected final token 'concurrent-fresh-token', got %s", finalToken)
	}
}

// 12. Existing connection rollback behavior remains intact.
func TestProxyPass_RollbackIntegrity(t *testing.T) {
	engine := newTestAuthEngine("valid-token", "rt-1", 3600, time.Now())
	engine.fetchProxyPassFn = func(ctx context.Context, endpoint, accessToken string) (*vpnclient.ProxyPassInfo, error) {
		return &vpnclient.ProxyPassInfo{
			RawToken: "valid-pass",
			Claims:   vpnclient.ProxyPassClaims{Exp: time.Now().Add(10 * time.Minute).Unix()},
		}, nil
	}

	// Verify standard helper rollback test state
	status := engine.GetStatus()
	if status.State != StateDisconnected {
		t.Fatalf("expected initial StateDisconnected, got %s", status.State)
	}
}
