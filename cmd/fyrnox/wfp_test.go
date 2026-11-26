//go:build windows

package main

import (
	"errors"
	"net"
	"sync"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

type recordedFilter struct {
	Name       string
	LayerKey   windows.GUID
	Action     uint32
	Weight     uint64
	Conditions []FwpmFilterCondition0
}

type mockWfpEngine struct {
	mu           sync.Mutex
	openCalled   bool
	closeCalled  bool
	txBegun      bool
	txCommitted  bool
	txAborted    bool
	sublayers    []FwpmSubLayer0
	filters      []recordedFilter
	nextFilterID uint64
	failOnFilter int // index on which to fail (1-based)
}

func (m *mockWfpEngine) Open(serverName *uint16, authnService uint32, authIdentity uintptr, session *FwpmSession0) (uintptr, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.openCalled = true
	return 0xCAFE0001, nil
}

func (m *mockWfpEngine) Close(handle uintptr) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true
	return nil
}

func (m *mockWfpEngine) BeginTx(handle uintptr) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txBegun = true
	return nil
}

func (m *mockWfpEngine) CommitTx(handle uintptr) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txCommitted = true
	return nil
}

func (m *mockWfpEngine) AbortTx(handle uintptr) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txAborted = true
	return nil
}

func (m *mockWfpEngine) AddSubLayer(handle uintptr, subLayer *FwpmSubLayer0) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sublayers = append(m.sublayers, *subLayer)
	return nil
}

func (m *mockWfpEngine) AddFilter(handle uintptr, filter *FwpmFilter0) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failOnFilter > 0 && len(m.filters)+1 == m.failOnFilter {
		return 0, errors.New("mock engine simulated filter addition failure")
	}

	m.nextFilterID++
	name := ""
	if filter.DisplayData.Name != nil {
		name = windows.UTF16PtrToString(filter.DisplayData.Name)
	}

	weightVal := uint64(0)
	if filter.Weight.Type == fwpUint64 && filter.Weight.Value != nil {
		weightVal = *(*uint64)(filter.Weight.Value)
	}

	var conds []FwpmFilterCondition0
	if filter.NumFilterConditions > 0 && filter.FilterCondition != nil {
		condSlice := unsafe.Slice(filter.FilterCondition, filter.NumFilterConditions)
		conds = append(conds, condSlice...)
	}

	rec := recordedFilter{
		Name:       name,
		LayerKey:   filter.LayerKey,
		Action:     filter.Action.Type,
		Weight:     weightVal,
		Conditions: conds,
	}
	m.filters = append(m.filters, rec)
	return m.nextFilterID, nil
}

func TestWfp_LayerConditionValidation(t *testing.T) {
	fakeField := windows.GUID{Data1: 0xDEADBEEF}

	// 1. ALE_AUTH_CONNECT_V6: supports Flags, IF Index, Remote Addr, Remote Port, Protocol
	validALEConds := []windows.GUID{
		ConditionFlags,
		ConditionInterfaceIndex,
		ConditionIPRemoteAddress,
		ConditionIPRemotePort,
		ConditionIPProtocol,
	}
	for _, c := range validALEConds {
		if err := ValidateFilterCondition(LayerALEAuthConnectV6, c); err != nil {
			t.Errorf("Expected condition %v to be valid at ALE_AUTH_CONNECT_V6, got %v", c, err)
		}
	}
	if err := ValidateFilterCondition(LayerALEAuthConnectV6, fakeField); !errors.Is(err, ErrUnsupportedLayerCondition) {
		t.Errorf("Expected ErrUnsupportedLayerCondition for fakeField at ALE, got %v", err)
	}

	// 2. OUTBOUND_TRANSPORT_V6: supports Flags, IF Index, Remote Addr, Remote Port, Protocol
	validTransConds := []windows.GUID{
		ConditionFlags,
		ConditionInterfaceIndex,
		ConditionIPRemoteAddress,
		ConditionIPRemotePort,
		ConditionIPProtocol,
	}
	for _, c := range validTransConds {
		if err := ValidateFilterCondition(LayerOutboundTransportV6, c); err != nil {
			t.Errorf("Expected condition %v to be valid at OUTBOUND_TRANSPORT_V6, got %v", c, err)
		}
	}
	if err := ValidateFilterCondition(LayerOutboundTransportV6, fakeField); !errors.Is(err, ErrUnsupportedLayerCondition) {
		t.Errorf("Expected ErrUnsupportedLayerCondition for fakeField at Transport, got %v", err)
	}

	// 3. OUTBOUND_IPPACKET_V6: supports ONLY Flags, IF Index, Remote Addr. Rejects Port and Protocol!
	validNetConds := []windows.GUID{
		ConditionFlags,
		ConditionInterfaceIndex,
		ConditionIPRemoteAddress,
	}
	for _, c := range validNetConds {
		if err := ValidateFilterCondition(LayerOutboundIPPacketV6, c); err != nil {
			t.Errorf("Expected condition %v to be valid at OUTBOUND_IPPACKET_V6, got %v", c, err)
		}
	}

	// Specifically verify that IPPACKET rejects IP_PROTOCOL and IP_REMOTE_PORT
	if err := ValidateFilterCondition(LayerOutboundIPPacketV6, ConditionIPRemotePort); !errors.Is(err, ErrUnsupportedLayerCondition) {
		t.Errorf("Expected rejection of ConditionIPRemotePort at OUTBOUND_IPPACKET_V6, got %v", err)
	}
	if err := ValidateFilterCondition(LayerOutboundIPPacketV6, ConditionIPProtocol); !errors.Is(err, ErrUnsupportedLayerCondition) {
		t.Errorf("Expected rejection of ConditionIPProtocol at OUTBOUND_IPPACKET_V6, got %v", err)
	}
}

