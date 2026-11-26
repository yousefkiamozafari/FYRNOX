package main

import (
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestRouteManager_TrackingAndRollback(t *testing.T) {
	var executedCmds []string

	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		cmdStr := name + " " + strings.Join(args, " ")
		executedCmds = append(executedCmds, cmdStr)
		if name == "netsh" && strings.Contains(cmdStr, "show route") {
			return "Publish Type Met Prefix Idx Gateway\nNo Manual 256 ::/0 4 fe80::1\n", nil
		}
		if name == "route" && strings.Contains(cmdStr, "print") {
			return "Network Destination Netmask Gateway Interface Metric\n0.0.0.0 0.0.0.0 192.168.1.1 192.168.1.50 25\n", nil
		}
		return "OK", nil
	}

	rm := NewRouteManager(42)
	rm.cmdRunner = mockRunner

	if err := rm.AddRouteIPv4("0.0.0.0", "128.0.0.0", "10.254.1.2", 1); err != nil {
		t.Fatalf("AddRouteIPv4 failed: %v", err)
	}
	if err := rm.AddRouteIPv4("128.0.0.0", "128.0.0.0", "10.254.1.2", 1); err != nil {
		t.Fatalf("AddRouteIPv4 2 failed: %v", err)
	}
	if err := rm.AddRouteIPv6("::/1", 1); err != nil {
		t.Fatalf("AddRouteIPv6 failed: %v", err)
	}
	if err := rm.AddRouteIPv6("8000::/1", 1); err != nil {
		t.Fatalf("AddRouteIPv6 2 failed: %v", err)
	}

	if count := rm.TrackedCount(); count != 4 {
		t.Fatalf("TrackedCount = %d, want 4", count)
	}
	if !rm.IsTracked("::/1") || !rm.IsTracked("0.0.0.0") {
		t.Fatalf("Expected routes to be tracked")
	}

	// Test rollback
	executedCmds = nil
	if err := rm.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if count := rm.TrackedCount(); count != 0 {
		t.Fatalf("TrackedCount after rollback = %d, want 0", count)
	}

	// Verify rollback executed in reverse order
	if len(executedCmds) != 4 {
		t.Fatalf("Executed deletion commands = %d, want 4", len(executedCmds))
	}
	if !strings.Contains(executedCmds[0], "8000::/1") {
		t.Errorf("First rollback command should be 8000::/1, got: %s", executedCmds[0])
	}
	if !strings.Contains(executedCmds[3], "0.0.0.0") {
		t.Errorf("Last rollback command should be 0.0.0.0, got: %s", executedCmds[3])
	}
}

func TestRouteManager_ConflictRefusal(t *testing.T) {
	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		cmdStr := name + " " + strings.Join(args, " ")
		if strings.Contains(cmdStr, "show route") {
			// Prefix ::/1 already exists on interface index 99!
			return "Publish Type Met Prefix Idx Gateway\nNo Manual 1 ::/1 99 fe80::1\n", nil
		}
		return "OK", nil
	}

	rm := NewRouteManager(42)
	rm.cmdRunner = mockRunner

	err := rm.AddRouteIPv6("::/1", 1)
	if err == nil {
		t.Fatal("Expected ErrRouteConflict when prefix already exists on ifIndex 99")
	}
	if !errors.Is(err, ErrRouteConflict) {
		t.Fatalf("Expected ErrRouteConflict, got %v", err)
	}
	if rm.TrackedCount() != 0 {
		t.Fatalf("TrackedCount = %d, want 0 on rejected route", rm.TrackedCount())
	}
}

func TestRouteManager_CommandFailureReturnsError(t *testing.T) {
	mockRunner := func(timeout time.Duration, name string, args ...string) (string, error) {
		if name == "netsh" && strings.Contains(strings.Join(args, " "), "add route") {
			return "Access is denied.", errors.New("exit code 1")
		}
		return "OK", nil
	}

	rm := NewRouteManager(42)
	rm.cmdRunner = mockRunner

	err := rm.AddRouteIPv6("::/1", 1)
	if err == nil {
		t.Fatal("Expected error when netsh command fails")
	}
	if rm.TrackedCount() != 0 {
		t.Fatalf("TrackedCount = %d, want 0 after failure", rm.TrackedCount())
	}
}

func TestRouteManager_IPHelperTableReader(t *testing.T) {
	mockTable := func(family uint16) ([]MibIpforwardRow2, error) {
		return []MibIpforwardRow2{
			{
				InterfaceIndex:          42,
				DestinationPrefixFamily: 23, // AF_INET6
				PrefixLength:            1,  // ::/1
			},
		}, nil
	}

	rm := NewRouteManager(42)
	rm.table2Reader = mockTable

	ok, err := rm.VerifyRouteExists("::/1", 6)
	if err != nil || !ok {
		t.Fatalf("Expected VerifyRouteExists via IP Helper to return true, got ok=%v, err=%v", ok, err)
	}

	// Missing route
	ok, err = rm.VerifyRouteExists("8000::/1", 6)
	if err != nil || ok {
		t.Fatalf("Expected VerifyRouteExists for missing route to return false, got ok=%v, err=%v", ok, err)
	}
}

