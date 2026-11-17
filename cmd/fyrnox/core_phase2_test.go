package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"syscall"
	"testing"
	"time"
)

func TestIsEligibleTlsHandshakeFailure(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context deadline exceeded during handshake",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "unexpected EOF during handshake",
			err:      io.ErrUnexpectedEOF,
			expected: true,
		},
		{
			name:     "EOF during handshake",
			err:      io.EOF,
			expected: true,
		},
		{
			name:     "connection reset error string",
			err:      errors.New("read: connection reset by peer"),
			expected: true,
		},
		{
			name:     "wsaconnreset Windows error string",
			err:      errors.New("read tcp: wsaconnreset: An existing connection was forcibly closed by the remote host."),
			expected: true,
		},
		{
			name:     "ineligible: context canceled by user",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "ineligible: x509 unknown authority error",
			err:      x509.UnknownAuthorityError{},
			expected: false,
		},
		{
			name:     "ineligible: x509 hostname error",
			err:      x509.HostnameError{Host: "example.com"},
			expected: false,
		},
		{
			name:     "ineligible: wrapped certificate verification failure string",
			err:      fmt.Errorf("tls: failed to verify certificate: %w", errors.New("certificate signed by unknown authority")),
			expected: false,
		},
		{
			name:     "ineligible: TCP dial connection refused",
			err:      fmt.Errorf("dialing 127.0.0.1:2499: dial tcp 127.0.0.1:2499: connectex: %w", syscall.ECONNREFUSED),
			expected: false,
		},
		{
			name:     "ineligible: TCP dial no route to host",
			err:      fmt.Errorf("dialing 10.255.255.1:2499: dial tcp 10.255.255.1:2499: no route to host"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := isEligibleTlsHandshakeFailure(tc.err)
			if actual != tc.expected {
				t.Errorf("isEligibleTlsHandshakeFailure(%v) = %v, expected %v", tc.err, actual, tc.expected)
			}
		})
	}
}

func TestComputeTlsStageTimeouts(t *testing.T) {
	// Case 1: Context without deadline -> Stage 1 = 4s, Stage 2 = 3s, canFallback = true (combined max 7s)
	t.Run("no deadline context", func(t *testing.T) {
		ctx := context.Background()
		s1, s2, canFallback := computeTlsStageTimeouts(ctx)
		if !canFallback {
			t.Errorf("expected canFallback = true for no-deadline context")
		}
		if s1 != 4*time.Second {
			t.Errorf("expected stage1 = 4s, got %v", s1)
		}
		if s2 != 3*time.Second {
			t.Errorf("expected stage2 = 3s, got %v", s2)
		}
		if s1+s2 > 7*time.Second {
			t.Errorf("combined budget must not exceed 7s, got %v", s1+s2)
		}
	})

	// Case 2: R < 5s -> fallback disabled, Stage 1 uses remaining time
	t.Run("R < 5s tight deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		s1, s2, canFallback := computeTlsStageTimeouts(ctx)
		if canFallback {
			t.Errorf("expected canFallback = false for R < 5s")
		}
		if s2 != 0 {
			t.Errorf("expected stage2 = 0 when fallback disabled, got %v", s2)
		}
		if s1 <= 0 || s1 > 4*time.Second {
			t.Errorf("expected stage1 to use remaining time <= 4s, got %v", s1)
		}
	})

	// Case 3: R = 5s -> T1 = min(4.5s, 2.5s, 2.0s) = 2.0s; Stage 2 reserve >= 3.0s
	t.Run("R = 5s exact threshold", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5050*time.Millisecond)
		defer cancel()

		s1, _, canFallback := computeTlsStageTimeouts(ctx)
		if !canFallback {
			t.Errorf("expected canFallback = true for R >= 5s")
		}
		// With ~5.0s remaining: R-3s is ~2.0s; R/2 is ~2.5s -> min is ~2.0s
		if s1 > 2200*time.Millisecond || s1 < 1900*time.Millisecond {
			t.Errorf("expected stage1 ~ 2.0s, got %v", s1)
		}
	})

	// Case 4: R = 6s -> T1 = min(4.5s, 3.0s, 3.0s) = 3.0s; Stage 2 reserve >= 3.0s
	t.Run("R = 6s moderate deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 6050*time.Millisecond)
		defer cancel()

		s1, _, canFallback := computeTlsStageTimeouts(ctx)
		if !canFallback {
			t.Errorf("expected canFallback = true for R = 6s")
		}
		if s1 > 3200*time.Millisecond || s1 < 2900*time.Millisecond {
			t.Errorf("expected stage1 ~ 3.0s, got %v", s1)
		}
	})

	// Case 5: R = 12s -> T1 = min(4.5s, 6.0s, 9.0s) = 4.5s; Stage 2 reserve >= 7.5s
	t.Run("R = 12s generous deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 12050*time.Millisecond)
		defer cancel()

		s1, _, canFallback := computeTlsStageTimeouts(ctx)
		if !canFallback {
			t.Errorf("expected canFallback = true for R = 12s")
		}
		if s1 != 4500*time.Millisecond {
			t.Errorf("expected stage1 = 4.5s cap, got %v", s1)
		}
	})
}
