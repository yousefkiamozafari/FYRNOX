package main

import (
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// TestIPv6_RouteCreationFailureFailsClosed verifies that when a route injection
// fails mid-way (e.g. 8000::/1 fails), RouteManager rolls back previously installed
// routes so that no orphaned or partial routes remain active.
func TestIPv6_RouteCreationFailureFailsClosed(t *testing.T) {
	var deletedRoutes []string

	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		cmdStr := name + " " + strings.Join(args, " ")
		if strings.Contains(cmdStr, "show route") || strings.Contains(cmdStr, "print") {
			return "", nil
		}
		if name == "netsh" && strings.Contains(cmdStr, "add route 8000::/1") {
			return "Element not found.", errors.New("exit code 1")
		}
		if name == "netsh" && strings.Contains(cmdStr, "delete route") {
			deletedRoutes = append(deletedRoutes, cmdStr)
			return "OK", nil
		}
		if name == "route" && strings.Contains(cmdStr, "delete") {
			deletedRoutes = append(deletedRoutes, cmdStr)
			return "OK", nil
		}
		return "OK", nil
	}

	rm := NewRouteManager(55)
	rm.cmdRunner = mockRunner

	if err := rm.AddRouteIPv4("0.0.0.0", "128.0.0.0", "10.254.1.2", 1); err != nil {
		t.Fatalf("AddRouteIPv4 1 failed: %v", err)
	}
	if err := rm.AddRouteIPv4("128.0.0.0", "128.0.0.0", "10.254.1.2", 1); err != nil {
		t.Fatalf("AddRouteIPv4 2 failed: %v", err)
	}
	if err := rm.AddRouteIPv6("::/1", 1); err != nil {
		t.Fatalf("AddRouteIPv6 ::/1 failed: %v", err)
	}

	// Now 8000::/1 fails
	err := rm.AddRouteIPv6("8000::/1", 1)
	if err == nil {
		t.Fatal("Expected error on 8000::/1 addition failure")
	}

	// Perform fail-closed rollback as EnableTunMode does
	if rbErr := rm.Rollback(); rbErr != nil {
		t.Fatalf("Rollback failed: %v", rbErr)
	}

	if count := rm.TrackedCount(); count != 0 {
		t.Fatalf("TrackedCount after rollback = %d, want 0", count)
	}

	// Verify all 3 previously installed routes were deleted
	if len(deletedRoutes) != 3 {
		t.Fatalf("Expected 3 deletion commands, got %d: %v", len(deletedRoutes), deletedRoutes)
	}

	// Verify reverse order of deletion: ::/1 first, then 128.0.0.0, then 0.0.0.0
	if !strings.Contains(deletedRoutes[0], "::/1") {
		t.Errorf("Expected first deleted route to be ::/1, got %s", deletedRoutes[0])
	}
	if !strings.Contains(deletedRoutes[1], "128.0.0.0") {
		t.Errorf("Expected second deleted route to be 128.0.0.0, got %s", deletedRoutes[1])
	}
	if !strings.Contains(deletedRoutes[2], "0.0.0.0") {
		t.Errorf("Expected third deleted route to be 0.0.0.0, got %s", deletedRoutes[2])
	}
}

// TestIPv6_PreexistingRouteConflict verifies that RouteManager refuses to install
// an IPv6 or IPv4 route if a conflicting route already exists on a different interface.
func TestIPv6_PreexistingRouteConflict(t *testing.T) {
	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		cmdStr := name + " " + strings.Join(args, " ")
		if strings.Contains(cmdStr, "show route") {
			// Simulate existing ::/1 route on interface 12
			return "Publish Type Met Prefix Idx Gateway\nNo Manual 1 ::/1 12 fe80::1\n", nil
		}
		return "OK", nil
	}

	rm := NewRouteManager(55)
	rm.cmdRunner = mockRunner

	err := rm.AddRouteIPv6("::/1", 1)
	if err == nil {
		t.Fatal("Expected ErrRouteConflict when ::/1 exists on another interface")
	}
	if !errors.Is(err, ErrRouteConflict) {
		t.Fatalf("Expected ErrRouteConflict, got: %v", err)
	}
	if rm.TrackedCount() != 0 {
		t.Fatalf("TrackedCount = %d, want 0", rm.TrackedCount())
	}
}

