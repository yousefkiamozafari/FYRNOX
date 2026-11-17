package vpnclient

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// 1. Verify that omitting refresh token preserves the existing valid token.
func TestSaveTokensPreservesRefreshTokenWhenOmitted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	// Initial save with complete credentials
	tok1 := &TokenResponse{AccessToken: "at1", RefreshToken: "rt1", Scope: "test-scope", ExpiresIn: 3600}
	if err := SaveTokensWithEmail(tok1, "user@example.com"); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Subsequent save with omitted refresh token (Mozilla default response)
	tok2 := &TokenResponse{AccessToken: "at2", ExpiresIn: 3600}
	if err := SaveTokens(tok2); err != nil {
		t.Fatalf("save with omitted rt failed: %v", err)
	}

	// Verify on-disk persistence
	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("loading tokens failed: %v", err)
	}
	if loaded.RefreshToken != "rt1" {
		t.Errorf("expected preserved rt1, got %s", loaded.RefreshToken)
	}
	if loaded.AccessToken != "at2" {
		t.Errorf("expected updated at2, got %s", loaded.AccessToken)
	}
	if loaded.UserEmail != "user@example.com" {
		t.Errorf("expected preserved email user@example.com, got %s", loaded.UserEmail)
	}
	if loaded.Scope != "test-scope" {
		t.Errorf("expected preserved scope test-scope, got %s", loaded.Scope)
	}

	// Verify in-memory caller's TokenResponse was updated with preserved refresh token
	if tok2.RefreshToken != "rt1" {
		t.Errorf("expected caller tok2.RefreshToken to be updated to rt1, got %s", tok2.RefreshToken)
	}
}

// 2. Verify that when Mozilla provides a new refresh token, it is updated.
func TestSaveTokensUpdatesRefreshTokenWhenProvided(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	tok1 := &TokenResponse{AccessToken: "at1", RefreshToken: "rt1", ExpiresIn: 3600}
	if err := SaveTokensWithEmail(tok1, "user@example.com"); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	tok2 := &TokenResponse{AccessToken: "at2", RefreshToken: "rt2-new", ExpiresIn: 3600}
	if err := SaveTokens(tok2); err != nil {
		t.Fatalf("save with new rt failed: %v", err)
	}

	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("loading tokens failed: %v", err)
	}
	if loaded.RefreshToken != "rt2-new" {
		t.Errorf("expected updated rt2-new, got %s", loaded.RefreshToken)
	}
	if loaded.AccessToken != "at2" {
		t.Errorf("expected updated at2, got %s", loaded.AccessToken)
	}
}

// 3. Verify that empty or nil access token is rejected without corrupting existing credentials.
func TestSaveTokensRejectsEmptyAccessToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	tok1 := &TokenResponse{AccessToken: "at1-valid", RefreshToken: "rt1-valid", ExpiresIn: 3600}
	if err := SaveTokensWithEmail(tok1, "user@example.com"); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Try saving nil
	if err := SaveTokens(nil); err == nil {
		t.Error("expected error saving nil token response, got nil")
	}

	// Try saving empty access token
	tokInvalid := &TokenResponse{AccessToken: "", RefreshToken: "rt-other"}
	if err := SaveTokens(tokInvalid); err == nil {
		t.Error("expected error saving empty access token, got nil")
	}

	// Verify original credentials remain untouched
	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("loading tokens failed: %v", err)
	}
	if loaded.AccessToken != "at1-valid" || loaded.RefreshToken != "rt1-valid" {
		t.Errorf("expected original credentials preserved, got at=%s, rt=%s", loaded.AccessToken, loaded.RefreshToken)
	}
}

// 4. Verify that saving without any refresh token on initial save is rejected.
func TestSaveTokensRejectsMissingRefreshTokenOnInitialSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	tok := &TokenResponse{AccessToken: "at-only", ExpiresIn: 3600}
	if err := SaveTokens(tok); err == nil {
		t.Error("expected error when saving without any refresh token on initial save, got nil")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected no tokens file to be written on initial failure")
	}
}

// 5. Verify that failure during write does not destroy previous valid token file.
func TestSaveTokensPersistenceFailureSafe(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	tok1 := &TokenResponse{AccessToken: "at1-safe", RefreshToken: "rt1-safe", ExpiresIn: 3600}
	if err := SaveTokensWithEmail(tok1, "user@example.com"); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Point customTokenPath to a path where write will fail (non-existent directory in writePrivateFile)
	invalidPath := filepath.Join(dir, "nonexistent-dir", "subdir", "tokens.json")
	customTokenPath = invalidPath

	tok2 := &TokenResponse{AccessToken: "at2-bad", RefreshToken: "rt2-bad"}
	if err := SaveTokens(tok2); err == nil {
		t.Error("expected persistence failure for invalid path, got nil")
	}

	// Restore path and verify original file is intact
	customTokenPath = path
	loaded, err := LoadTokens()
	if err != nil {
		t.Fatalf("loading tokens failed: %v", err)
	}
	if loaded.AccessToken != "at1-safe" || loaded.RefreshToken != "rt1-safe" {
		t.Errorf("expected original file untouched, got at=%s, rt=%s", loaded.AccessToken, loaded.RefreshToken)
	}
}

// 6. Verify concurrent token saves and loads operate safely without data races or corruption.
func TestSaveTokensConcurrentSafe(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	customTokenPath = path
	defer func() { customTokenPath = "" }()

	// Initial token
	initial := &TokenResponse{AccessToken: "at-init", RefreshToken: "rt-fixed", ExpiresIn: 3600}
	if err := SaveTokensWithEmail(initial, "concurrent@example.com"); err != nil {
		t.Fatalf("initial save: %v", err)
	}

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				tok := &TokenResponse{
					AccessToken: fmt.Sprintf("at-%d", id),
					ExpiresIn:   3600,
				}
				_ = SaveTokens(tok)
			} else {
				loaded, err := LoadTokens()
				if err == nil && loaded != nil {
					if loaded.RefreshToken != "rt-fixed" {
						t.Errorf("worker %d: lost refresh token: %s", id, loaded.RefreshToken)
					}
				}
			}
		}(i)
	}

	wg.Wait()

	finalTokens, err := LoadTokens()
	if err != nil {
		t.Fatalf("final load failed: %v", err)
	}
	if finalTokens.RefreshToken != "rt-fixed" {
		t.Errorf("expected preserved rt-fixed after concurrency, got %s", finalTokens.RefreshToken)
	}
}
