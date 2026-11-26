//go:build windows

package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	ErrRouteConflict = errors.New("route already exists on a different interface")
	ErrRouteNotFound = errors.New("route not found in kernel routing table")

	modiphlpapi            = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetIpForwardTable2 = modiphlpapi.NewProc("GetIpForwardTable2")
	procFreeMibTable       = modiphlpapi.NewProc("FreeMibTable")
	procGetBestRoute2      = modiphlpapi.NewProc("GetBestRoute2")
)

type MibIpforwardRow2 struct {
	InterfaceLuid           uint64
	InterfaceIndex          uint32
	DestinationPrefixFamily uint16
	DestinationPrefixData   [26]byte
	PrefixLength            uint8
	_pad0                   [3]byte
	NextHopFamily           uint16
	NextHopData             [26]byte
	SitePrefixLength        uint8
	ValidLifetime           uint32
	PreferredLifetime       uint32
	Metric                  uint32
	Protocol                uint32
	Loopback                uint8
	AutoconfigureAddress    uint8
	Publish                 uint8
	Immortal                uint8
	Age                     uint32
	Origin                  uint32
}

type MibIpforwardTable2 struct {
	NumEntries uint32
	_pad       uint32
	Table      [1]MibIpforwardRow2
}

type SockaddrIn6 struct {
	Sin6Family   uint16
	Sin6Port     uint16
	Sin6Flowinfo uint32
	Sin6Addr     [16]byte
	Sin6ScopeId  uint32
}

type RouteRecord struct {
	DestinationPrefix string
	Mask              string
	NextHop           string
	InterfaceIndex    int
	Metric            int
	Family            int // 4 or 6
}

type RouteManager struct {
	mu               sync.Mutex
	tunIfIndex       int
	trackedRoutes    []RouteRecord
	cmdRunner        func(timeout time.Duration, name string, args ...string) (string, error)
	table2Reader     func(family uint16) ([]MibIpforwardRow2, error)
	bestRoute2Reader func(destIP net.IP) (uint32, uint8, error)
}

func readIpForwardTable2(family uint16) ([]MibIpforwardRow2, error) {
	var tablePtr *MibIpforwardTable2
	r1, _, err := procGetIpForwardTable2.Call(uintptr(family), uintptr(unsafe.Pointer(&tablePtr)))
	if r1 != 0 {
		return nil, fmt.Errorf("GetIpForwardTable2 failed (code %d): %v", r1, err)
	}
	if tablePtr == nil {
		return nil, nil
	}
	defer procFreeMibTable.Call(uintptr(unsafe.Pointer(tablePtr)))

	num := tablePtr.NumEntries
	rows := make([]MibIpforwardRow2, num)
	rowSize := unsafe.Sizeof(MibIpforwardRow2{})
	for i := uint32(0); i < num; i++ {
		row := (*MibIpforwardRow2)(unsafe.Add(unsafe.Pointer(&tablePtr.Table[0]), uintptr(i)*rowSize))
		rows[i] = *row
	}
	return rows, nil
}

func queryBestRoute2(destIP net.IP) (uint32, uint8, error) {
	ip16 := destIP.To16()
	if ip16 == nil {
		return 0, 0, fmt.Errorf("invalid destination ip: %v", destIP)
	}

	var dst SockaddrIn6
	dst.Sin6Family = windows.AF_INET6
	copy(dst.Sin6Addr[:], ip16)

	var bestRoute [104]byte
	var bestSrc [28]byte

	r1, _, err := procGetBestRoute2.Call(
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&dst)),
		0,
		uintptr(unsafe.Pointer(&bestRoute)),
		uintptr(unsafe.Pointer(&bestSrc)),
	)
	if r1 != 0 {
		return 0, 0, fmt.Errorf("GetBestRoute2 failed (code %d): %v", r1, err)
	}
	ifIndex := *(*uint32)(unsafe.Pointer(&bestRoute[8]))
	prefixLen := bestRoute[44]
	return ifIndex, prefixLen, nil
}

func NewRouteManager(tunIfIndex int) *RouteManager {
	return &RouteManager{
		tunIfIndex:       tunIfIndex,
		cmdRunner:        runSilentCmdWithTimeout,
		table2Reader:     readIpForwardTable2,
		bestRoute2Reader: queryBestRoute2,
	}
}