func TestWfpManager_RuleGeneration(t *testing.T) {
	mock := &mockWfpEngine{}
	wfp := &WfpManager{api: mock}

	serverIPv6 := net.ParseIP("2606:4700::1111")
	err := wfp.EnableIPv6LeakProtection(15, serverIPv6, 2499)
	if err != nil {
		t.Fatalf("EnableIPv6LeakProtection returned unexpected error: %v", err)
	}

	if !wfp.IsActive() {
		t.Fatalf("Expected WfpManager to be active")
	}
	if !mock.openCalled {
		t.Errorf("Expected engine Open to be called")
	}
	if !mock.txBegun {
		t.Errorf("Expected transaction to begin")
	}
	if !mock.txCommitted {
		t.Errorf("Expected transaction to be committed")
	}

	// 1. Verify Sublayer configuration
	if len(mock.sublayers) != 1 {
		t.Fatalf("Expected 1 sublayer, got %d", len(mock.sublayers))
	}
	if mock.sublayers[0].SubLayerKey != SublayerFyrnoxV6Leak {
		t.Errorf("Sublayer key mismatch")
	}
	if mock.sublayers[0].Weight != 0xF000 {
		t.Errorf("Expected sublayer weight 0xF000, got 0x%04x", mock.sublayers[0].Weight)
	}

	// 2. Verify Filter Rules count
	// Expected: 6 ALE + 6 Transport + 6 IPPacket = 18 filters total
	if len(mock.filters) != 18 {
		t.Fatalf("Expected 18 filters installed across 3 layers, got %d", len(mock.filters))
	}
	if wfp.FilterCount() != 18 {
		t.Errorf("FilterCount = %d, want 18", wfp.FilterCount())
	}

	// Verify all layers are populated
	layerCounts := make(map[windows.GUID]int)
	for _, f := range mock.filters {
		layerCounts[f.LayerKey]++
	}
	if layerCounts[LayerALEAuthConnectV6] != 6 {
		t.Errorf("ALE layer filters count = %d, want 6", layerCounts[LayerALEAuthConnectV6])
	}
	if layerCounts[LayerOutboundTransportV6] != 6 {
		t.Errorf("Transport layer filters count = %d, want 6", layerCounts[LayerOutboundTransportV6])
	}
	if layerCounts[LayerOutboundIPPacketV6] != 6 {
		t.Errorf("IPPacket layer filters count = %d, want 6", layerCounts[LayerOutboundIPPacketV6])
	}

	// 3. Verify Rule 1: Loopback (Weights 15, PERMIT on all 3 layers)
	for _, f := range mock.filters[:3] {
		if f.Action != fwpActionPermit || f.Weight != 15 {
			t.Errorf("Rule 1 attributes mismatch: %+v", f)
		}
		if len(f.Conditions) != 1 || f.Conditions[0].FieldKey != ConditionFlags {
			t.Errorf("Rule 1 condition mismatch: %+v", f.Conditions)
		}
	}

	// 4. Verify Rule 2: Wintun IF Index 15 (Weights 14, PERMIT on all 3 layers)
	for _, f := range mock.filters[3:6] {
		if f.Action != fwpActionPermit || f.Weight != 14 {
			t.Errorf("Rule 2 attributes mismatch: %+v", f)
		}
		if len(f.Conditions) != 1 || f.Conditions[0].FieldKey != ConditionInterfaceIndex {
			t.Errorf("Rule 2 condition field mismatch: %+v", f.Conditions)
		}
		if f.Conditions[0].MatchType != fwpMatchEqual {
			t.Errorf("Rule 2 match type want equal, got %d", f.Conditions[0].MatchType)
		}
		if val := *(*uint32)(f.Conditions[0].ConditionValue.Value); val != 15 {
			t.Errorf("Rule 2 interface index want 15, got %d", val)
		}
	}

	// 5. Verify Rule 3: ICMPv6 Link-Local & Multicast (fe80::/10 and ff02::/16)
	// Transport layer ICMPv6 (filters 6 and 7): IP_PROTOCOL == 58 AND remote in fe80::/10 or ff02::/16
	f3TransFe80 := mock.filters[6]
	if f3TransFe80.LayerKey != LayerOutboundTransportV6 || f3TransFe80.Action != fwpActionPermit || f3TransFe80.Weight != 13 {
		t.Errorf("Rule 3 Transport fe80 attributes mismatch: %+v", f3TransFe80)
	}
	if len(f3TransFe80.Conditions) != 2 {
		t.Fatalf("Rule 3 Transport fe80 conditions count want 2, got %d", len(f3TransFe80.Conditions))
	}
	proto := *(*uint8)(f3TransFe80.Conditions[0].ConditionValue.Value)
	if proto != ipProtoICMPv6 {
		t.Errorf("Rule 3 protocol want 58, got %d", proto)
	}
	fe80Mask := *(*FwpV6AddrAndMask)(f3TransFe80.Conditions[1].ConditionValue.Value)
	if fe80Mask.PrefixLength != 10 || fe80Mask.Addr[0] != 0xfe || fe80Mask.Addr[1] != 0x80 {
		t.Errorf("Rule 3 fe80 mask mismatch: %+v", fe80Mask)
	}

	// IPPacket layer link-local permit (filters 8 and 9): ONLY remote in fe80::/10 or ff02::/16
	f3NetFe80 := mock.filters[8]
	if f3NetFe80.LayerKey != LayerOutboundIPPacketV6 || f3NetFe80.Action != fwpActionPermit || f3NetFe80.Weight != 13 {
		t.Errorf("Rule 3 IPPacket fe80 attributes mismatch: %+v", f3NetFe80)
	}
	if len(f3NetFe80.Conditions) != 1 || f3NetFe80.Conditions[0].FieldKey != ConditionIPRemoteAddress {
		t.Errorf("Rule 3 IPPacket fe80 condition mismatch (must only match remote address): %+v", f3NetFe80.Conditions)
	}

	// 6. Verify Rule 4: Outer MASQUE IPv6 transport
	// ALE (filter 10): IP == serverIPv6, Port == 2499, Protocol == TCP
	fMasqueALE := mock.filters[10]
	if fMasqueALE.LayerKey != LayerALEAuthConnectV6 || fMasqueALE.Action != fwpActionPermit || fMasqueALE.Weight != 12 {
		t.Errorf("Rule 4 ALE attributes mismatch: %+v", fMasqueALE)
	}
	if len(fMasqueALE.Conditions) != 3 {
		t.Fatalf("Rule 4 ALE conditions count want 3, got %d", len(fMasqueALE.Conditions))
	}
	port := *(*uint16)(fMasqueALE.Conditions[1].ConditionValue.Value)
	if port != 2499 {
		t.Errorf("Rule 4 ALE port want 2499, got %d", port)
	}

	// Transport (filter 11): IP == serverIPv6, Port == 2499, Protocol == TCP
	fMasqueTrans := mock.filters[11]
	if fMasqueTrans.LayerKey != LayerOutboundTransportV6 || fMasqueTrans.Action != fwpActionPermit || fMasqueTrans.Weight != 12 {
		t.Errorf("Rule 4 Transport attributes mismatch: %+v", fMasqueTrans)
	}

	// IPPacket (filter 12): IP == serverIPv6 (NO port or protocol condition!)
	fMasqueNet := mock.filters[12]
	if fMasqueNet.LayerKey != LayerOutboundIPPacketV6 || fMasqueNet.Action != fwpActionPermit || fMasqueNet.Weight != 12 {
		t.Errorf("Rule 4 IPPacket attributes mismatch: %+v", fMasqueNet)
	}
	if len(fMasqueNet.Conditions) != 1 || fMasqueNet.Conditions[0].FieldKey != ConditionIPRemoteAddress {
		t.Errorf("Rule 4 IPPacket conditions count want 1 (RemoteAddress only), got %+v", fMasqueNet.Conditions)
	}

	// 7. Verify Rule 5: DNS Blocking on Non-Wintun (ports 53 and 853 on ALE)
	fDNS53 := mock.filters[13]
	if fDNS53.Action != fwpActionBlock || fDNS53.LayerKey != LayerALEAuthConnectV6 || fDNS53.Weight != 11 {
		t.Errorf("Rule 5 DNS 53 attributes mismatch: %+v", fDNS53)
	}
	dnsPort53 := *(*uint16)(fDNS53.Conditions[1].ConditionValue.Value)
	if dnsPort53 != 53 {
		t.Errorf("Rule 5 DNS port want 53, got %d", dnsPort53)
	}

	fDNS853 := mock.filters[14]
	dnsPort853 := *(*uint16)(fDNS853.Conditions[1].ConditionValue.Value)
	if dnsPort853 != 853 {
		t.Errorf("Rule 5 DNS port want 853, got %d", dnsPort853)
	}

	// 8. Verify Rule 6: Default-Deny on Non-Wintun (ALE, Transport, IPPacket)
	for _, f := range mock.filters[15:18] {
		if f.Action != fwpActionBlock || f.Weight != 10 {
			t.Errorf("Rule 6 Default-Deny attributes mismatch: %+v", f)
		}
		if len(f.Conditions) != 1 || f.Conditions[0].FieldKey != ConditionInterfaceIndex {
			t.Errorf("Rule 6 condition mismatch: %+v", f.Conditions)
		}
		if f.Conditions[0].MatchType != fwpMatchNotEqual {
			t.Errorf("Rule 6 match want not-equal, got %d", f.Conditions[0].MatchType)
		}
	}

	// 9. Verify Close purges engine
	if err := wfp.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if wfp.IsActive() {
		t.Errorf("Expected inactive after Close")
	}
	if !mock.closeCalled {
		t.Errorf("Expected engine Close to be invoked")
	}
}

