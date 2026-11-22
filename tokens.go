package vpnclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const tokenFileName = ".firefox-vpn-tokens.json"

var (
	tokenMu         sync.Mutex
	refreshMu       sync.Mutex
	customTokenPath string
)

type SavedTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	ObtainedAt   time.Time `json:"obtained_at"`
	ExpiresIn    int       `json:"expires_in"`
	UserEmail    string    `json:"user_email,omitempty"`
}

func (s *SavedTokens) AccessTokenValid() bool {
	if s.AccessToken == "" {
		return false
	}
	expiry := s.ObtainedAt.Add(time.Duration(s.ExpiresIn) * time.Second)
	return time.Now().Before(expiry.Add(-60 * time.Second))
}

func tokenFilePath() string {
	if customTokenPath != "" {
		return customTokenPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return tokenFileName
	}
	return filepath.Join(home, tokenFileName)
}

func loadTokens() (*SavedTokens, error) {
	data, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return nil, err
	}
	var tokens SavedTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}
	return &tokens, nil
}

func saveTokensLocked(tok *TokenResponse, email string) error {
	if tok == nil || tok.AccessToken == "" {
		return errors.New("cannot save empty or nil token response")
	}

	var existingEmail string
	var existingRefreshToken string
	var existingScope string
	if existing, err := loadTokens(); err == nil && existing != nil {
		existingEmail = existing.UserEmail
		existingRefreshToken = existing.RefreshToken
		existingScope = existing.Scope
	}

	refreshToken := tok.RefreshToken
	if refreshToken == "" {
		if existingRefreshToken != "" {
			refreshToken = existingRefreshToken
		} else {
			return errors.New("cannot save tokens without a valid refresh token")
		}
	}

	userEmail := email
	if userEmail == "" && existingEmail != "" {
		userEmail = existingEmail
	}

	scope := tok.Scope
	if scope == "" && existingScope != "" {
		scope = existingScope
	}

	expiresIn := tok.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 86400
	}

	saved := SavedTokens{
		AccessToken:  tok.AccessToken,
		RefreshToken: refreshToken,
		Scope:        scope,
		ObtainedAt:   time.Now(),
		ExpiresIn:    expiresIn,
		UserEmail:    userEmail,
	}
	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}

	tok.RefreshToken = refreshToken

	path := tokenFilePath()
	if err := writePrivateFile(path, data); err != nil {
		return fmt.Errorf("saving tokens to %s: %w", path, err)
	}
	return nil
}

func writePrivateFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return err
		}
		if retryErr := os.Rename(tmpPath, path); retryErr != nil {
			return retryErr
		}
	}
	return os.Chmod(path, 0600)
}

func deleteTokens() {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	os.Remove(tokenFilePath())
}

func TokenFilePath() string {
	return tokenFilePath()
}

func LoadTokens() (*SavedTokens, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	return loadTokens()
}

func SaveTokens(tok *TokenResponse) error {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	return saveTokensLocked(tok, "")
}

func SaveTokensWithEmail(tok *TokenResponse, email string) error {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	return saveTokensLocked(tok, email)
}

// RefreshAndSaveTokens performs the complete thread-safe OAuth token refresh lifecycle:
// 1. Reads the current valid refresh token under tokenMu lock.
// 2. Returns existing tokens if another concurrent routine already completed a refresh.
// 3. Executes the network OAuth token refresh without holding locks.
// 4. Validates the resulting token response.
// 5. Atomically persists the new access token while preserving the refresh token under lock.
func RefreshAndSaveTokens(ctx context.Context) (*TokenResponse, error) {
	return refreshAndSaveTokens(ctx, false)
}

// ForceRefreshAndSaveTokens forces a network OAuth refresh even if the saved
// access token has not yet reached its expiration timestamp (e.g. after Guardian
// rejected the Bearer token with HTTP 401).
func ForceRefreshAndSaveTokens(ctx context.Context) (*TokenResponse, error) {
	return refreshAndSaveTokens(ctx, true)
}

func refreshAndSaveTokens(ctx context.Context, force bool) (*TokenResponse, error) {
	refreshMu.Lock()
	defer refreshMu.Unlock()

	var currentRefreshToken string
	tokenMu.Lock()
	existing, err := loadTokens()
	if err == nil && existing != nil {
		if !force && existing.AccessTokenValid() {
			tokenMu.Unlock()
			return &TokenResponse{
				AccessToken:  existing.AccessToken,
				RefreshToken: existing.RefreshToken,
				ExpiresIn:    existing.ExpiresIn,
				Scope:        existing.Scope,
				TokenType:    "bearer",
			}, nil
		}
		currentRefreshToken = existing.RefreshToken
	}
	tokenMu.Unlock()

	if currentRefreshToken == "" {
		return nil, fmt.Errorf("%w: no saved refresh token found", ErrReauthRequired)
	}

	newTok, err := fxaRefreshToken(ctx, currentRefreshToken)
	if err != nil {
		if errors.Is(err, ErrReauthRequired) {
			return nil, fmt.Errorf("%w: %v", ErrReauthRequired, err)
		}
		return nil, fmt.Errorf("oauth token refresh: %w", err)
	}
	if newTok == nil || newTok.AccessToken == "" {
		return nil, errors.New("received empty or nil token from oauth refresh")
	}

	if err := SaveTokens(newTok); err != nil {
		return nil, fmt.Errorf("persisting refreshed token: %w", err)
	}

	return newTok, nil
}

func DeleteTokens() {
	deleteTokens()
}
