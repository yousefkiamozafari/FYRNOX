package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestKillSwitch_NonAdminRejected(t *testing.T) {
	called := false
	ks := &KillSwitchManager{
		isAdminFn: func() bool { return false },
		cmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			called = true
			return "", nil
		},
	}

	err := ks.Arm("198.51.100.10")
	if err == nil {
		t.Fatalf("expected error when Arm is called without Admin privileges")
	}
	if ks.IsActive() {
		t.Fatalf("expected KillSwitch to remain inactive when non-admin")
	}
	if called {
		t.Fatalf("expected no netsh commands to execute when non-admin")
	}
}

func TestKillSwitch_CommandFailureAndPartialRollback(t *testing.T) {
	var deleted []string
	ks := &KillSwitchManager{
		isAdminFn: func() bool { return true },
		cmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "delete rule") {
				for _, a := range args {
					if strings.HasPrefix(a, "name=") {
						deleted = append(deleted, strings.TrimPrefix(a, "name="))
					}
				}
				return "Ok.", nil
			}
			// Fail on the 3rd rule (AllowVPN)
			if strings.Contains(joined, "add rule") && strings.Contains(joined, "FYRNOX-KillSwitch-AllowVPN") {
				return "Access is denied.", errors.New("exit status 1")
			}
			return "Ok.", nil
		},
	}

	err := ks.Arm("198.51.100.10")
	if err == nil {
		t.Fatalf("expected Arm to fail on AllowVPN rule error")
	}
	if ks.IsActive() {
		t.Fatalf("expected KillSwitch inactive after partial failure")
	}

	// Verify previously installed rules (AllowDHCP, AllowLAN) were rolled back in reverse order
	rbCount := 0
	for _, d := range deleted {
		if d == "FYRNOX-KillSwitch-AllowDHCP" || d == "FYRNOX-KillSwitch-AllowLAN" {
			rbCount++
		}
	}
	if rbCount < 2 {
		t.Fatalf("expected AllowDHCP and AllowLAN to be rolled back, deleted=%v", deleted)
	}
}

func TestKillSwitch_RollbackFailureReported(t *testing.T) {
	ks := &KillSwitchManager{
		isAdminFn: func() bool { return true },
		cmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "add rule") && strings.Contains(joined, "FYRNOX-KillSwitch-Block") {
				return "failed", errors.New("add error")
			}
			if strings.Contains(joined, "delete rule") && strings.Contains(joined, "FYRNOX-KillSwitch-AllowLAN") {
				return "fatal delete error", errors.New("delete error")
			}
			return "Ok.", nil
		},
	}

	err := ks.Arm("198.51.100.10")
	if err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected rollback failure to be reported, got: %v", err)
	}
	if ks.IsActive() {
		t.Fatalf("expected KillSwitch inactive when Arm+rollback fails")
	}
}

func TestKillSwitch_SuccessfulActivationAndIdempotentDisarm(t *testing.T) {
	var commands []string
	ks := &KillSwitchManager{
		isAdminFn: func() bool { return true },
		cmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			commands = append(commands, joined)
			if strings.Contains(joined, "delete rule") {
				return "No rules match the specified criteria.", errors.New("exit status 1")
			}
			return "Ok.", nil
		},
	}

	if err := ks.Arm("203.0.113.5"); err != nil {
		t.Fatalf("expected Arm to succeed, got: %v", err)
	}
	if !ks.IsActive() {
		t.Fatalf("expected KillSwitch active after Arm")
	}

	// Verify IPv6 loopback and link-local are included in AllowLAN
	foundIPv6LAN := false
	verifiedRules := 0
	for _, c := range commands {
		if strings.Contains(c, "FYRNOX-KillSwitch-AllowLAN") && strings.Contains(c, "::1,fe80::/10") {
			foundIPv6LAN = true
		}
		if strings.Contains(c, "show rule") {
			verifiedRules++
		}
	}
	if !foundIPv6LAN {
		t.Errorf("expected AllowLAN rule to include ::1,fe80::/10")
	}
	if verifiedRules != 4 {
		t.Errorf("expected 4 rules verified via show rule, got %d", verifiedRules)
	}

	// Repeated Disarm must be idempotent and safe
	if err := ks.Disarm(); err != nil {
		t.Fatalf("first Disarm failed: %v", err)
	}
	if ks.IsActive() {
		t.Fatalf("expected inactive after Disarm")
	}
	if err := ks.Disarm(); err != nil {
		t.Fatalf("second Disarm failed: %v", err)
	}
}

func TestKillSwitch_DisarmFailurePreservesActiveState(t *testing.T) {
	failBlockDelete := true
	ks := &KillSwitchManager{
		isAdminFn: func() bool { return true },
		cmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "delete rule") && strings.Contains(joined, "FYRNOX-KillSwitch-Block") && failBlockDelete {
				return "Access is denied.", errors.New("delete failed")
			}
			return "Ok.", nil
		},
	}

	if err := ks.Arm("203.0.113.5"); err != nil {
		t.Fatalf("Arm failed: %v", err)
	}
	if err := ks.Disarm(); err == nil {
		t.Fatalf("expected Disarm to return error when Block rule deletion fails")
	}
	if !ks.IsActive() {
		t.Fatalf("expected KillSwitch to remain active when Disarm fails to delete Block rule")
	}

	// Retry Disarm once failure resolves
	failBlockDelete = false
	if err := ks.Disarm(); err != nil {
		t.Fatalf("retry Disarm failed: %v", err)
	}
	if ks.IsActive() {
		t.Fatalf("expected KillSwitch inactive after successful retry Disarm")
	}
}