func TestWfpManager_OuterIPv4_NoOuterIPv6Rule(t *testing.T) {
	mock := &mockWfpEngine{}
	wfp := &WfpManager{api: mock}

	// When MASQUE connects via IPv4, outerMasqueIPv6 is nil
	err := wfp.EnableIPv6LeakProtection(8, nil, 0)
	if err != nil {
		t.Fatalf("EnableIPv6LeakProtection error: %v", err)
	}

	// Filters: 18 - 3 (no outer MASQUE exception on ALE, Transport, IPPacket) = 15 filters
	if len(mock.filters) != 15 {
		t.Fatalf("Expected 15 filters when outer IPv4 used, got %d", len(mock.filters))
	}

	layerCounts := make(map[windows.GUID]int)
	for _, f := range mock.filters {
		layerCounts[f.LayerKey]++
		if f.Name == "FYRNOX-Permit-OuterMasque-ALE" ||
			f.Name == "FYRNOX-Permit-OuterMasque-Transport" ||
			f.Name == "FYRNOX-Permit-OuterMasque-Packet" {
			t.Errorf("Unexpected outer MASQUE rule present in IPv4 outer mode: %s", f.Name)
		}
	}
	if layerCounts[LayerALEAuthConnectV6] != 5 || layerCounts[LayerOutboundTransportV6] != 5 || layerCounts[LayerOutboundIPPacketV6] != 5 {
		t.Errorf("Expected 5/5/5 filters per layer in IPv4 outer mode, got ALE=%d Transport=%d IPPacket=%d",
			layerCounts[LayerALEAuthConnectV6], layerCounts[LayerOutboundTransportV6], layerCounts[LayerOutboundIPPacketV6])
	}
}

