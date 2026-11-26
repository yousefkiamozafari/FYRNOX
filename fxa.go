package vpnclient

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
)

const (
	fxaAuthServer    = "https://api.accounts.firefox.com/v1"
	fxaOAuthTokenURL = "https://oauth.accounts.firefox.com/v1/token"
	firefoxClientID  = "5882386c6d801776"
	oauthScope       = "profile https://identity.mozilla.com/apps/vpn"
	protocolVersion  = "identity.mozilla.com/picl/v1/"
	pbkdf2Rounds     = 1000
	stretchedPWLen   = 32
	hkdfLen          = 32

	// maxChallengeAttempts bounds how many times an FxA call re-solves the
	// Fastly anti-bot challenge; with rotating-exit proxies a solved cookie
	// can be rejected when the API request leaves through another IP.
	maxChallengeAttempts = 5

	// fxaCallMinBudget is the minimum wall-clock budget an FxA call gets:
	// the caller's per-request timeout is often shorter than a full
	// challenge-solving round trip.
	fxaCallMinBudget = 90 * time.Second

	// verificationMethodEmail2FA tells FxA to deliver the sign-in
	// confirmation as a code emailed to the account instead of the
	// default confirmation link: the link redirects through
	// accounts.firefox.com, which warns about non-Firefox browsers and
	// cannot be followed from this CLI client.
	verificationMethodEmail2FA = "email-2fa"

	// fxaErrnoInvalidParameter is FxA's "Invalid parameter" errno; a
	// deployment that does not accept verificationMethod in the login
	// body answers with it, so we fall back to a plain login.
	fxaErrnoInvalidParameter = 107
	fxaErrnoInvalidToken     = 108
)

