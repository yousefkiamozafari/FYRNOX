//go:build windows

package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modWfpFwpuclnt            = windows.NewLazySystemDLL("fwpuclnt.dll")
	procFwpmEngineOpen0       = modWfpFwpuclnt.NewProc("FwpmEngineOpen0")
	procFwpmEngineClose0      = modWfpFwpuclnt.NewProc("FwpmEngineClose0")
	procFwpmSubLayerAdd0      = modWfpFwpuclnt.NewProc("FwpmSubLayerAdd0")
	procFwpmFilterAdd0        = modWfpFwpuclnt.NewProc("FwpmFilterAdd0")
	procFwpmTransactionBegin  = modWfpFwpuclnt.NewProc("FwpmTransactionBegin0")
	procFwpmTransactionCommit = modWfpFwpuclnt.NewProc("FwpmTransactionCommit0")
	procFwpmTransactionAbort  = modWfpFwpuclnt.NewProc("FwpmTransactionAbort0")
)

// WFP Filtering Layer GUIDs (Authoritative Microsoft Windows SDK Definitions)
var (
	// FWPM_LAYER_ALE_AUTH_CONNECT_V6: {c38d57d1-05a7-4c33-904f-7fb4ecfb7809}
	// Evaluates outbound socket connection requests prior to transport binding.
	LayerALEAuthConnectV6 = windows.GUID{
		Data1: 0xc38d57d1,
		Data2: 0x05a7,
		Data3: 0x4c33,
		Data4: [8]byte{0x90, 0x4f, 0x7f, 0xb4, 0xec, 0xfb, 0x78, 0x09},
	}

	// FWPM_LAYER_OUTBOUND_TRANSPORT_V6: {e1735bde-013f-4655-b351-a49e15762df0}
	// Evaluates outgoing transport-layer packets (TCP, UDP, ICMPv6) after headers are constructed.
	LayerOutboundTransportV6 = windows.GUID{
		Data1: 0xe1735bde,
		Data2: 0x013f,
		Data3: 0x4655,
		Data4: [8]byte{0xb3, 0x51, 0xa4, 0x9e, 0x15, 0x76, 0x2d, 0xf0},
	}

	// FWPM_LAYER_OUTBOUND_IPPACKET_V6: {0aee2411-1d97-4a9f-ab2b-e82ed306f7e4}
	// Evaluates raw IP-level packets in the send path just before packet fragmentation and NDIS handoff.
	LayerOutboundIPPacketV6 = windows.GUID{
		Data1: 0x0aee2411,
		Data2: 0x1d97,
		Data3: 0x4a9f,
		Data4: [8]byte{0xab, 0x2b, 0xe8, 0x2e, 0xd3, 0x06, 0xf7, 0xe4},
	}
)