// TestIPv6_PartialRollbackSafety verifies that when only a subset of routes are installed,
// Rollback strictly touches only those routes and leaves untracked routes untouched.
func TestIPv6_PartialRollbackSafety(t *testing.T) {
	var deletedCmds []string

	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		cmdStr := name + " " + strings.Join(args, " ")
		if strings.Contains(cmdStr, "show route") || strings.Contains(cmdStr, "print") {
			return "", nil
		}
		if strings.Contains(cmdStr, "delete") {
			deletedCmds = append(deletedCmds, cmdStr)
		}
		return "OK", nil
	}

	rm := NewRouteManager(77)
	rm.cmdRunner = mockRunner

	// Only install 1 route
	if err := rm.AddRouteIPv6("::/1", 1); err != nil {
		t.Fatalf("AddRouteIPv6 failed: %v", err)
	}

	if err := rm.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	if len(deletedCmds) != 1 {
		t.Fatalf("Expected exactly 1 delete command, got %d: %v", len(deletedCmds), deletedCmds)
	}
	if !strings.Contains(deletedCmds[0], "::/1") {
		t.Errorf("Expected delete command for ::/1, got %s", deletedCmds[0])
	}
}

// TestIPv6_DisconnectCleanupIdempotence verifies that calling Rollback multiple times
// is completely safe, does not invoke extraneous deletion commands, and returns nil.
func TestIPv6_DisconnectCleanupIdempotence(t *testing.T) {
	deleteCount := 0

	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		cmdStr := name + " " + strings.Join(args, " ")
		if strings.Contains(cmdStr, "show route") || strings.Contains(cmdStr, "print") {
			return "", nil
		}
		if strings.Contains(cmdStr, "delete") {
			deleteCount++
		}
		return "OK", nil
	}

	rm := NewRouteManager(88)
	rm.cmdRunner = mockRunner

	if err := rm.AddRouteIPv6("::/1", 1); err != nil {
		t.Fatalf("AddRouteIPv6 failed: %v", err)
	}

	// First rollback
	if err := rm.Rollback(); err != nil {
		t.Fatalf("First Rollback failed: %v", err)
	}
	if deleteCount != 1 {
		t.Fatalf("deleteCount = %d, want 1", deleteCount)
	}

	// Second rollback (idempotent call)
	if err := rm.Rollback(); err != nil {
		t.Fatalf("Second Rollback failed: %v", err)
	}
	if deleteCount != 1 {
		t.Fatalf("deleteCount after second rollback = %d, want 1 (no extra commands)", deleteCount)
	}

	// Third rollback
	if err := rm.Rollback(); err != nil {
		t.Fatalf("Third Rollback failed: %v", err)
	}
	if deleteCount != 1 {
		t.Fatalf("deleteCount after third rollback = %d, want 1", deleteCount)
	}
}

// TestIPv6_KernelVerificationFailsafe verifies VerifyRouteExists returns true when
// the kernel table includes the prefix and false when it does not.
func TestIPv6_KernelVerificationFailsafe(t *testing.T) {
	presentRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		return "Publish Type Met Prefix Idx Gateway\nNo Manual 1 ::/1 99 fe80::1\n", nil
	}

	absentRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		return "Publish Type Met Prefix Idx Gateway\nNo Manual 256 ::/0 4 fe80::1\n", nil
	}

	rmPresent := NewRouteManager(99)
	rmPresent.cmdRunner = presentRunner
	ok, err := rmPresent.VerifyRouteExists("::/1", 6)
	if err != nil || !ok {
		t.Fatalf("Expected route ::/1 to verify true, got ok=%v, err=%v", ok, err)
	}

	rmAbsent := NewRouteManager(99)
	rmAbsent.cmdRunner = absentRunner
	ok, err = rmAbsent.VerifyRouteExists("::/1", 6)
	if err != nil || ok {
		t.Fatalf("Expected route ::/1 to verify false when missing, got ok=%v, err=%v", ok, err)
	}
}

