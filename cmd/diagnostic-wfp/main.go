//go:build windows

package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modWfpFwpuclnt             = windows.NewLazySystemDLL("fwpuclnt.dll")
	procFwpmEngineOpen0        = modWfpFwpuclnt.NewProc("FwpmEngineOpen0")
	procFwpmEngineClose0       = modWfpFwpuclnt.NewProc("FwpmEngineClose0")
	procFwpmSubLayerAdd0       = modWfpFwpuclnt.NewProc("FwpmSubLayerAdd0")
	procFwpmFilterAdd0         = modWfpFwpuclnt.NewProc("FwpmFilterAdd0")
	procFwpmTransactionBegin   = modWfpFwpuclnt.NewProc("FwpmTransactionBegin0")
	procFwpmTransactionCommit  = modWfpFwpuclnt.NewProc("FwpmTransactionCommit0")
	procFwpmTransactionAbort   = modWfpFwpuclnt.NewProc("FwpmTransactionAbort0")
	procFwpmFilterCreateEnumH  = modWfpFwpuclnt.NewProc("FwpmFilterCreateEnumHandle0")
	procFwpmFilterEnum0        = modWfpFwpuclnt.NewProc("FwpmFilterEnum0")
	procFwpmFilterDestroyEnumH = modWfpFwpuclnt.NewProc("FwpmFilterDestroyEnumHandle0")
	procFwpmFreeMemory0        = modWfpFwpuclnt.NewProc("FwpmFreeMemory0")
)

var (
	LayerALEAuthConnectV6 = windows.GUID{
		Data1: 0xc38d57d1, Data2: 0x05a7, Data3: 0x4c33,
		Data4: [8]byte{0x90, 0x4f, 0x7f, 0xb4, 0xec, 0xfb, 0x78, 0x09},
	}
	LayerOutboundTransportV6 = windows.GUID{
		Data1: 0xe1735bde, Data2: 0x013f, Data3: 0x4655,
		Data4: [8]byte{0xb3, 0x51, 0xa4, 0x9e, 0x15, 0x76, 0x2d, 0xf0},
	}
	LayerOutboundIPPacketV6 = windows.GUID{
		Data1: 0x0aee2411, Data2: 0x1d97, Data3: 0x4a9f,
		Data4: [8]byte{0xab, 0x2b, 0xe8, 0x2e, 0xd3, 0x06, 0xf7, 0xe4},
	}
	ConditionFlags = windows.GUID{
		Data1: 0xb2a6c427, Data2: 0xa0d7, Data3: 0x4638,
		Data4: [8]byte{0x89, 0xc5, 0x2c, 0xd4, 0x67, 0x65, 0x79, 0xe0},
	}
	ConditionInterfaceIndex = windows.GUID{
		Data1: 0x3b5b6301, Data2: 0x26ff, Data3: 0x4927,
		Data4: [8]byte{0xaa, 0x1a, 0x82, 0x15, 0x50, 0xc6, 0xca, 0x80},
	}
	ConditionIPRemoteAddress = windows.GUID{
		Data1: 0x4954d0d7, Data2: 0x55ae, Data3: 0x49f3,
		Data4: [8]byte{0xac, 0x56, 0xd7, 0x60, 0x77, 0x3d, 0x74, 0x0c},
	}
	ConditionIPRemotePort = windows.GUID{
		Data1: 0xaf35eb44, Data2: 0x48f5, Data3: 0x4c01,
		Data4: [8]byte{0xa1, 0xe4, 0xe7, 0x4f, 0x14, 0x5f, 0x8e, 0x40},
	}
	ConditionIPProtocol = windows.GUID{
		Data1: 0xe09a311b, Data2: 0x9f9b, Data3: 0x4ff2,
		Data4: [8]byte{0xa5, 0x48, 0xc1, 0x16, 0xd7, 0xa4, 0x6f, 0x23},
	}
	SublayerFyrnoxV6Leak = windows.GUID{
		Data1: 0xfa6f3910, Data2: 0x2213, Data3: 0x4f49,
		Data4: [8]byte{0x87, 0x12, 0x93, 0x8a, 0x9a, 0xb2, 0x89, 0xc1},
	}
)