// WFP Filtering Condition Field GUIDs
var (
	// FWPM_CONDITION_FLAGS: {b2a6c427-a0d7-4638-89c5-2cd4676579e0}
	ConditionFlags = windows.GUID{
		Data1: 0xb2a6c427,
		Data2: 0xa0d7,
		Data3: 0x4638,
		Data4: [8]byte{0x89, 0xc5, 0x2c, 0xd4, 0x67, 0x65, 0x79, 0xe0},
	}

	// FWPM_CONDITION_INTERFACE_INDEX: {3b5b6301-26ff-4927-aa1a-821550c6ca80}
	ConditionInterfaceIndex = windows.GUID{
		Data1: 0x3b5b6301,
		Data2: 0x26ff,
		Data3: 0x4927,
		Data4: [8]byte{0xaa, 0x1a, 0x82, 0x15, 0x50, 0xc6, 0xca, 0x80},
	}

	// FWPM_CONDITION_IP_REMOTE_ADDRESS: {4954d0d7-55ae-49f3-ac56-d760773d740c}
	ConditionIPRemoteAddress = windows.GUID{
		Data1: 0x4954d0d7,
		Data2: 0x55ae,
		Data3: 0x49f3,
		Data4: [8]byte{0xac, 0x56, 0xd7, 0x60, 0x77, 0x3d, 0x74, 0x0c},
	}

	// FWPM_CONDITION_IP_REMOTE_PORT: {af35eb44-48f5-4c01-a1e4-e74f145f8e40}
	ConditionIPRemotePort = windows.GUID{
		Data1: 0xaf35eb44,
		Data2: 0x48f5,
		Data3: 0x4c01,
		Data4: [8]byte{0xa1, 0xe4, 0xe7, 0x4f, 0x14, 0x5f, 0x8e, 0x40},
	}

	// FWPM_CONDITION_IP_PROTOCOL: {e09a311b-9f9b-4ff2-a548-c116d7a46f23}
	ConditionIPProtocol = windows.GUID{
		Data1: 0xe09a311b,
		Data2: 0x9f9b,
		Data3: 0x4ff2,
		Data4: [8]byte{0xa5, 0x48, 0xc1, 0x16, 0xd7, 0xa4, 0x6f, 0x23},
	}

	// FYRNOX Custom Sublayer GUID: {fa6f3910-2213-4f49-8712-938a9ab289c1}
	SublayerFyrnoxV6Leak = windows.GUID{
		Data1: 0xfa6f3910,
		Data2: 0x2213,
		Data3: 0x4f49,
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
	fwpEmpty                   uint32 = 0
	fwpUint8                   uint32 = 1
	fwpUint16                  uint32 = 2
	fwpUint32                  uint32 = 3
	fwpUint64                  uint32 = 4
	fwpByteArray16Type         uint32 = 11
	fwpV6AddrMask              uint32 = 21
	fwpConditionFlagIsLoopback uint32 = 0x00000001

	ipProtoICMPv6 uint8 = 58
	ipProtoTCP    uint8 = 6
	ipProtoUDP    uint8 = 17

	fyrnoxSublayerWeight uint16 = 0xF000
)

var (
	ErrUnsupportedLayerCondition = errors.New("filtering condition field is not supported at the specified WFP layer")
)

// ValidateFilterCondition verifies that fieldKey is documented by Microsoft as supported at layerKey.
// This prevents submitting incompatible layer-condition combinations to the Base Filtering Engine (BFE).
func ValidateFilterCondition(layerKey windows.GUID, fieldKey windows.GUID) error {
	switch layerKey {
	case LayerALEAuthConnectV6:
		switch fieldKey {
		case ConditionFlags, ConditionInterfaceIndex, ConditionIPRemoteAddress,
			ConditionIPRemotePort, ConditionIPProtocol:
			return nil
		default:
			return fmt.Errorf("%w: field %v at ALE_AUTH_CONNECT_V6", ErrUnsupportedLayerCondition, fieldKey)
		}
	case LayerOutboundTransportV6:
		switch fieldKey {
		case ConditionFlags, ConditionInterfaceIndex, ConditionIPRemoteAddress,
			ConditionIPRemotePort, ConditionIPProtocol:
			return nil
		default:
			return fmt.Errorf("%w: field %v at OUTBOUND_TRANSPORT_V6", ErrUnsupportedLayerCondition, fieldKey)
		}
	case LayerOutboundIPPacketV6:
		switch fieldKey {
		case ConditionFlags, ConditionInterfaceIndex, ConditionIPRemoteAddress:
			return nil
		default:
			return fmt.Errorf("%w: field %v at OUTBOUND_IPPACKET_V6", ErrUnsupportedLayerCondition, fieldKey)
		}
	default:
		return fmt.Errorf("%w: unknown layer %v", ErrUnsupportedLayerCondition, layerKey)
	}
}

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

type wfpEngineAPI interface {
	Open(serverName *uint16, authnService uint32, authIdentity uintptr, session *FwpmSession0) (uintptr, error)
	Close(handle uintptr) error
	BeginTx(handle uintptr) error
	CommitTx(handle uintptr) error
	AbortTx(handle uintptr) error
	AddSubLayer(handle uintptr, subLayer *FwpmSubLayer0) error
	AddFilter(handle uintptr, filter *FwpmFilter0) (uint64, error)
}

type realWfpEngine struct{}

func (r *realWfpEngine) Open(serverName *uint16, authnService uint32, authIdentity uintptr, session *FwpmSession0) (uintptr, error) {
	var handle uintptr
	r1, _, err := procFwpmEngineOpen0.Call(
		uintptr(unsafe.Pointer(serverName)),
		uintptr(authnService),
		authIdentity,
		uintptr(unsafe.Pointer(session)),
		uintptr(unsafe.Pointer(&handle)),
	)
	if r1 != 0 {
		return 0, fmt.Errorf("FwpmEngineOpen0 failed (0x%08x): %w", uint32(r1), err)
	}
	return handle, nil
}

func (r *realWfpEngine) Close(handle uintptr) error {
	if handle == 0 {
		return nil
	}
	r1, _, err := procFwpmEngineClose0.Call(handle)
	if r1 != 0 {
		return fmt.Errorf("FwpmEngineClose0 failed (0x%08x): %w", uint32(r1), err)
	}
	return nil
}

func (r *realWfpEngine) BeginTx(handle uintptr) error {
	r1, _, err := procFwpmTransactionBegin.Call(handle, 0)
	if r1 != 0 {
		return fmt.Errorf("FwpmTransactionBegin0 failed (0x%08x): %w", uint32(r1), err)
	}
	return nil
}

func (r *realWfpEngine) CommitTx(handle uintptr) error {
	r1, _, err := procFwpmTransactionCommit.Call(handle)
	if r1 != 0 {
		return fmt.Errorf("FwpmTransactionCommit0 failed (0x%08x): %w", uint32(r1), err)
	}
	return nil
}

func (r *realWfpEngine) AbortTx(handle uintptr) error {
	r1, _, err := procFwpmTransactionAbort.Call(handle)
	if r1 != 0 {
		return fmt.Errorf("FwpmTransactionAbort0 failed (0x%08x): %w", uint32(r1), err)
	}
	return nil
}

func (r *realWfpEngine) AddSubLayer(handle uintptr, subLayer *FwpmSubLayer0) error {
	r1, _, err := procFwpmSubLayerAdd0.Call(handle, uintptr(unsafe.Pointer(subLayer)), 0)
	if r1 != 0 {
		return fmt.Errorf("FwpmSubLayerAdd0 failed (0x%08x): %w", uint32(r1), err)
	}
	return nil
}

func (r *realWfpEngine) AddFilter(handle uintptr, filter *FwpmFilter0) (uint64, error) {
	var filterID uint64
	r1, _, err := procFwpmFilterAdd0.Call(handle, uintptr(unsafe.Pointer(filter)), 0, uintptr(unsafe.Pointer(&filterID)))
	if r1 != 0 {
		return 0, fmt.Errorf("FwpmFilterAdd0 failed (0x%08x): %w", uint32(r1), err)
	}
	return filterID, nil
}

// WfpManager manages the Windows Filtering Platform session and leak prevention filters.
//
// DYNAMIC SESSION CRASH POLICY:
// The WFP engine session is created using FWPM_SESSION_FLAG_DYNAMIC. Under Microsoft WFP
// semantics, all sublayers and filters created within a dynamic session are bound to the
// caller's RPC process handle in the Windows Base Filtering Engine (BFE).
// If fyrnox.exe terminates normally, is forcefully killed (taskkill /F), crashes, or encounters
// an unhandled exception, BFE automatically and atomically removes all session-owned filters.
//
// DESIGN INTENT & LIMITATION:
// FYRNOX uses dynamic sessions to guarantee fail-safe restoration of host networking upon
// process termination, preventing user internet connectivity from remaining permanently broken.
// A dynamic session alone DOES NOT provide a crash-persistent hard Kill Switch after the client
// process terminates. If a threat model requires persistent blocking after process death, an
// independent Windows Service or persistent filters (flags = 0) with an external recovery agent
// are required, which constitutes a distinct architecture subject to separate review.
type WfpManager struct {
	mu           sync.Mutex
	engineHandle uintptr
	active       bool
	filterIDs    []uint64
	api          wfpEngineAPI
}

// NewWfpManager creates a new instance of WfpManager
func NewWfpManager() *WfpManager {
	return &WfpManager{
		api: &realWfpEngine{},
	}
}

// buildFilterRule constructs an FwpmFilter0 structure after validating that every condition
// is documented as supported at the target layerKey.
func buildFilterRule(name, desc string, layerKey windows.GUID, sublayerKey windows.GUID, actionType uint32, weightVal uint64, conditions []FwpmFilterCondition0) (*FwpmFilter0, *uint64, error) {
	for _, c := range conditions {
		if err := ValidateFilterCondition(layerKey, c.FieldKey); err != nil {
			return nil, nil, fmt.Errorf("rule %q: %w", name, err)
		}
	}

	wName, _ := syscall.UTF16PtrFromString(name)
	wDesc, _ := syscall.UTF16PtrFromString(desc)

	wPtr := new(uint64)
	*wPtr = weightVal

	var condPtr *FwpmFilterCondition0
	if len(conditions) > 0 {
		condPtr = &conditions[0]
	}

	filter := &FwpmFilter0{
		DisplayData: FwpmDisplayData0{
			Name:        wName,
			Description: wDesc,
		},
		LayerKey:            layerKey,
		SubLayerKey:         sublayerKey,
		Weight:              FwpValue0{Type: fwpUint64, Value: unsafe.Pointer(wPtr)},
		NumFilterConditions: uint32(len(conditions)),
		FilterCondition:     condPtr,
		Action:              FwpAction0{Type: actionType},
	}
	return filter, wPtr, nil
}

// EnableIPv6LeakProtection configures kernel-level WFP outbound filtering across ALE, Transport,
// and IP Packet layers to unconditionally prevent IPv6 egress and physical DNS queries on non-Wintun interfaces.
func (m *WfpManager) EnableIPv6LeakProtection(tunIfIndex uint32, outerMasqueIPv6 net.IP, outerMasquePort uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.active {
		return nil
	}
	if tunIfIndex == 0 {
		return errors.New("invalid Wintun interface index (must be > 0)")
	}

	// 1. Open dynamic session
	var session FwpmSession0
	session.Flags = fwpmSessionFlagDynamic
	sName, _ := syscall.UTF16PtrFromString("FYRNOX IPv6 Leak Prevention Session")
	session.DisplayData.Name = sName

	handle, err := m.api.Open(nil, rpcCAuthnDefault, 0, &session)
	if err != nil {
		return fmt.Errorf("opening WFP dynamic session: %w", err)
	}
	m.engineHandle = handle

	// 2. Start atomic transaction
	if err := m.api.BeginTx(handle); err != nil {
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("beginning WFP transaction: %w", err)
	}

	// 3. Add custom sublayer with high priority
	subName, _ := syscall.UTF16PtrFromString("FYRNOX Leak Prevention SubLayer")
	subDesc, _ := syscall.UTF16PtrFromString("Enforces strict IPv6 containment to Wintun interface")
	subLayer := FwpmSubLayer0{
		SubLayerKey: SublayerFyrnoxV6Leak,
		DisplayData: FwpmDisplayData0{
			Name:        subName,
			Description: subDesc,
		},
		Weight: fyrnoxSublayerWeight,
	}
	if err := m.api.AddSubLayer(handle, &subLayer); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding WFP sublayer: %w", err)
	}

	// Heap allocations to anchor filter condition values during syscall execution
	var (
		anchoredU32     []uint32
		anchoredU16     []uint16
		anchoredU8      []uint8
		anchoredIP16    [][16]byte
		anchoredMask    []FwpV6AddrAndMask
		anchoredWeights []*uint64
		installedIDs    []uint64
	)

	addFilterWithRollback := func(filter *FwpmFilter0, wPtr *uint64) error {
		anchoredWeights = append(anchoredWeights, wPtr)
		id, err := m.api.AddFilter(handle, filter)
		if err != nil {
			return err
		}
		installedIDs = append(installedIDs, id)
		return nil
	}

	// -------------------------------------------------------------------------
	// RULE 1: Allow Loopback IPv6 traffic on ALE, Transport, and IP Packet layers
	// -------------------------------------------------------------------------
	loopbackVal := fwpConditionFlagIsLoopback
	anchoredU32 = append(anchoredU32, loopbackVal)
	loopbackCond := []FwpmFilterCondition0{
		{
			FieldKey:  ConditionFlags,
			MatchType: fwpMatchFlagsAllSet,
			ConditionValue: FwpConditionValue0{
				Type:  fwpUint32,
				Value: unsafe.Pointer(&anchoredU32[len(anchoredU32)-1]),
			},
		},
	}
	f1ALE, w1ALE, err := buildFilterRule("FYRNOX-Permit-Loopback-ALE", "Allow loopback IPv6 connections", LayerALEAuthConnectV6, SublayerFyrnoxV6Leak, fwpActionPermit, 15, loopbackCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f1ALE, w1ALE); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Loopback ALE rule: %w", err)
	}

	f1Trans, w1Trans, err := buildFilterRule("FYRNOX-Permit-Loopback-Transport", "Allow loopback IPv6 transport", LayerOutboundTransportV6, SublayerFyrnoxV6Leak, fwpActionPermit, 15, loopbackCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f1Trans, w1Trans); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Loopback Transport rule: %w", err)
	}

	f1Net, w1Net, err := buildFilterRule("FYRNOX-Permit-Loopback-Packet", "Allow loopback IPv6 packets", LayerOutboundIPPacketV6, SublayerFyrnoxV6Leak, fwpActionPermit, 15, loopbackCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f1Net, w1Net); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Loopback Packet rule: %w", err)
	}

	// -------------------------------------------------------------------------
	// RULE 2: Allow all IPv6 traffic traversing the Wintun interface
	// -------------------------------------------------------------------------
	anchoredU32 = append(anchoredU32, tunIfIndex)
	tunIfIdxPtr := unsafe.Pointer(&anchoredU32[len(anchoredU32)-1])
	wintunCond := []FwpmFilterCondition0{
		{
			FieldKey:  ConditionInterfaceIndex,
			MatchType: fwpMatchEqual,
			ConditionValue: FwpConditionValue0{
				Type:  fwpUint32,
				Value: tunIfIdxPtr,
			},
		},
	}
	f2ALE, w2ALE, err := buildFilterRule("FYRNOX-Permit-Wintun-ALE", "Allow all IPv6 connections via Wintun", LayerALEAuthConnectV6, SublayerFyrnoxV6Leak, fwpActionPermit, 14, wintunCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f2ALE, w2ALE); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Wintun ALE rule: %w", err)
	}

	f2Trans, w2Trans, err := buildFilterRule("FYRNOX-Permit-Wintun-Transport", "Allow all IPv6 transport via Wintun", LayerOutboundTransportV6, SublayerFyrnoxV6Leak, fwpActionPermit, 14, wintunCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f2Trans, w2Trans); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Wintun Transport rule: %w", err)
	}

	f2Net, w2Net, err := buildFilterRule("FYRNOX-Permit-Wintun-Packet", "Allow all IPv6 packets via Wintun", LayerOutboundIPPacketV6, SublayerFyrnoxV6Leak, fwpActionPermit, 14, wintunCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f2Net, w2Net); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Wintun Packet rule: %w", err)
	}

	// -------------------------------------------------------------------------
	// RULE 3: Allow essential Link-Local ICMPv6 (Neighbor Discovery: NS, NA, RS)
	//
	// Exact Traffic Permitted & Security Implications:
	// - NDP operates exclusively in link-local scopes: unicast fe80::/10 and multicast ff02::/16.
	// - At LayerOutboundTransportV6, we permit IP_PROTOCOL == 58 (ICMPv6) bound to fe80::/10 or ff02::/16.
	// - At LayerOutboundIPPacketV6, IP_PROTOCOL is not supported; we permit raw packets addressed to
	//   fe80::/10 or ff02::/16.
	// - Note: Because ICMP_TYPE is unavailable at these layers, all ICMPv6 types within fe80::/10
	//   and ff02::/16 are permitted. This is LAN-scoped traffic (non-routable across gateways), not
	//   a guarantee of zero local-network exposure. Global unicast IPv6 (2000::/3) remains blocked.
	// - Inbound Router Advertisements (Type 134) arrive inbound from the router and are not impacted by
	//   outbound filtering.
	// -------------------------------------------------------------------------
	anchoredMask = append(anchoredMask, FwpV6AddrAndMask{
		Addr:         [16]byte{0xfe, 0x80},
		PrefixLength: 10,
	})
	fe80MaskPtr := unsafe.Pointer(&anchoredMask[len(anchoredMask)-1])

	anchoredMask = append(anchoredMask, FwpV6AddrAndMask{
		Addr:         [16]byte{0xff, 0x02},
		PrefixLength: 16,
	})
	ff02MaskPtr := unsafe.Pointer(&anchoredMask[len(anchoredMask)-1])

	anchoredU8 = append(anchoredU8, ipProtoICMPv6)
	icmpProtoPtr := unsafe.Pointer(&anchoredU8[len(anchoredU8)-1])

	// Transport layer ICMPv6 permit rules
	icmpFe80TransCond := []FwpmFilterCondition0{
		{FieldKey: ConditionIPProtocol, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint8, Value: icmpProtoPtr}},
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: fe80MaskPtr}},
	}
	f3TransFe80, w3TransFe80, err := buildFilterRule("FYRNOX-Permit-ICMPv6-LinkLocal-Trans", "Permit link-local ICMPv6 unicast transport", LayerOutboundTransportV6, SublayerFyrnoxV6Leak, fwpActionPermit, 13, icmpFe80TransCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f3TransFe80, w3TransFe80); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding ICMPv6 LinkLocal Transport rule: %w", err)
	}

	icmpFf02TransCond := []FwpmFilterCondition0{
		{FieldKey: ConditionIPProtocol, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint8, Value: icmpProtoPtr}},
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: ff02MaskPtr}},
	}
	f3TransFf02, w3TransFf02, err := buildFilterRule("FYRNOX-Permit-ICMPv6-Multicast-Trans", "Permit link-local ICMPv6 multicast transport", LayerOutboundTransportV6, SublayerFyrnoxV6Leak, fwpActionPermit, 13, icmpFf02TransCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f3TransFf02, w3TransFf02); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding ICMPv6 Multicast Transport rule: %w", err)
	}

	// IPPacket layer link-local permit rules (preserves NDP packet transmission)
	packetFe80Cond := []FwpmFilterCondition0{
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: fe80MaskPtr}},
	}
	f3NetFe80, w3NetFe80, err := buildFilterRule("FYRNOX-Permit-NDP-LinkLocal-Packet", "Permit link-local NDP raw packets", LayerOutboundIPPacketV6, SublayerFyrnoxV6Leak, fwpActionPermit, 13, packetFe80Cond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f3NetFe80, w3NetFe80); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding NDP LinkLocal Packet rule: %w", err)
	}

	packetFf02Cond := []FwpmFilterCondition0{
		{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpV6AddrMask, Value: ff02MaskPtr}},
	}
	f3NetFf02, w3NetFf02, err := buildFilterRule("FYRNOX-Permit-NDP-Multicast-Packet", "Permit link-local NDP multicast packets", LayerOutboundIPPacketV6, SublayerFyrnoxV6Leak, fwpActionPermit, 13, packetFf02Cond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f3NetFf02, w3NetFf02); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding NDP Multicast Packet rule: %w", err)
	}

	// -------------------------------------------------------------------------
	// RULE 4: Outer MASQUE transport exception (Destination IP, Port 2499, TCP)
	//
	// Scope & Architectural Limitation:
	// - ALE Layer: Enforces exact destination IP, destination port 2499, and protocol TCP.
	// - Transport Layer: Enforces exact destination IP, destination port 2499, and protocol TCP.
	// - IPPacket Layer: Matches destination IP (port/protocol not supported by BFE at IPPacket).
	// - Limitation: This exception is destination-and-port scoped, NOT application-exclusive or
	//   4-tuple-exclusive. Restricting by local ephemeral port would break session rebuilds in
	//   proxyController.RebuildSession, and FWPM_CONDITION_ALE_APP_ID is unsupported at
	//   OUTBOUND_TRANSPORT_V6 and OUTBOUND_IPPACKET_V6.
	// - Multi-Layer Defense: Under WFP arbitration, a later-layer permit at IPPacket CANNOT undo
	//   an earlier terminating block at ALE or Transport. Any non-TCP or non-2499 traffic to
	//   outerMasqueIPv6 is blocked upstream at ALE / Transport.
	// - If MASQUE connects via IPv4, zero outer IPv6 exception is created.
	// -------------------------------------------------------------------------
	if outerMasqueIPv6 != nil && outerMasqueIPv6.To4() == nil && outerMasquePort > 0 {
		var ipBytes [16]byte
		copy(ipBytes[:], outerMasqueIPv6.To16())
		anchoredIP16 = append(anchoredIP16, ipBytes)
		ipPtr := unsafe.Pointer(&anchoredIP16[len(anchoredIP16)-1])

		anchoredU16 = append(anchoredU16, outerMasquePort)
		portPtr := unsafe.Pointer(&anchoredU16[len(anchoredU16)-1])

		anchoredU8 = append(anchoredU8, ipProtoTCP)
		tcpProtoPtr := unsafe.Pointer(&anchoredU8[len(anchoredU8)-1])

		masqueALECond := []FwpmFilterCondition0{
			{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpByteArray16Type, Value: ipPtr}},
			{FieldKey: ConditionIPRemotePort, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint16, Value: portPtr}},
			{FieldKey: ConditionIPProtocol, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint8, Value: tcpProtoPtr}},
		}
		f4ALE, w4ALE, err := buildFilterRule("FYRNOX-Permit-OuterMasque-ALE", "Permit outer MASQUE IPv6 socket", LayerALEAuthConnectV6, SublayerFyrnoxV6Leak, fwpActionPermit, 12, masqueALECond)
		if err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return err
		}
		if err := addFilterWithRollback(f4ALE, w4ALE); err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return fmt.Errorf("adding Outer MASQUE ALE rule: %w", err)
		}

		f4Trans, w4Trans, err := buildFilterRule("FYRNOX-Permit-OuterMasque-Transport", "Permit outer MASQUE IPv6 transport", LayerOutboundTransportV6, SublayerFyrnoxV6Leak, fwpActionPermit, 12, masqueALECond)
		if err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return err
		}
		if err := addFilterWithRollback(f4Trans, w4Trans); err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return fmt.Errorf("adding Outer MASQUE Transport rule: %w", err)
		}

		masqueNetCond := []FwpmFilterCondition0{
			{FieldKey: ConditionIPRemoteAddress, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpByteArray16Type, Value: ipPtr}},
		}
		f4Net, w4Net, err := buildFilterRule("FYRNOX-Permit-OuterMasque-Packet", "Permit outer MASQUE IPv6 packets", LayerOutboundIPPacketV6, SublayerFyrnoxV6Leak, fwpActionPermit, 12, masqueNetCond)
		if err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return err
		}
		if err := addFilterWithRollback(f4Net, w4Net); err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return fmt.Errorf("adding Outer MASQUE Packet rule: %w", err)
		}
	}

	// -------------------------------------------------------------------------
	// RULE 5: Block physical IPv6 DNS queries (ports 53 and 853) on non-Wintun interfaces
	// -------------------------------------------------------------------------
	dnsPorts := []uint16{53, 853}
	for _, dPort := range dnsPorts {
		anchoredU16 = append(anchoredU16, dPort)
		dPortPtr := unsafe.Pointer(&anchoredU16[len(anchoredU16)-1])
		dnsCond := []FwpmFilterCondition0{
			{FieldKey: ConditionInterfaceIndex, MatchType: fwpMatchNotEqual, ConditionValue: FwpConditionValue0{Type: fwpUint32, Value: tunIfIdxPtr}},
			{FieldKey: ConditionIPRemotePort, MatchType: fwpMatchEqual, ConditionValue: FwpConditionValue0{Type: fwpUint16, Value: dPortPtr}},
		}
		ruleName := fmt.Sprintf("FYRNOX-Block-Physical-DNS-Port%d", dPort)
		f5, w5, err := buildFilterRule(ruleName, "Block unencrypted IPv6 DNS queries on physical interfaces", LayerALEAuthConnectV6, SublayerFyrnoxV6Leak, fwpActionBlock, 11, dnsCond)
		if err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return err
		}
		if err := addFilterWithRollback(f5, w5); err != nil {
			_ = m.api.AbortTx(handle)
			_ = m.api.Close(handle)
			m.engineHandle = 0
			return fmt.Errorf("adding DNS block port %d rule: %w", dPort, err)
		}
	}

	// -------------------------------------------------------------------------
	// RULE 6: Default-Deny all remaining IPv6 traffic on non-Wintun physical interfaces
	// -------------------------------------------------------------------------
	nonWintunCond := []FwpmFilterCondition0{
		{FieldKey: ConditionInterfaceIndex, MatchType: fwpMatchNotEqual, ConditionValue: FwpConditionValue0{Type: fwpUint32, Value: tunIfIdxPtr}},
	}
	f6ALE, w6ALE, err := buildFilterRule("FYRNOX-Block-Physical-IPv6-ALE", "Drop outbound IPv6 connections on physical adapters", LayerALEAuthConnectV6, SublayerFyrnoxV6Leak, fwpActionBlock, 10, nonWintunCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f6ALE, w6ALE); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Default-Deny ALE rule: %w", err)
	}

	f6Trans, w6Trans, err := buildFilterRule("FYRNOX-Block-Physical-IPv6-Transport", "Drop outbound IPv6 transport on physical adapters", LayerOutboundTransportV6, SublayerFyrnoxV6Leak, fwpActionBlock, 10, nonWintunCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f6Trans, w6Trans); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Default-Deny Transport rule: %w", err)
	}

	f6Net, w6Net, err := buildFilterRule("FYRNOX-Block-Physical-IPv6-Packet", "Drop outbound IPv6 raw packets on physical adapters", LayerOutboundIPPacketV6, SublayerFyrnoxV6Leak, fwpActionBlock, 10, nonWintunCond)
	if err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return err
	}
	if err := addFilterWithRollback(f6Net, w6Net); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("adding Default-Deny Packet rule: %w", err)
	}

	// 4. Commit atomic transaction
	if err := m.api.CommitTx(handle); err != nil {
		_ = m.api.AbortTx(handle)
		_ = m.api.Close(handle)
		m.engineHandle = 0
		return fmt.Errorf("committing WFP transaction: %w", err)
	}

	m.active = true
	m.filterIDs = installedIDs
	LogInfo("WFP IPv6 leak prevention installed successfully (%d dynamic filters registered in kernel)", len(installedIDs))
	return nil
}

// Close gracefully closes the WFP engine session, purging all dynamic filters atomically
func (m *WfpManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.engineHandle == 0 {
		m.active = false
		m.filterIDs = nil
		return nil
	}

	err := m.api.Close(m.engineHandle)
	m.engineHandle = 0
	m.active = false
	m.filterIDs = nil
	if err != nil {
		return fmt.Errorf("closing WFP engine: %w", err)
	}
	LogInfo("WFP IPv6 leak prevention dynamic session closed and kernel filters purged")
	return nil
}

// IsActive returns whether WFP leak prevention filters are currently active
func (m *WfpManager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// FilterCount returns the number of active filters installed in the dynamic session
func (m *WfpManager) FilterCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.filterIDs)
}