func TestRouteManager_MoreSpecificPhysicalRoutes(t *testing.T) {
	mockTable := func(family uint16) ([]MibIpforwardRow2, error) {
		var r1 MibIpforwardRow2
		r1.InterfaceIndex = 4 // physical adapter
		r1.DestinationPrefixFamily = 23
		r1.PrefixLength = 64 // on-link /64
		// fe80:: in DestinationPrefixData[6:22]
		r1.DestinationPrefixData[6] = 0xfe
		r1.DestinationPrefixData[7] = 0x80

		var r2 MibIpforwardRow2
		r2.InterfaceIndex = 42 // tun adapter
		r2.DestinationPrefixFamily = 23
		r2.PrefixLength = 1 // ::/1

		return []MibIpforwardRow2{r1, r2}, nil
	}

	rm := NewRouteManager(42)
	rm.table2Reader = mockTable

	specific, err := rm.DetectMoreSpecificPhysicalRoutes()
	if err != nil {
		t.Fatalf("DetectMoreSpecificPhysicalRoutes failed: %v", err)
	}
	if len(specific) != 1 {
		t.Fatalf("Expected 1 more-specific route, got %d: %v", len(specific), specific)
	}
	if !strings.Contains(specific[0], "ifIndex=4") || !strings.Contains(specific[0], "/64") {
		t.Errorf("Unexpected specific route format: %s", specific[0])
	}
}

func TestRouteManager_GetBestRoute2(t *testing.T) {
	mockBest := func(destIP net.IP) (uint32, uint8, error) {
		return 42, 1, nil // forwards via tun adapter 42
	}

	rm := NewRouteManager(42)
	rm.bestRoute2Reader = mockBest

	ok, err := rm.VerifyBestRoute("2606:4700:4700::1111", 42)
	if err != nil || !ok {
		t.Fatalf("Expected VerifyBestRoute to return true, got ok=%v, err=%v", ok, err)
	}

	// Different interface expected
	ok, err = rm.VerifyBestRoute("2606:4700:4700::1111", 99)
	if err != nil || ok {
		t.Fatalf("Expected VerifyBestRoute for interface 99 to return false, got ok=%v, err=%v", ok, err)
	}
}

func TestCleanupOrphanedFyrnoxRoutes_OwnershipSafety(t *testing.T) {
	// Helper to build IPv4 row
	makeV4Row := func(dest, gw net.IP, ifIdx uint32) MibIpforwardRow2 {
		var r MibIpforwardRow2
		r.InterfaceIndex = ifIdx
		r.DestinationPrefixFamily = windows.AF_INET
		r.PrefixLength = 1
		copy(r.DestinationPrefixData[4:8], dest.To4())
		r.NextHopFamily = windows.AF_INET
		copy(r.NextHopData[4:8], gw.To4())
		return r
	}
	// Helper to build IPv6 row
	makeV6Row := func(dest net.IP, ifIdx uint32) MibIpforwardRow2 {
		var r MibIpforwardRow2
		r.InterfaceIndex = ifIdx
		r.DestinationPrefixFamily = windows.AF_INET6
		r.PrefixLength = 1
		copy(r.DestinationPrefixData[6:22], dest.To16())
		return r
	}

	var cmds []string
	deps := RouteCleanupDeps{
		IsTunActive:   func() bool { return false },
		FyrnoxIfIndex: func() uint32 { return 0 }, // FyrnoxTun adapter already gone after crash
		LiveIfIndices: func() (map[uint32]bool, error) {
			return map[uint32]bool{7: true, 12: true}, nil // ifIndex 7 is an unrelated active VPN (e.g. xray_tun)
		},
		Table2Reader: func(family uint16) ([]MibIpforwardRow2, error) {
			if family == windows.AF_INET {
				return []MibIpforwardRow2{
					// 4. Confirmed FYRNOX-owned stale IPv4 routes (NextHop 10.254.1.2 on dead ifIndex 99) -> MUST DELETE
					makeV4Row(net.IPv4(0, 0, 0, 0), net.IPv4(10, 254, 1, 2), 99),
					makeV4Row(net.IPv4(128, 0, 0, 0), net.IPv4(10, 254, 1, 2), 99),
					// 5. Active unrelated VPN 0.0.0.0/1 route (NextHop 10.8.0.1 on live ifIndex 7) -> MUST PRESERVE
					makeV4Row(net.IPv4(0, 0, 0, 0), net.IPv4(10, 8, 0, 1), 7),
					// 2. Dead unrelated interface index > 0 with non-FYRNOX gateway -> MUST PRESERVE & REPORT UNRESOLVED
					makeV4Row(net.IPv4(0, 0, 0, 0), net.IPv4(10, 9, 0, 1), 88),
				}, nil
			}
			return []MibIpforwardRow2{
				// 4. Confirmed FYRNOX-owned stale IPv6 routes (correlated via 10.254.1.2 on ifIndex 99) -> MUST DELETE
				makeV6Row(net.IPv6zero, 99),
				makeV6Row(net.ParseIP("8000::"), 99),
				// 1. Interface index zero -> MUST PRESERVE & REPORT UNRESOLVED
				makeV6Row(net.IPv6zero, 0),
				// 2 & 3. Dead unrelated interface index > 0 with missing ownership evidence -> MUST PRESERVE & REPORT UNRESOLVED
				makeV6Row(net.IPv6zero, 88),
				// 5. Active unrelated VPN ::/1 route on live ifIndex 7 -> MUST PRESERVE
				makeV6Row(net.IPv6zero, 7),
			}, nil
		},
		CmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			cmds = append(cmds, name+" "+strings.Join(args, " "))
			return "", nil
		},
	}

	res, err := CleanupOrphanedFyrnoxRoutes(deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 4 {
		t.Fatalf("expected 4 FYRNOX-owned orphaned routes removed, got %d (cmds=%v)", res.Removed, cmds)
	}
	if len(res.Unresolved) != 3 {
		t.Fatalf("expected 3 unresolved routes reported (v4 ifIndex=88, v6 ifIndex=0, v6 ifIndex=88), got %d: %v", len(res.Unresolved), res.Unresolved)
	}
	for _, c := range cmds {
		if strings.Contains(c, "10.8.0.1") || strings.Contains(c, "if 7") || strings.Contains(c, "interface=7") ||
			strings.Contains(c, "10.9.0.1") || strings.Contains(c, "if 88") || strings.Contains(c, "interface=88") {
			t.Fatalf("unrelated route was deleted: %s", c)
		}
		if strings.Contains(c, "netsh interface ipv6 delete") && !strings.Contains(c, "interface=") {
			t.Fatalf("IPv6 deletion omitted explicit interface= qualification: %s", c)
		}
	}

	// Active tunnel must never have its routes removed
	deps.IsTunActive = func() bool { return true }
	cmds = nil
	resActive, err := CleanupOrphanedFyrnoxRoutes(deps)
	if err != nil || resActive.Removed != 0 || len(cmds) != 0 {
		t.Fatalf("expected active tunnel routes to be preserved, removed=%d cmds=%v err=%v", resActive.Removed, cmds, err)
	}
}