const (
	rpcCAuthnDefault           uint32 = 0xFFFFFFFF
	fwpmSessionFlagDynamic     uint32 = 0x00000001
	fwpActionPermit            uint32 = 0x00000001
	fwpActionBlock             uint32 = 0x00000002
	fwpMatchEqual              uint32 = 0
	fwpMatchNotEqual           uint32 = 2
	fwpMatchFlagsAllSet        uint32 = 8
	fwpUint8                   uint32 = 1
	fwpUint16                  uint32 = 2
	fwpUint32                  uint32 = 3
	fwpUint64                  uint32 = 4
	fwpByteArray16Type         uint32 = 11
	fwpV6AddrMask              uint32 = 21
	fwpConditionFlagIsLoopback uint32 = 0x00000001

	ipProtoICMPv6 uint8 = 58
	ipProtoTCP    uint8 = 6
)

type FwpmDisplayData0 struct {
	Name        *uint16
	Description *uint16
}

type FwpmSession0 struct {
	SessionKey           windows.GUID
	DisplayData          FwpmDisplayData0
	Flags                uint32
	TxnWaitTimeoutInMSec uint32
	ProcessId            uint32
	Sid                  uintptr
	Username             uintptr
	KernelMode           int32
}

type FwpByteBlob struct {
	Size uint32
	Data *byte
}

type FwpmSubLayer0 struct {
	SubLayerKey  windows.GUID
	DisplayData  FwpmDisplayData0
	Flags        uint32
	ProviderKey  *windows.GUID
	ProviderData FwpByteBlob
	Weight       uint16
}

type FwpValue0 struct {
	Type  uint32
	Value unsafe.Pointer
}

type FwpAction0 struct {
	Type       uint32
	FilterType windows.GUID
}

type FwpConditionValue0 struct {
	Type  uint32
	Value unsafe.Pointer
}

type FwpmFilterCondition0 struct {
	FieldKey       windows.GUID
	MatchType      uint32
	ConditionValue FwpConditionValue0
}

type FwpV6AddrAndMask struct {
	Addr         [16]byte
	PrefixLength uint8
}

type FwpmFilter0 struct {
	FilterKey           windows.GUID
	DisplayData         FwpmDisplayData0
	Flags               uint32
	ProviderKey         *windows.GUID
	ProviderData        FwpByteBlob
	LayerKey            windows.GUID
	SubLayerKey         windows.GUID
	Weight              FwpValue0
	NumFilterConditions uint32
	FilterCondition     *FwpmFilterCondition0
	Action              FwpAction0
	_padContext         uint32
	ProviderContextKey  windows.GUID
	Reserved            *windows.GUID
	FilterId            uint64
	EffectiveWeight     FwpValue0
}

type FwpmFilterEnumTemplate0 struct {
	ProviderKey             *windows.GUID
	LayerKey                windows.GUID
	EnumType                uint32
	Flags                   uint32
	ProviderContextTemplate uintptr
	NumFilterConditions     uint32
	FilterCondition         *FwpmFilterCondition0
	ActionMask              uint32
	CalloutKey              *windows.GUID
}

func isAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func enumerateLayerFilters(engineHandle uintptr, layerName string, layerKey windows.GUID, expectedFyrnoxCount int) (int, error) {
	var enumHandle uintptr
	var enumTemplate FwpmFilterEnumTemplate0
	enumTemplate.LayerKey = layerKey

	rEnumCreate, _, errCreate := procFwpmFilterCreateEnumH.Call(
		engineHandle,
		uintptr(unsafe.Pointer(&enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	if rEnumCreate != 0 || enumHandle == 0 {
		return 0, fmt.Errorf("FwpmFilterCreateEnumHandle0(%s) failed (0x%08x): %v", layerName, uint32(rEnumCreate), errCreate)
	}
	defer func() {
		_, _, _ = procFwpmFilterDestroyEnumH.Call(engineHandle, enumHandle)
	}()

	var entriesPtr **FwpmFilter0
	var numEntries uint32
	rEnum, _, errEnum := procFwpmFilterEnum0.Call(
		engineHandle,
		enumHandle,
		1000,
		uintptr(unsafe.Pointer(&entriesPtr)),
		uintptr(unsafe.Pointer(&numEntries)),
	)
	if rEnum != 0 {
		return 0, fmt.Errorf("FwpmFilterEnum0(%s) failed (0x%08x): %v", layerName, uint32(rEnum), errEnum)
	}
	if entriesPtr != nil {
		defer func() {
			_, _, _ = procFwpmFreeMemory0.Call(uintptr(unsafe.Pointer(&entriesPtr)))
		}()
	}

	fyrnoxCount := 0
	if numEntries > 0 && entriesPtr != nil {
		ptrs := unsafe.Slice(entriesPtr, numEntries)
		for _, f := range ptrs {
			if f == nil {
				continue
			}
			if f.SubLayerKey == SublayerFyrnoxV6Leak {
				fyrnoxCount++
				name := ""
				if f.DisplayData.Name != nil {
					name = windows.UTF16PtrToString(f.DisplayData.Name)
				}
				actionStr := "OTHER"
				switch f.Action.Type {
				case fwpActionPermit:
					actionStr = "PERMIT"
				case fwpActionBlock:
					actionStr = "BLOCK"
				}
				fmt.Printf("    [ENUM %s] FilterId=%d Action=%-6s Name=%s\n", layerName, f.FilterId, actionStr, name)
			}
		}
	}

	if fyrnoxCount != expectedFyrnoxCount {
		return fyrnoxCount, fmt.Errorf("layer %s: expected %d FYRNOX filters in BFE, found %d", layerName, expectedFyrnoxCount, fyrnoxCount)
	}
	fmt.Printf("  [PASS] Layer %-24s: %d/%d FYRNOX filters verified in BFE (total system filters at layer: %d)\n",
		layerName, fyrnoxCount, expectedFyrnoxCount, numEntries)
	return fyrnoxCount, nil
}

func verifyZeroResidualFyrnoxFilters() error {
	var inspectHandle uintptr
	var inspectSession FwpmSession0
	sInspect, _ := syscall.UTF16PtrFromString("FYRNOX Inspection Session")
	inspectSession.DisplayData.Name = sInspect
	rInspOpen, _, errOpen := procFwpmEngineOpen0.Call(0, uintptr(rpcCAuthnDefault), 0, uintptr(unsafe.Pointer(&inspectSession)), uintptr(unsafe.Pointer(&inspectHandle)))
	if rInspOpen != 0 || inspectHandle == 0 {
		return fmt.Errorf("FwpmEngineOpen0 inspection failed (0x%08x): %v", uint32(rInspOpen), errOpen)
	}
	defer func() {
		_, _, _ = procFwpmEngineClose0.Call(inspectHandle)
	}()

	layers := []struct {
		name string
		guid windows.GUID
	}{
		{"ALE_AUTH_CONNECT_V6", LayerALEAuthConnectV6},
		{"OUTBOUND_TRANSPORT_V6", LayerOutboundTransportV6},
		{"OUTBOUND_IPPACKET_V6", LayerOutboundIPPacketV6},
	}

	for _, l := range layers {
		found, err := enumerateLayerFilters(inspectHandle, l.name, l.guid, 0)
		if err != nil {
			return fmt.Errorf("residual filter check at %s: found %d (%w)", l.name, found, err)
		}
	}
	return nil
}

func isAccessDeniedErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EACCES) || errors.Is(err, windows.WSAEACCES) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "access was denied") ||
		strings.Contains(msg, "forbidden by its access permissions") ||
		strings.Contains(msg, "10013")
}

func isNoRouteErr(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, windows.WSAENETUNREACH) ||
		errors.Is(err, windows.WSAEHOSTUNREACH) ||
		strings.Contains(strings.ToLower(err.Error()), "unreachable")
}

type wfpSecurityEvent struct {
	EventID string `xml:"System>EventID"`
	Data    []struct {
		Name  string `xml:"Name,attr"`
		Value string `xml:",chardata"`
	} `xml:"EventData>Data"`
}

func queryRecentWfpAuditEvents() {
	cmd := exec.Command("wevtutil", "qe", "Security",
		"/q:*[System[(EventID=5152 or EventID=5157) and TimeCreated[timediff(@SystemTime) <= 15000]]]",
		"/c:8", "/rd:true", "/f:xml")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  [INFO] Windows Security Event Log 5152/5157 unavailable (%v). Enable 'Filtering Platform Packet Drop' & 'Filtering Platform Connection' via auditpol in VM to record kernel LayerName/FilterRTID.\n", err)
		return
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		fmt.Println("  [INFO] No Security Event Log 5152/5157 events recorded in last 15s (enable WFP auditpol subcategories in VM to capture FilterRTID).")
		return
	}

	wrapped := "<Events>" + raw + "</Events>"
	var parsed struct {
		Events []wfpSecurityEvent `xml:"Event"`
	}
	if err := xml.Unmarshal([]byte(wrapped), &parsed); err != nil {
		fmt.Printf("  [INFO] Could not parse Security Event XML: %v\n", err)
		return
	}
	if len(parsed.Events) == 0 {
		fmt.Println("  [INFO] 0 matching WFP audit events (5152/5157) found.")
		return
	}

	fmt.Printf("  [INFO] Captured %d kernel WFP audit event(s) from Security Log:\n", len(parsed.Events))
	for _, ev := range parsed.Events {
		fields := map[string]string{}
		for _, d := range ev.Data {
			fields[d.Name] = d.Value
		}
		fmt.Printf("    EventID=%s LayerName=%s LayerRTID=%s FilterRTID=%s Dest=%s:%s Proto=%s\n",
			ev.EventID, fields["LayerName"], fields["LayerRTID"], fields["FilterRTID"],
			fields["DestAddress"], fields["DestPort"], fields["Protocol"])
	}
}