func TestWfpManager_TransactionRollbackOnError(t *testing.T) {
	mock := &mockWfpEngine{failOnFilter: 4}
	wfp := &WfpManager{api: mock}

	err := wfp.EnableIPv6LeakProtection(10, net.ParseIP("2606:4700::1111"), 2499)
	if err == nil {
		t.Fatalf("Expected error when filter addition fails")
	}

	if wfp.IsActive() {
		t.Errorf("Expected WfpManager to remain inactive after failure")
	}
	if !mock.txAborted {
		t.Errorf("Expected transaction Abort to be called on failure")
	}
	if !mock.closeCalled {
		t.Errorf("Expected engine Close to be called on failure")
	}
}

func TestWfpManager_InputValidation(t *testing.T) {
	mock := &mockWfpEngine{}
	wfp := &WfpManager{api: mock}

	// tunIfIndex 0 must be rejected
	err := wfp.EnableIPv6LeakProtection(0, nil, 0)
	if err == nil {
		t.Fatalf("Expected error for tunIfIndex = 0")
	}
	if mock.openCalled {
		t.Errorf("Engine should not be opened when validation fails")
	}
}

func TestWfpManager_CloseIdempotency(t *testing.T) {
	wfp := NewWfpManager()
	if err := wfp.Close(); err != nil {
		t.Fatalf("Initial Close on inactive manager failed: %v", err)
	}
	if err := wfp.Close(); err != nil {
		t.Fatalf("Second Close failed: %v", err)
	}
}