type LoginResponse struct {
	SessionToken       string `json:"sessionToken"`
	UID                string `json:"uid"`
	Verified           bool   `json:"verified"`
	AuthAt             int64  `json:"authAt"`
	VerificationMethod string `json:"verificationMethod"`
	VerificationReason string `json:"verificationReason"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// ErrReauthRequired signals that FxA permanently rejected the stored
// credential (refresh token or session token), e.g. after the session was
// revoked server-side. Retrying cannot help; the user must sign in again.
var ErrReauthRequired = errors.New("Firefox Account sign-in required")

// fxaAPIError carries a non-2xx FxA response so callers can branch on the
// API errno (e.g. invalid parameter) instead of parsing error strings.
type fxaAPIError struct {
	Status int
	Errno  int
	Body   string
}

func (e *fxaAPIError) Error() string {
	if e.Errno != 0 {
		return fmt.Sprintf("FxA error (HTTP %d, errno %d): %s", e.Status, e.Errno, e.Body)
	}
	return fmt.Sprintf("FxA request failed (HTTP %d): %s", e.Status, e.Body)
}

// Is allows errors.Is to match ErrReauthRequired for credential rejections. FxA
// answers an invalid or revoked token with HTTP 400 (invalid_grant), 401, or 403;
// HTTP 429 and 5xx stay transient.
func (e *fxaAPIError) Is(target error) bool {
	if target != ErrReauthRequired {
		return false
	}
	switch e.Status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
		return e.Errno != fxaErrnoInvalidParameter
	}
	return false
}

func newFxaAPIError(status int, body []byte) *fxaAPIError {
	var parsed struct {
		Errno int `json:"errno"`
	}
	_ = json.Unmarshal(body, &parsed)
	return &fxaAPIError{Status: status, Errno: parsed.Errno, Body: string(body)}
}

func deriveAuthPW(email, password string) ([]byte, error) {
	salt := []byte(protocolVersion + "quickStretch:" + email)
	quickStretchedPW := pbkdf2.Key([]byte(password), salt, pbkdf2Rounds, stretchedPWLen, sha256.New)

	hkdfSalt := []byte{0x00}
	info := []byte(protocolVersion + "authPW")
	hkdfReader := hkdf.New(sha256.New, quickStretchedPW, hkdfSalt, info)
	authPW := make([]byte, hkdfLen)
	if _, err := io.ReadFull(hkdfReader, authPW); err != nil {
		return nil, err
	}
	return authPW, nil
}

func deriveHawkCredentials(tokenHex, context string) (id string, key []byte, err error) {
	tokenBytes, err := hex.DecodeString(tokenHex)
	if err != nil {
		return "", nil, fmt.Errorf("invalid token hex: %w", err)
	}
	info := []byte(protocolVersion + context)
	hkdfReader := hkdf.New(sha256.New, tokenBytes, nil, info)
	out := make([]byte, 3*32)
	if _, err := io.ReadFull(hkdfReader, out); err != nil {
		return "", nil, err
	}
	return hex.EncodeToString(out[:32]), out[32:64], nil
}

func hawkHeader(method, rawURL, hawkID string, hawkKey []byte, payload string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 6)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}
	nonceStr := hex.EncodeToString(nonce)
	ts := fmt.Sprintf("%d", time.Now().Unix())

	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	var payloadHash string
	if payload != "" {
		h := sha256.New()
		h.Write([]byte("hawk.1.payload\napplication/json\n"))
		h.Write([]byte(payload))
		h.Write([]byte("\n"))
		payloadHash = hex.EncodeToString(h.Sum(nil))
	}

	normalized := strings.Join([]string{
		"hawk.1.header",
		ts,
		nonceStr,
		strings.ToUpper(method),
		u.RequestURI(),
		u.Hostname(),
		port,
		payloadHash,
		"",
		"",
	}, "\n")

	mac := hmac.New(sha256.New, hawkKey)
	mac.Write([]byte(normalized))
	macStr := hex.EncodeToString(mac.Sum(nil))

	header := fmt.Sprintf(`Hawk id="%s", ts="%s", nonce="%s", mac="%s"`, hawkID, ts, nonceStr, macStr)
	if payloadHash != "" {
		header += fmt.Sprintf(`, hash="%s"`, payloadHash)
	}
	return header, nil
}

// fxaDo sends an FxA API request built by newRequest, transparently solving
// the Fastly anti-bot client challenge when the edge answers HTTP 406 (empty
// body, no FxA error payload). The request factory is invoked again after
// each successful solve because request bodies are single-use readers.
func fxaDo(ctx context.Context, newRequest func() (*http.Request, error)) (*http.Response, error) {
	return fxaDoWith406Fallback(ctx, newRequest, nil)
}

// fxaDoWith406Fallback sends an FxA API request with support for a custom 406 handler
// (e.g. alternate OAuth endpoint fallback) before or in lieu of the Fastly anti-bot challenge.
func fxaDoWith406Fallback(ctx context.Context, newRequest func() (*http.Request, error), on406 func(context.Context, int) (*http.Response, error)) (*http.Response, error) {
	ctx, cancel := contextWithMinDeadline(ctx, fxaCallMinBudget)
	defer cancel()

	for solved := 0; solved <= maxChallengeAttempts; solved++ {
		req, err := newRequest()
		if err != nil {
			return nil, err
		}
		resp, err := doControlPlane(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusNotAcceptable {
			return resp, nil
		}
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()

		if on406 != nil {
			if alternate, err := on406(ctx, solved); alternate != nil || err != nil {
				return alternate, err
			}
		}

		// Avoid attempting Fastly HTML challenge parsing on standard JSON API errors.
		if isJSONAPIResponse(resp.Header, bodyBytes) {
			return nil, newFxaAPIError(http.StatusNotAcceptable, bodyBytes)
		}

		if solved == maxChallengeAttempts {
			return nil, fmt.Errorf("Fastly anti-bot challenge still failing after %d attempts (HTTP 406)", maxChallengeAttempts)
		}

		if err := solveFastlyChallenge(ctx); err != nil {
			return nil, fmt.Errorf("Fastly anti-bot challenge: %w", err)
		}
	}
	return nil, fmt.Errorf("Fastly anti-bot challenge still failing after %d attempts (HTTP 406)", maxChallengeAttempts)
}

func isJSONAPIResponse(header http.Header, body []byte) bool {
	ct := header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		return true
	}
	trimmed := strings.TrimSpace(string(body))
	return strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")
}

func fxaLogin(ctx context.Context, email, password string) (*LoginResponse, error) {
	loginResp, err := fxaLoginAttempt(ctx, email, password, verificationMethodEmail2FA)
	if err != nil {
		var apiErr *fxaAPIError
		if errors.As(err, &apiErr) && apiErr.Errno == fxaErrnoInvalidParameter {
			return fxaLoginAttempt(ctx, email, password, "")
		}
		return nil, err
	}
	return loginResp, nil
}

func fxaLoginAttempt(ctx context.Context, email, password, verificationMethod string) (*LoginResponse, error) {
	authPW, err := deriveAuthPW(email, password)
	if err != nil {
		return nil, fmt.Errorf("deriving authPW: %w", err)
	}

	body := map[string]string{
		"email":  email,
		"authPW": hex.EncodeToString(authPW),
	}
	if verificationMethod != "" {
		body["verificationMethod"] = verificationMethod
	}
	bodyJSON, _ := json.Marshal(body)

	loginURL := fxaAuthServer + "/account/login"

	resp, err := fxaDo(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(string(bodyJSON)))
		if err != nil {
			return nil, fmt.Errorf("creating login request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		applyMozillaVPNHeaders(req)
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, newFxaAPIError(resp.StatusCode, data)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(data, &loginResp); err != nil {
		return nil, fmt.Errorf("parsing login response: %w", err)
	}
	return &loginResp, nil
}

// fxaVerifySession submits the email confirmation code for an unverified
// session created by fxaLogin; FxA sends the code to the account email at
// login time.
func fxaVerifySession(ctx context.Context, sessionToken, code string) error {
	hawkID, hawkKey, err := deriveHawkCredentials(sessionToken, "sessionToken")
	if err != nil {
		return fmt.Errorf("deriving hawk credentials: %w", err)
	}

	body := map[string]string{"code": code}
	bodyJSON, _ := json.Marshal(body)
	verifyURL := fxaAuthServer + "/session/verify_code"

	resp, err := fxaDo(ctx, func() (*http.Request, error) {
		authHeader, err := hawkHeader("POST", verifyURL, hawkID, hawkKey, string(bodyJSON))
		if err != nil {
			return nil, fmt.Errorf("generating hawk header: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, strings.NewReader(string(bodyJSON)))
		if err != nil {
			return nil, fmt.Errorf("creating verify code request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		applyMozillaVPNHeaders(req)
		return req, nil
	})
	if err != nil {
		return fmt.Errorf("verify code request: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("verification failed (HTTP %d): %s", resp.StatusCode, truncate(string(data), 256))
	}
	return nil
}

func fxaOAuthToken(ctx context.Context, sessionToken string) (*TokenResponse, error) {
	hawkID, hawkKey, err := deriveHawkCredentials(sessionToken, "sessionToken")
	if err != nil {
		return nil, fmt.Errorf("deriving hawk credentials: %w", err)
	}

	body := map[string]interface{}{
		"client_id":   firefoxClientID,
		"grant_type":  "fxa-credentials",
		"scope":       oauthScope,
		"access_type": "offline",
	}
	bodyJSON, _ := json.Marshal(body)

	tokenURL := fxaAuthServer + "/oauth/token"

	resp, err := fxaDo(ctx, func() (*http.Request, error) {
		authHeader, err := hawkHeader("POST", tokenURL, hawkID, hawkKey, string(bodyJSON))
		if err != nil {
			return nil, fmt.Errorf("generating hawk header: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(string(bodyJSON)))
		if err != nil {
			return nil, fmt.Errorf("creating oauth token request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		applyMozillaVPNHeaders(req)
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("oauth token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, newFxaAPIError(resp.StatusCode, []byte(readErrorBody(resp.Body)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	var tok TokenResponse
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}
	return &tok, nil
}

func requestOAuthRefresh(ctx context.Context, newRequest func(string) (*http.Request, error)) (*http.Response, error) {
	const primary = fxaAuthServer + "/oauth/token"
	return fxaDoWith406Fallback(ctx, func() (*http.Request, error) {
		return newRequest(primary)
	}, func(ctx context.Context, solved int) (*http.Response, error) {
		if solved != 0 {
			return nil, nil
		}
		log.Printf("[INFO] Primary OAuth token refresh challenged (HTTP 406). Attempting alternate Mozilla OAuth endpoint...")
		req, err := newRequest(fxaOAuthTokenURL)
		if err != nil {
			return nil, err
		}
		return refreshAtOAuthEndpoint(req.WithContext(ctx))
	})
}

// refreshAtOAuthEndpoint sends the refresh request to Mozilla's alternate OAuth endpoint (oauth.accounts.firefox.com).
func refreshAtOAuthEndpoint(req *http.Request) (*http.Response, error) {
	resp, err := doControlPlane(req)
	if err != nil {
		log.Printf("[WARN] Alternate Mozilla OAuth endpoint network error: %v", err)
		return nil, nil
	}
	log.Printf("[INFO] Alternate Mozilla OAuth endpoint returned HTTP %d", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		log.Printf("[INFO] Token refreshed successfully via alternate Mozilla OAuth endpoint.")
		return resp, nil
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotAcceptable:
		return nil, nil
	case http.StatusBadRequest:
		apiErr := newFxaAPIError(resp.StatusCode, []byte(readErrorBody(resp.Body)))
		if apiErr.Errno == fxaErrnoInvalidToken {
			apiErr.Body = "invalid refresh token"
			return nil, apiErr
		}
		return nil, apiErr
	}
	return nil, fmt.Errorf("Mozilla's OAuth endpoint returned HTTP %d: %s", resp.StatusCode, readErrorBody(resp.Body))
}

func fxaRefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	body := map[string]interface{}{
		"client_id":     firefoxClientID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"scope":         oauthScope,
	}
	bodyJSON, _ := json.Marshal(body)

	resp, err := requestOAuthRefresh(ctx, func(tokenURL string) (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(bodyJSON))
		if err != nil {
			return nil, fmt.Errorf("creating refresh request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		applyFirefoxHeaders(req)
		return req, nil
	})
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, newFxaAPIError(resp.StatusCode, []byte(readErrorBody(resp.Body)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading refresh response: %w", err)
	}

	var tok TokenResponse
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("parsing refresh response: %w", err)
	}
	// Refresh response may not include a new refresh_token; keep the old one
	if tok.RefreshToken == "" {
		tok.RefreshToken = refreshToken
	}
	return &tok, nil
}

func FxaLogin(ctx context.Context, email, password string) (*LoginResponse, error) {
	return fxaLogin(ctx, email, password)
}

func FxaVerifySession(ctx context.Context, sessionToken, code string) error {
	return fxaVerifySession(ctx, sessionToken, code)
}

func FxaOAuthToken(ctx context.Context, sessionToken string) (*TokenResponse, error) {
	return fxaOAuthToken(ctx, sessionToken)
}

func FxaRefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return fxaRefreshToken(ctx, refreshToken)
}