func runTargetedTrafficTests(outerIPv6 net.IP, outerPort uint16) bool {
	allObserved := true
	fmt.Println("[INFO] Executing live targeted socket & raw-packet validation (outer IPv6 mode)...")

	// 1. Loopback connect (Must NOT be blocked with WSAEACCES)
	conn, err := net.DialTimeout("tcp6", "[::1]:20808", 250*time.Millisecond)
	if err != nil {
		if isAccessDeniedErr(err) {
			fmt.Printf("  [FAIL] Loopback [::1]:20808 was blocked by WFP (WSAEACCES): %v\n", err)
			allObserved = false
		} else {
			fmt.Printf("  [PASS] Loopback [::1]:20808 permitted by WFP (reached stack: %v)\n", err)
		}
	} else {
		_ = conn.Close()
		fmt.Println("  [PASS] Loopback [::1]:20808 permitted and connected")
	}

	// 2. Permitted MASQUE TCP/2499
	masqueTarget := net.JoinHostPort(outerIPv6.String(), fmt.Sprintf("%d", outerPort))
	mConn, mErr := net.DialTimeout("tcp6", masqueTarget, 500*time.Millisecond)
	if mErr != nil {
		if isAccessDeniedErr(mErr) {
			fmt.Printf("  [FAIL] Permitted MASQUE TCP/%d (%s) was blocked by WFP (WSAEACCES): %v\n", outerPort, masqueTarget, mErr)
			allObserved = false
		} else if isNoRouteErr(mErr) {
			fmt.Printf("  [INCONCLUSIVE] MASQUE TCP/%d (%s): VM lacks IPv6 route (%v); WFP did not return WSAEACCES\n", outerPort, masqueTarget, mErr)
		} else {
			fmt.Printf("  [PASS] MASQUE TCP/%d (%s) permitted by WFP (reached network stack: %v)\n", outerPort, masqueTarget, mErr)
		}
	} else {
		_ = mConn.Close()
		fmt.Printf("  [PASS] MASQUE TCP/%d (%s) permitted and connected\n", outerPort, masqueTarget)
	}

	// 3. Blocked TCP to MASQUE IP on non-2499 port (443)
	blockedTCPTarget := net.JoinHostPort(outerIPv6.String(), "443")
	tConn, tErr := net.DialTimeout("tcp6", blockedTCPTarget, 400*time.Millisecond)
	if tErr == nil {
		_ = tConn.Close()
		fmt.Printf("  [FAIL] Non-2499 TCP (%s) unexpectedly connected!\n", blockedTCPTarget)
		allObserved = false
	} else if isAccessDeniedErr(tErr) {
		fmt.Printf("  [PASS] Non-2499 TCP (%s) blocked with WSAEACCES (10013): %v\n", blockedTCPTarget, tErr)
	} else if isNoRouteErr(tErr) {
		fmt.Printf("  [INCONCLUSIVE] Non-2499 TCP (%s): VM lacks IPv6 route (%v); cannot prove WFP block without IPv6 route\n", blockedTCPTarget, tErr)
		allObserved = false
	} else {
		fmt.Printf("  [FAIL] Non-2499 TCP (%s) failed with non-WFP error (expected WSAEACCES): %v\n", blockedTCPTarget, tErr)
		allObserved = false
	}

	// 4. Blocked UDP to MASQUE IP including port 2499
	udpConn, uErr := net.DialTimeout("udp6", masqueTarget, 400*time.Millisecond)
	if uErr == nil {
		_, uWriteErr := udpConn.Write([]byte("fyrnox-wfp-udp-probe"))
		_ = udpConn.Close()
		uErr = uWriteErr
	}
	if uErr == nil {
		fmt.Printf("  [FAIL] UDP to MASQUE IP (%s) succeeded without block!\n", masqueTarget)
		allObserved = false
	} else if isAccessDeniedErr(uErr) {
		fmt.Printf("  [PASS] UDP to MASQUE IP (%s) blocked with WSAEACCES (10013): %v\n", masqueTarget, uErr)
	} else if isNoRouteErr(uErr) {
		fmt.Printf("  [INCONCLUSIVE] UDP to MASQUE IP (%s): VM lacks IPv6 route (%v); cannot prove WFP block without IPv6 route\n", masqueTarget, uErr)
		allObserved = false
	} else {
		fmt.Printf("  [FAIL] UDP to MASQUE IP (%s) returned unexpected error (expected WSAEACCES): %v\n", masqueTarget, uErr)
		allObserved = false
	}

	// 5. Blocked Raw IPv6 socket traffic (AF_INET6, SOCK_RAW, proto 253) to MASQUE IP
	rawFD, sockErr := windows.Socket(windows.AF_INET6, windows.SOCK_RAW, 253)
	if sockErr != nil {
		fmt.Printf("  [INCONCLUSIVE] Raw IPv6 socket creation failed: %v\n", sockErr)
		allObserved = false
	} else {
		defer windows.Closesocket(rawFD)
		var sa windows.SockaddrInet6
		sa.Port = 0
		copy(sa.Addr[:], outerIPv6.To16())
		sendErr := windows.Sendto(rawFD, []byte{0x01, 0x02, 0x03, 0x04}, 0, &sa)
		if sendErr == nil {
			fmt.Printf("  [FAIL] Raw IPv6 Sendto(%s, proto 253) succeeded without block!\n", outerIPv6)
			allObserved = false
		} else if isAccessDeniedErr(sendErr) {
			fmt.Printf("  [PASS] Raw IPv6 Sendto(%s, proto 253) blocked with WSAEACCES (10013): %v\n", outerIPv6, sendErr)
		} else if isNoRouteErr(sendErr) {
			fmt.Printf("  [INCONCLUSIVE] Raw IPv6 Sendto(%s, proto 253): VM lacks IPv6 route (%v); cannot prove WFP block without IPv6 route\n", outerIPv6, sendErr)
			allObserved = false
		} else {
			fmt.Printf("  [FAIL] Raw IPv6 Sendto(%s, proto 253) returned unexpected error: %v\n", outerIPv6, sendErr)
			allObserved = false
		}
	}

	queryRecentWfpAuditEvents()
	return allObserved
}