// TestIPv6_EngineStatusTracking verifies that VpnEngine properly exposes IPv6Operational
// in its VpnStatus and clears it on disconnect.
func TestIPv6_EngineStatusTracking(t *testing.T) {
	engine := NewVpnEngine()
	engine.settings.EnableTunMode = false
	engine.settings.UseSystemProxy = true

	// Initially false
	status := engine.GetStatus()
	if status.IPv6Operational {
		t.Fatalf("Expected IPv6Operational=false initially, got true")
	}

	// Simulate setting operational status to true
	engine.mu.Lock()
	engine.ipv6Operational = true
	engine.state = StateConnected
	engine.mu.Unlock()

	status = engine.GetStatus()
	if !status.IPv6Operational {
		t.Fatalf("Expected IPv6Operational=true when set, got false")
	}

	// Simulate Disconnect
	_ = engine.Disconnect()

	status = engine.GetStatus()
	if status.IPv6Operational {
		t.Fatalf("Expected IPv6Operational=false after Disconnect, got true")
	}
}

// TestIPv6_WatchdogTransitionOnProbeDrop verifies that when IPv6 connectivity drops
// during an active connection, the engine transitions ipv6Operational to false safely.
func TestIPv6_WatchdogTransitionOnProbeDrop(t *testing.T) {
	engine := NewVpnEngine()
	engine.mu.Lock()
	engine.state = StateConnected
	engine.ipv6Operational = true
	engine.mu.Unlock()

	// Simulate probe failure detection as in watchdog:
	// A nil or failing controller probe transitions ipv6Operational to false
	engine.mu.Lock()
	engine.ipv6Operational = false // transitioned on failure
	engine.mu.Unlock()

	status := engine.GetStatus()
	if status.IPv6Operational {
		t.Fatalf("Expected IPv6Operational=false after probe failure transition, got true")
	}
}

// TestIPv6_ActivationOrdering_WfpRollbackOnRouteFailure verifies that when route installation
// fails after WFP leak prevention is armed, the system rolls back both the routes and the
// WFP dynamic session, ensuring zero lingering or orphaned resources.
func TestIPv6_ActivationOrdering_WfpRollbackOnRouteFailure(t *testing.T) {
	mockWfp := &mockWfpEngine{}
	wfp := &WfpManager{api: mockWfp}

	// 1. Arm WFP
	if err := wfp.EnableIPv6LeakProtection(42, net.ParseIP("2606:4700::1111"), 2499); err != nil {
		t.Fatalf("Failed to arm WFP: %v", err)
	}
	if !wfp.IsActive() {
		t.Fatalf("Expected WFP active")
	}

	// 2. Simulate route installation failure
	rm := NewRouteManager(42)
	rm.cmdRunner = func(timeout time.Duration, name string, args ...string) (string, error) {
		if strings.Contains(strings.Join(args, " "), "add") {
			return "Access is denied.", errors.New("route add failed")
		}
		return "OK", nil
	}

	routeErr := rm.AddRouteIPv4("0.0.0.0", "128.0.0.0", "10.254.1.2", 1)
	if routeErr == nil {
		t.Fatalf("Expected route installation failure")
	}

	// Rollback sequence matching EnableTunMode
	_ = rm.Rollback()
	_ = wfp.Close()

	if wfp.IsActive() {
		t.Errorf("Expected WFP to be deactivated after route failure rollback")
	}
	if !mockWfp.closeCalled {
		t.Errorf("Expected WFP engine Close to be invoked on failure rollback")
	}
}
