package vpnclient

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
)

const defaultFirefoxMajor = "158"

var firefoxMajor atomic.Value

// SetFirefoxVersion sets the reported Firefox release major version.
func SetFirefoxVersion(version string) {
	major, _, _ := strings.Cut(strings.TrimSpace(version), ".")
	if major == "" || strings.Trim(major, "0123456789") != "" {
		return
	}
	firefoxMajor.Store(major)
}

// FirefoxUserAgent returns a dynamic, platform-appropriate Firefox desktop User-Agent string.
func FirefoxUserAgent() string {
	major, _ := firefoxMajor.Load().(string)
	if major == "" {
		major = defaultFirefoxMajor
	}
	return fmt.Sprintf("Mozilla/5.0 (%s; rv:%s.0) Gecko/20100101 Firefox/%s.0", firefoxPlatform(runtime.GOOS, runtime.GOARCH), major, major)
}

func firefoxPlatform(goos, goarch string) string {
	switch goos {
	case "windows", "android":
		return "Windows NT 10.0; Win64; x64"
	case "darwin":
		return "Macintosh; Intel Mac OS X 10.15"
	}
	if goarch == "arm64" {
		return "X11; Linux aarch64"
	}
	return "X11; Linux x86_64"
}

// applyFirefoxHeaders sets the common control-plane request headers matching Firefox desktop.
func applyFirefoxHeaders(req *http.Request) {
	req.Header.Set("User-Agent", FirefoxUserAgent())
	req.Header.Set("Accept", "*/*")
	// Keep net/http from adding gzip as the sole encoding. That header was
	// associated with HTTP 406 on the HTTP/1.1 OAuth path in live A/B tests.
	req.Header.Set("Accept-Encoding", "identity")
}

// applyGuardianHeaders sets the request headers for Mozilla Guardian requests.
func applyGuardianHeaders(req *http.Request, accessToken string) {
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "max-age=0")
	applyFirefoxHeaders(req)
}

// applyMozillaVPNHeaders provides compatibility for existing callers.
func applyMozillaVPNHeaders(req *http.Request) {
	applyFirefoxHeaders(req)
}
