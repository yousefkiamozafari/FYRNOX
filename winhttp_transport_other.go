//go:build !windows

package vpnclient

import (
	"errors"
	"net/http"
)

// shouldUseWinHTTP always returns false on non-Windows platforms.
func shouldUseWinHTTP(_ ControlPlaneTransportOptions) bool {
	return false
}

// IsWinHTTPTransport always returns false on non-Windows platforms.
func IsWinHTTPTransport(_ http.RoundTripper) bool {
	return false
}

// newWinHTTPTransport is unavailable on non-Windows platforms.
func newWinHTTPTransport(_ ControlPlaneTransportOptions) (http.RoundTripper, error) {
	return nil, errors.New("winhttp: unavailable on non-Windows platform")
}