// CheckConflict verifies whether a prefix is already in use by a pre-existing interface.
func (rm *RouteManager) CheckConflict(prefix, mask string, family int) (bool, error) {
	if rm.cmdRunner == nil {
		rm.cmdRunner = runSilentCmdWithTimeout
	}

	if family == 4 {
		out, _ := rm.cmdRunner(1500*time.Millisecond, "route", "print", prefix)
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 5 && fields[0] == prefix {
				// fields[0]=Destination, fields[1]=Netmask, fields[2]=Gateway, fields[3]=Interface
				netmask := fields[1]
				if mask != "" && netmask != mask {
					// Different mask (e.g. 0.0.0.0/1 vs 0.0.0.0/0) -> no conflict
					continue
				}
				iface := fields[3]
				if rm.tunIfIndex > 0 && iface != strconv.Itoa(rm.tunIfIndex) && iface != "10.254.1.2" {
					return true, nil
				}
			}
		}
		return false, nil
	}

	// IPv6 conflict check
	out, err := rm.cmdRunner(1500*time.Millisecond, "netsh", "interface", "ipv6", "show", "route", "store=active")
	if err != nil {
		return false, nil
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// Format: Publish Type Met Prefix Idx Gateway/Interface
			for i, field := range fields {
				if field == prefix && i+1 < len(fields) {
					idxStr := fields[i+1]
					if rm.tunIfIndex > 0 && idxStr != strconv.Itoa(rm.tunIfIndex) {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

// AddRouteIPv4 installs an IPv4 route and tracks it in the internal ledger.
func (rm *RouteManager) AddRouteIPv4(prefix, mask, nextHop string, metric int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.cmdRunner == nil {
		rm.cmdRunner = runSilentCmdWithTimeout
	}

	conflict, err := rm.CheckConflict(prefix, mask, 4)
	if err != nil {
		return fmt.Errorf("conflict check error for %s: %w", prefix, err)
	}
	if conflict {
		return fmt.Errorf("%w: %s (mask %s)", ErrRouteConflict, prefix, mask)
	}

	args := []string{"add", prefix, "mask", mask, nextHop, "metric", strconv.Itoa(metric)}
	if rm.tunIfIndex > 0 {
		args = append(args, "if", strconv.Itoa(rm.tunIfIndex))
	}

	out, err := rm.cmdRunner(2*time.Second, "route", args...)
	if err != nil {
		return fmt.Errorf("failed to add IPv4 route %s (mask %s -> %s): %v (%s)", prefix, mask, nextHop, err, strings.TrimSpace(out))
	}

	rm.trackedRoutes = append(rm.trackedRoutes, RouteRecord{
		DestinationPrefix: prefix,
		Mask:              mask,
		NextHop:           nextHop,
		InterfaceIndex:    rm.tunIfIndex,
		Metric:            metric,
		Family:            4,
	})
	return nil
}

// AddRouteIPv6 installs an IPv6 on-link route on FyrnoxTun and tracks it in the internal ledger.
func (rm *RouteManager) AddRouteIPv6(prefix string, metric int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.cmdRunner == nil {
		rm.cmdRunner = runSilentCmdWithTimeout
	}

	conflict, err := rm.CheckConflict(prefix, "", 6)
	if err != nil {
		return fmt.Errorf("conflict check error for %s: %w", prefix, err)
	}
	if conflict {
		return fmt.Errorf("%w: %s", ErrRouteConflict, prefix)
	}

	args := []string{"interface", "ipv6", "add", "route", prefix}
	if rm.tunIfIndex > 0 {
		args = append(args, fmt.Sprintf("interface=%d", rm.tunIfIndex))
	} else {
		args = append(args, fmt.Sprintf("interface=%s", tunDeviceName))
	}
	args = append(args, fmt.Sprintf("metric=%d", metric), "store=active")

	out, err := rm.cmdRunner(2*time.Second, "netsh", args...)
	if err != nil {
		return fmt.Errorf("failed to add IPv6 route %s on ifIndex %d: %v (%s)", prefix, rm.tunIfIndex, err, strings.TrimSpace(out))
	}

	rm.trackedRoutes = append(rm.trackedRoutes, RouteRecord{
		DestinationPrefix: prefix,
		InterfaceIndex:    rm.tunIfIndex,
		Metric:            metric,
		Family:            6,
	})
	return nil
}

// VerifyRouteExists confirms that the prefix actually exists on FyrnoxTun in the kernel table.
func (rm *RouteManager) VerifyRouteExists(prefix string, family int) (bool, error) {
	if family == 6 && rm.table2Reader != nil {
		rows, err := rm.table2Reader(windows.AF_INET6)
		if err == nil {
			var expPrefixLen uint8
			var expIP net.IP
			if _, ipNet, pErr := net.ParseCIDR(prefix); pErr == nil {
				expIP = ipNet.IP
				ones, _ := ipNet.Mask.Size()
				expPrefixLen = uint8(ones)
			} else {
				parts := strings.Split(prefix, "/")
				if len(parts) == 2 {
					expIP = net.ParseIP(parts[0])
					if l, lErr := strconv.Atoi(parts[1]); lErr == nil {
						expPrefixLen = uint8(l)
					}
				}
			}

			for _, row := range rows {
				if rm.tunIfIndex > 0 && row.InterfaceIndex == uint32(rm.tunIfIndex) {
					if row.PrefixLength == expPrefixLen {
						var rowIP net.IP
						if row.DestinationPrefixFamily == windows.AF_INET6 {
							rowIP = net.IP(row.DestinationPrefixData[6:22])
						}
						if expIP == nil || rowIP.Equal(expIP) {
							return true, nil
						}
					}
				}
			}
		}
	}

	if rm.cmdRunner == nil {
		rm.cmdRunner = runSilentCmdWithTimeout
	}

	if family == 4 {
		out, err := rm.cmdRunner(1500*time.Millisecond, "route", "print", prefix)
		if err != nil {
			return false, err
		}
		return strings.Contains(out, prefix) && (strings.Contains(out, "10.254.1.2") || (rm.tunIfIndex > 0 && strings.Contains(out, strconv.Itoa(rm.tunIfIndex)))), nil
	}

	out, err := rm.cmdRunner(1500*time.Millisecond, "netsh", "interface", "ipv6", "show", "route", "store=active")
	if err != nil {
		return false, err
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		for i, field := range fields {
			if field == prefix && i+1 < len(fields) {
				idxStr := fields[i+1]
				if rm.tunIfIndex > 0 && idxStr == strconv.Itoa(rm.tunIfIndex) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// VerifyBestRoute uses Windows route evaluation to confirm that targetIP forwards via FyrnoxTun.
func (rm *RouteManager) VerifyBestRoute(remoteIP string, expectedIfIndex int) (bool, error) {
	if rm.bestRoute2Reader != nil {
		ip := net.ParseIP(remoteIP)
		if ip != nil && ip.To4() == nil {
			idx, _, err := rm.bestRoute2Reader(ip)
			if err == nil {
				return idx == uint32(expectedIfIndex), nil
			}
		}
	}

	if rm.cmdRunner == nil {
		rm.cmdRunner = runSilentCmdWithTimeout
	}

	psCmd := fmt.Sprintf("(Find-NetRoute -RemoteIPAddress '%s').InterfaceIndex", remoteIP)
	out, err := rm.cmdRunner(2*time.Second, "powershell", "-NoProfile", "-Command", psCmd)
	if err != nil {
		return false, err
	}
	out = strings.TrimSpace(out)
	idx, perr := strconv.Atoi(out)
	if perr != nil {
		return false, fmt.Errorf("unparseable interface index from Find-NetRoute: %q", out)
	}
	return idx == expectedIfIndex, nil
}

// DetectMoreSpecificPhysicalRoutes inspects the kernel IPv6 forwarding table using IP Helper
// and returns all routes on non-Wintun physical adapters that have prefix length > 1 (e.g. on-link /64 prefixes).
func (rm *RouteManager) DetectMoreSpecificPhysicalRoutes() ([]string, error) {
	if rm.table2Reader == nil {
		return nil, nil
	}
	rows, err := rm.table2Reader(windows.AF_INET6)
	if err != nil {
		return nil, err
	}

	var specificRoutes []string
	for _, row := range rows {
		if rm.tunIfIndex > 0 && row.InterfaceIndex == uint32(rm.tunIfIndex) {
			continue
		}
		// Prefix length > 1 and < 128 (excluding default routes ::/0, split routes ::/1, and single host routes)
		if row.PrefixLength > 1 && row.PrefixLength < 128 {
			var ip net.IP
			if row.DestinationPrefixFamily == windows.AF_INET6 {
				ip = net.IP(row.DestinationPrefixData[6:22])
			}
			specificRoutes = append(specificRoutes, fmt.Sprintf("ifIndex=%d prefix=%s/%d metric=%d", row.InterfaceIndex, ip, row.PrefixLength, row.Metric))
		}
	}
	return specificRoutes, nil
}

// Rollback deletes only the exact routes tracked in the ledger, in reverse order.
func (rm *RouteManager) Rollback() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.cmdRunner == nil {
		rm.cmdRunner = runSilentCmdWithTimeout
	}

	for i := len(rm.trackedRoutes) - 1; i >= 0; i-- {
		r := rm.trackedRoutes[i]
		if r.Family == 4 {
			args := []string{"delete", r.DestinationPrefix, "mask", r.Mask}
			if r.NextHop != "" {
				args = append(args, r.NextHop)
			}
			if r.InterfaceIndex > 0 {
				args = append(args, "if", strconv.Itoa(r.InterfaceIndex))
			}
			_, _ = rm.cmdRunner(500*time.Millisecond, "route", args...)
		} else if r.Family == 6 {
			args := []string{"interface", "ipv6", "delete", "route", r.DestinationPrefix}
			if r.InterfaceIndex > 0 {
				args = append(args, fmt.Sprintf("interface=%d", r.InterfaceIndex))
			}
			args = append(args, "store=active")
			_, _ = rm.cmdRunner(500*time.Millisecond, "netsh", args...)
		}
	}

	rm.trackedRoutes = nil
	return nil
}

// TrackedCount returns the number of active routes owned by this manager.
func (rm *RouteManager) TrackedCount() int {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return len(rm.trackedRoutes)
}

// IsTracked reports whether a specific destination prefix is owned by this manager.
func (rm *RouteManager) IsTracked(prefix string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for _, r := range rm.trackedRoutes {
		if r.DestinationPrefix == prefix {
			return true
		}
	}
	return false
}

// RouteCleanupDeps holds injectable dependencies for ownership-verified orphaned route recovery.
type RouteCleanupDeps struct {
	Table2Reader  func(family uint16) ([]MibIpforwardRow2, error)
	LiveIfIndices func() (map[uint32]bool, error)
	FyrnoxIfIndex func() uint32
	IsTunActive   func() bool
	CmdRunner     func(timeout time.Duration, name string, args ...string) (string, error)
}

func defaultRouteCleanupDeps() RouteCleanupDeps {
	return RouteCleanupDeps{
		Table2Reader: readIpForwardTable2,
		LiveIfIndices: func() (map[uint32]bool, error) {
			ifaces, err := net.Interfaces()
			if err != nil {
				return nil, err
			}
			m := make(map[uint32]bool, len(ifaces))
			for _, iface := range ifaces {
				m[uint32(iface.Index)] = true
			}
			return m, nil
		},
		FyrnoxIfIndex: func() uint32 {
			iface, err := net.InterfaceByName(tunDeviceName)
			if err != nil {
				return 0
			}
			return uint32(iface.Index)
		},
		IsTunActive: func() bool {
			return IsTunModeActive() || IsWintunActive()
		},
		CmdRunner: runSilentCmdWithTimeout,
	}
}

// RouteCleanupResult summarizes orphaned route recovery without modifying unverified routes.
type RouteCleanupResult struct {
	Removed    int
	Unresolved []string
}

// CleanupOrphanedFyrnoxRoutes inspects IPv4 and IPv6 kernel forwarding tables and removes
// split routes only when independent FYRNOX ownership is established via:
// 1. Active FyrnoxTun interface index (fyrnoxIdx > 0 && InterfaceIndex == fyrnoxIdx), or
// 2. FYRNOX Wintun next-hop (10.254.1.2) for IPv4, which also correlates dead interface indices for IPv6.
// Unknown or dead interface indices without FYRNOX ownership evidence are left untouched and reported as Unresolved.
func CleanupOrphanedFyrnoxRoutes(deps RouteCleanupDeps) (RouteCleanupResult, error) {
	var res RouteCleanupResult
	if deps.IsTunActive != nil && deps.IsTunActive() {
		return res, nil
	}
	if deps.Table2Reader == nil || deps.CmdRunner == nil {
		return res, nil
	}

	var fyrnoxIdx uint32
	if deps.FyrnoxIfIndex != nil {
		fyrnoxIdx = deps.FyrnoxIfIndex()
	}
	var liveIfs map[uint32]bool
	if deps.LiveIfIndices != nil {
		liveIfs, _ = deps.LiveIfIndices()
	}

	ownedIfIndices := make(map[uint32]bool)
	if fyrnoxIdx > 0 {
		ownedIfIndices[fyrnoxIdx] = true
	}

	var errs []error

	// 1. Inspect IPv4 split routes (0.0.0.0/1 and 128.0.0.0/1)
	if v4Rows, err := deps.Table2Reader(windows.AF_INET); err == nil {
		fyrnoxGW := net.ParseIP(tunIPAddr).To4()
		for _, row := range v4Rows {
			if row.PrefixLength != 1 || row.DestinationPrefixFamily != windows.AF_INET {
				continue
			}
			destIP := net.IP(row.DestinationPrefixData[4:8]).To4()
			if destIP == nil || (!destIP.Equal(net.IPv4(0, 0, 0, 0)) && !destIP.Equal(net.IPv4(128, 0, 0, 0))) {
				continue
			}
			nextHop := net.IP(row.NextHopData[4:8]).To4()
			gwOwned := fyrnoxGW != nil && nextHop != nil && nextHop.Equal(fyrnoxGW)
			ifOwned := fyrnoxIdx > 0 && row.InterfaceIndex == fyrnoxIdx
			if row.InterfaceIndex == 0 || (!gwOwned && !ifOwned) {
				if row.InterfaceIndex == 0 || (liveIfs != nil && !liveIfs[row.InterfaceIndex]) {
					res.Unresolved = append(res.Unresolved, fmt.Sprintf("ipv4 %s/1 ifIndex=%d nextHop=%v", destIP.String(), row.InterfaceIndex, nextHop))
				}
				continue
			}

			ownedIfIndices[row.InterfaceIndex] = true
			args := []string{"delete", destIP.String(), "mask", "128.0.0.0"}
			if nextHop != nil && !nextHop.Equal(net.IPv4zero) {
				args = append(args, nextHop.String())
			}
			args = append(args, "if", strconv.FormatUint(uint64(row.InterfaceIndex), 10))
			if _, rErr := deps.CmdRunner(1000*time.Millisecond, "route", args...); rErr != nil {
				errs = append(errs, fmt.Errorf("delete orphaned IPv4 route %s/1: %w", destIP.String(), rErr))
			} else {
				res.Removed++
			}
		}
	}

	// 2. Inspect IPv6 split routes (::/1 and 8000::/1)
	if v6Rows, err := deps.Table2Reader(windows.AF_INET6); err == nil {
		halfV6 := net.ParseIP("8000::").To16()
		for _, row := range v6Rows {
			if row.PrefixLength != 1 || row.DestinationPrefixFamily != windows.AF_INET6 {
				continue
			}
			destIP := net.IP(row.DestinationPrefixData[6:22]).To16()
			if destIP == nil {
				continue
			}
			var prefixStr string
			if destIP.Equal(net.IPv6zero) {
				prefixStr = "::/1"
			} else if halfV6 != nil && destIP.Equal(halfV6) {
				prefixStr = "8000::/1"
			} else {
				continue
			}

			if row.InterfaceIndex == 0 || !ownedIfIndices[row.InterfaceIndex] {
				if row.InterfaceIndex == 0 || (liveIfs != nil && !liveIfs[row.InterfaceIndex]) {
					res.Unresolved = append(res.Unresolved, fmt.Sprintf("ipv6 %s ifIndex=%d", prefixStr, row.InterfaceIndex))
				}
				continue
			}

			args := []string{
				"interface", "ipv6", "delete", "route", prefixStr,
				fmt.Sprintf("interface=%d", row.InterfaceIndex),
				"store=active",
			}
			if _, rErr := deps.CmdRunner(1000*time.Millisecond, "netsh", args...); rErr != nil {
				errs = append(errs, fmt.Errorf("delete orphaned IPv6 route %s: %w", prefixStr, rErr))
			} else {
				res.Removed++
			}
		}
	}

	return res, errors.Join(errs...)
}

// RecoverOrphanedTunRoutes cleans up any orphaned FYRNOX split routes from crashed sessions.
func RecoverOrphanedTunRoutes() (RouteCleanupResult, error) {
	return CleanupOrphanedFyrnoxRoutes(defaultRouteCleanupDeps())
}