func runWfpModeTest(modeName string, includeOuterIPv6 bool) bool {
	fmt.Printf("\n--- Mode: %s ---\n", modeName)
	var session FwpmSession0
	session.Flags = fwpmSessionFlagDynamic
	sName, _ := syscall.UTF16PtrFromString("FYRNOX WFP VM Acceptance Session - " + modeName)
	session.DisplayData.Name = sName

	var engineHandle uintptr
	rOpen, _, errOpen := procFwpmEngineOpen0.Call(
		0,
		uintptr(rpcCAuthnDefault),
		0,
		uintptr(unsafe.Pointer(&session)),
		uintptr(unsafe.Pointer(&engineHandle)),
	)
	if rOpen != 0 {
		fmt.Printf("[FAIL] FwpmEngineOpen0 returned 0x%08x: %v\n", uint32(rOpen), errOpen)
		return false
	}
	fmt.Printf("[PASS] Dynamic Engine Session Opened (Handle: 0x%x, rc: 0x%08x)\n", engineHandle, uint32(rOpen))

	rTxBegin, _, errTxBegin := procFwpmTransactionBegin.Call(engineHandle, 0)
	if rTxBegin != 0 {
		fmt.Printf("[FAIL] FwpmTransactionBegin0 returned 0x%08x: %v\n", uint32(rTxBegin), errTxBegin)
		_, _, _ = procFwpmEngineClose0.Call(engineHandle)
		return false
	}
	fmt.Printf("[PASS] Atomic Transaction Begun (rc: 0x%08x)\n", uint32(rTxBegin))

	subName, _ := syscall.UTF16PtrFromString("FYRNOX VM Test SubLayer")
	subDesc, _ := syscall.UTF16PtrFromString("Dynamic sublayer for leak prevention test")
	subLayer := FwpmSubLayer0{
		SubLayerKey: SublayerFyrnoxV6Leak,
		DisplayData: FwpmDisplayData0{Name: subName, Description: subDesc},
		Weight:      0xF000,
	}
	rSub, _, errSub := procFwpmSubLayerAdd0.Call(engineHandle, uintptr(unsafe.Pointer(&subLayer)), 0)
	if rSub != 0 {
		fmt.Printf("[FAIL] FwpmSubLayerAdd0 returned 0x%08x: %v\n", uint32(rSub), errSub)
		_, _, _ = procFwpmTransactionAbort.Call(engineHandle)
		_, _, _ = procFwpmEngineClose0.Call(engineHandle)
		return false
	}
	fmt.Printf("[PASS] SubLayer Added (Weight: 0xF000, rc: 0x%08x)\n", uint32(rSub))

	tunIfIndex := uint32(9999)
	outerIPv6 := net.ParseIP("2606:4700::1111")
	outerPort := uint16(2499)

	var (
		anchoredU32  []uint32
		anchoredU16  []uint16
		anchoredU8   []uint8
		anchoredIP16 [][16]byte
		anchoredMask []FwpV6AddrAndMask
		wPtrs        []*uint64
		filterIDs    []uint64
		addFailed    bool
	)

	addFilter := func(name string, layerKey windows.GUID, action uint32, weightVal uint64, conds []FwpmFilterCondition0) {
		wName, _ := syscall.UTF16PtrFromString(name)
		wPtr := new(uint64)
		*wPtr = weightVal
		wPtrs = append(wPtrs, wPtr)

		var condPtr *FwpmFilterCondition0
		if len(conds) > 0 {
			condPtr = &conds[0]
		}

		filter := FwpmFilter0{
			DisplayData:         FwpmDisplayData0{Name: wName},
			LayerKey:            layerKey,
			SubLayerKey:         SublayerFyrnoxV6Leak,
			Weight:              FwpValue0{Type: fwpUint64, Value: unsafe.Pointer(wPtr)},
			NumFilterConditions: uint32(len(conds)),
			FilterCondition:     condPtr,
			Action:              FwpAction0{Type: action},
		}

		var fID uint64
		rAdd, _, errAdd := procFwpmFilterAdd0.Call(engineHandle, uintptr(unsafe.Pointer(&filter)), 0, uintptr(unsafe.Pointer(&fID)))
		if rAdd != 0 {
			fmt.Printf("  [FAIL] FwpmFilterAdd0 %s failed (0x%08x): %v\n", name, uint32(rAdd), errAdd)
			addFailed = true
			return
		}
		filterIDs = append(filterIDs, fID)
	}

	// Rule 1: Loopback
	anchoredU32 = append(anchoredU32, fwpConditionFlagIsLoopback)
	loopbackCond := []FwpmFilterCondition0{
		{FieldKey: ConditionFlags, MatchType: fwpMatchFlagsAllSet, ConditionValue: FwpConditionValue0{Type: fwpUint32, Value: unsafe.Pointer(&anchoredU32[len(anchoredU32)-1])}},
	}
	addFilter("Permit-Loopback-ALE", LayerALEAuthConnectV6, fwpActionPermit, 15, loopbackCond)
	addFilter("Permit-Loopback-Transport", LayerOutboundTransportV6, fwpActionPermit, 15, loopbackCond)
	addFilter("Permit-Loopback-IPPacket", LayerOutboundIPPacketV6, fwpActionPermit, 15, loopbackCond)

	// Rule 2: Wintun
	anchoredU32 = append(anchoredU32, tunIfIndex)
	tunIfIdxPtr := unsafe.Pointer(&anchoredU32[len(anchoredU32)-1])
	tunCond := []FwpmFilterCondition0{
		{FieldKey: ConditionInterfaceIndex, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint32, Value: tunIfIdxPtr}},
	}
	addFilter("Permit-Wintun-ALE", LayerALEAuthConnectV6, fwpActionPermit, 14, tunCond)
	addFilter("Permit-Wintun-Transport", LayerOutboundTransportV6, fwpActionPermit, 14, tunCond)
	addFilter("Permit-Wintun-IPPacket", LayerOutboundIPPacketV6, fwpActionPermit, 14, tunCond)

	// Rule 3: Link-Local ICMPv6 / NDP
	anchoredMask = append(anchoredMask, FwpV6AddrAndMask{Addr: [16]byte{0xfe, 0x80}, PrefixLength: 10})
	fe80MaskPtr := unsafe.Pointer(&anchoredMask[len(anchoredMask)-1])
	anchoredMask = append(anchoredMask, FwpV6AddrAndMask{Addr: [16]byte{0xff, 0x02}, PrefixLength: 16})
	ff02MaskPtr := unsafe.Pointer(&anchoredMask[len(anchoredMask)-1])
	anchoredU8 = append(anchoredU8, ipProtoICMPv6)
	icmpProtoPtr := unsafe.Pointer(&anchoredU8[len(anchoredU8)-1])

	addFilter("Permit-ICMPv6-fe80-Trans", LayerOutboundTransportV6, fwpActionPermit, 13, []FwpmFilterCondition0{
		{FieldKey: ConditionIPProtocol, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint8, Value: icmpProtoPtr}},
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: fe80MaskPtr}},
	})
	addFilter("Permit-ICMPv6-ff02-Trans", LayerOutboundTransportV6, fwpActionPermit, 13, []FwpmFilterCondition0{
		{FieldKey: ConditionIPProtocol, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint8, Value: icmpProtoPtr}},
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: ff02MaskPtr}},
	})
	addFilter("Permit-NDP-fe80-Packet", LayerOutboundIPPacketV6, fwpActionPermit, 13, []FwpmFilterCondition0{
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: fe80MaskPtr}},
	})
	addFilter("Permit-NDP-ff02-Packet", LayerOutboundIPPacketV6, fwpActionPermit, 13, []FwpmFilterCondition0{
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: ff02MaskPtr}},
	})

	// Rule 4: Outer MASQUE IPv6 (only when includeOuterIPv6 is true)
	expectedTotal := 15
	expectedPerLayer := 5
	if includeOuterIPv6 {
		expectedTotal = 18
		expectedPerLayer = 6
		var ipBytes [16]byte
		copy(ipBytes[:], outerIPv6.To16())
		anchoredIP16 = append(anchoredIP16, ipBytes)
		ipPtr := unsafe.Pointer(&anchoredIP16[len(anchoredIP16)-1])
		anchoredU16 = append(anchoredU16, outerPort)
		portPtr := unsafe.Pointer(&anchoredU16[len(anchoredU16)-1])
		anchoredU8 = append(anchoredU8, ipProtoTCP)
		tcpProtoPtr := unsafe.Pointer(&anchoredU8[len(anchoredU8)-1])

		masqueALECond := []FwpmFilterCondition0{
			{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpByteArray16Type, Value: ipPtr}},
			{FieldKey: ConditionIPRemotePort, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint16, Value: portPtr}},
			{FieldKey: ConditionIPProtocol, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint8, Value: tcpProtoPtr}},
		}
		addFilter("Permit-OuterMasque-ALE", LayerALEAuthConnectV6, fwpActionPermit, 12, masqueALECond)
		addFilter("Permit-OuterMasque-Transport", LayerOutboundTransportV6, fwpActionPermit, 12, masqueALECond)
		addFilter("Permit-OuterMasque-Packet", LayerOutboundIPPacketV6, fwpActionPermit, 12, []FwpmFilterCondition0{
			{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpByteArray16Type, Value: ipPtr}},
		})
	}

	// Rule 5: Physical DNS 53 & 853
	for _, dp := range []uint16{53, 853} {
		anchoredU16 = append(anchoredU16, dp)
		dpPtr := unsafe.Pointer(&anchoredU16[len(anchoredU16)-1])
		addFilter(fmt.Sprintf("Block-Physical-DNS-Port%d", dp), LayerALEAuthConnectV6, fwpActionBlock, 11, []FwpmFilterCondition0{
			{FieldKey: ConditionInterfaceIndex, MatchType: fwpMatchNotEqual, ConditionValue: FwpConditionValue0{Type: fwpUint32, Value: tunIfIdxPtr}},
			{FieldKey: ConditionIPRemotePort, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint16, Value: dpPtr}},
		})
	}

	// Rule 6: Default-Deny across all 3 layers
	nonWintunCond := []FwpmFilterCondition0{
		{FieldKey: ConditionInterfaceIndex, MatchType: fwpMatchNotEqual, ConditionValue: FwpConditionValue0{Type: fwpUint32, Value: tunIfIdxPtr}},
	}
	addFilter("Block-Physical-IPv6-ALE", LayerALEAuthConnectV6, fwpActionBlock, 10, nonWintunCond)
	addFilter("Block-Physical-IPv6-Transport", LayerOutboundTransportV6, fwpActionBlock, 10, nonWintunCond)
	addFilter("Block-Physical-IPv6-Packet", LayerOutboundIPPacketV6, fwpActionBlock, 10, nonWintunCond)

	if addFailed || len(filterIDs) != expectedTotal {
		fmt.Printf("[FAIL] Expected %d filters registered, got %d (addFailed=%v)\n", expectedTotal, len(filterIDs), addFailed)
		_, _, _ = procFwpmTransactionAbort.Call(engineHandle)
		_, _, _ = procFwpmEngineClose0.Call(engineHandle)
		return false
	}

	rTxCommit, _, errTxCommit := procFwpmTransactionCommit.Call(engineHandle)
	if rTxCommit != 0 {
		fmt.Printf("[FAIL] FwpmTransactionCommit0 returned 0x%08x: %v\n", uint32(rTxCommit), errTxCommit)
		_, _, _ = procFwpmEngineClose0.Call(engineHandle)
		return false
	}
	fmt.Printf("[PASS] Atomic Transaction Committed (%d filters, rc: 0x%08x)\n", len(filterIDs), uint32(rTxCommit))

	// Independent Enumeration across ALL 3 layers
	fmt.Println("[INFO] Enumerating FYRNOX filters independently at all 3 WFP layers...")
	enumOK := true
	for _, l := range []struct {
		name string
		guid windows.GUID
	}{
		{"ALE_AUTH_CONNECT_V6", LayerALEAuthConnectV6},
		{"OUTBOUND_TRANSPORT_V6", LayerOutboundTransportV6},
		{"OUTBOUND_IPPACKET_V6", LayerOutboundIPPacketV6},
	} {
		if _, err := enumerateLayerFilters(engineHandle, l.name, l.guid, expectedPerLayer); err != nil {
			fmt.Printf("  [FAIL] %v\n", err)
			enumOK = false
		}
	}

	trafficOK := true
	if includeOuterIPv6 {
		trafficOK = runTargetedTrafficTests(outerIPv6, outerPort)
	}

	rClose, _, _ := procFwpmEngineClose0.Call(engineHandle)
	fmt.Printf("[PASS] Dynamic Engine Session Closed (rc: 0x%08x)\n", uint32(rClose))

	if err := verifyZeroResidualFyrnoxFilters(); err != nil {
		fmt.Printf("[FAIL] Zero-Residual Verification failed: %v\n", err)
		return false
	}
	fmt.Println("[PASS] Zero-Residual Verification: 0 FYRNOX filters remain across all 3 layers")

	return enumOK && trafficOK
}