func TestCleanupOrphanedFyrnoxRoutes_DeadUnrelatedInterfacePreserved(t *testing.T) {
	var cmds []string
	deps := RouteCleanupDeps{
		IsTunActive:   func() bool { return false },
		FyrnoxIfIndex: func() uint32 { return 0 },
		LiveIfIndices: func() (map[uint32]bool, error) {
			return map[uint32]bool{7: true}, nil
		},
		Table2Reader: func(family uint16) ([]MibIpforwardRow2, error) {
			if family == windows.AF_INET6 {
				var r MibIpforwardRow2
				r.InterfaceIndex = 55 // dead interface > 0, no FyrnoxTun and no 10.254.1.2 route
				r.DestinationPrefixFamily = windows.AF_INET6
				r.PrefixLength = 1
				copy(r.DestinationPrefixData[6:22], net.IPv6zero.To16())
				return []MibIpforwardRow2{r}, nil
			}
			return nil, nil
		},
		CmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			cmds = append(cmds, name+" "+strings.Join(args, " "))
			return "", nil
		},
	}

	res, err := CleanupOrphanedFyrnoxRoutes(deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 0 || len(cmds) != 0 {
		t.Fatalf("expected dead unrelated interface route to remain untouched, removed=%d cmds=%v", res.Removed, cmds)
	}
	if len(res.Unresolved) != 1 {
		t.Fatalf("expected 1 unresolved route reported, got %d: %v", len(res.Unresolved), res.Unresolved)
	}
}

func TestCleanupOrphanedFyrnoxRoutes_SkipsZeroInterfaceIndex(t *testing.T) {
	var cmds []string
	deps := RouteCleanupDeps{
		IsTunActive:   func() bool { return false },
		FyrnoxIfIndex: func() uint32 { return 0 },
		LiveIfIndices: func() (map[uint32]bool, error) {
			return map[uint32]bool{7: true}, nil
		},
		Table2Reader: func(family uint16) ([]MibIpforwardRow2, error) {
			if family == windows.AF_INET6 {
				var r MibIpforwardRow2
				r.InterfaceIndex = 0
				r.DestinationPrefixFamily = windows.AF_INET6
				r.PrefixLength = 1
				copy(r.DestinationPrefixData[6:22], net.IPv6zero.To16())
				return []MibIpforwardRow2{r}, nil
			}
			return nil, nil
		},
		CmdRunner: func(timeout time.Duration, name string, args ...string) (string, error) {
			cmds = append(cmds, name+" "+strings.Join(args, " "))
			return "", nil
		},
	}

	res, err := CleanupOrphanedFyrnoxRoutes(deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 0 || len(cmds) != 0 {
		t.Fatalf("expected InterfaceIndex=0 route to be skipped, removed=%d cmds=%v", res.Removed, cmds)
	}
	if len(res.Unresolved) != 1 {
		t.Fatalf("expected 1 unresolved route reported for ifIndex=0, got %d", len(res.Unresolved))
	}
}