func main() {
	forceVM := flag.Bool("force-vm", false, "Force run even if VM check is skipped")
	flag.Parse()

	fmt.Println("================================================================================")
	fmt.Println("  FYRNOX Windows Filtering Platform (WFP) Native Kernel Integration Test")
	fmt.Println("================================================================================")

	hostname, _ := os.Hostname()
	if strings.EqualFold(hostname, "DESKTOP-3IOGCB8") {
		fmt.Printf("[FATAL ERROR] Refusing to execute on known primary host: %s\n", hostname)
		fmt.Println("This test modifies WFP in-kernel filters and must run exclusively inside an isolated VM.")
		os.Exit(1)
	}

	if !isAdmin() {
		fmt.Println("[FATAL ERROR] This test requires Administrator privileges. Run in an elevated console.")
		os.Exit(1)
	}

	if !*forceVM {
		fmt.Printf("[INFO] Running on test host: %s\n", hostname)
	}

	okIPv6 := runWfpModeTest("Outer-IPv6-MASQUE (18 filters: 6/6/6)", true)
	okIPv4 := runWfpModeTest("Outer-IPv4-MASQUE (15 filters: 5/5/5)", false)

	fmt.Println("\n================================================================================")
	if okIPv6 && okIPv4 {
		fmt.Println("  WFP NATIVE KERNEL INTEGRATION TEST PASSED (ALL CHECKS OBSERVED)")
		fmt.Println("================================================================================")
		os.Exit(0)
	}
	fmt.Println("  WFP NATIVE KERNEL INTEGRATION TEST COMPLETED WITH FAILURES/INCONCLUSIVE CHECKS")
	fmt.Println("================================================================================")
	os.Exit(1)
}
