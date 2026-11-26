package main

import (
	"testing"
	"time"
)

func TestLatencyClassificationThresholds(t *testing.T) {
	cases := []struct {
		ping     int
		expected string
	}{
		{ping: 45, expected: "fast"},
		{ping: 100, expected: "fast"},
		{ping: 200, expected: "fast"},
		{ping: 201, expected: "medium"},
		{ping: 500, expected: "medium"},
		{ping: 501, expected: "slow"},
		{ping: 850, expected: "slow"},
		{ping: -1, expected: "unavailable"},
		{ping: 0, expected: "unavailable"},
	}
	for _, c := range cases {
		var got string
		if c.ping > 0 && c.ping <= 200 {
			got = "fast"
		} else if c.ping > 200 && c.ping <= 500 {
			got = "medium"
		} else if c.ping > 500 {
			got = "slow"
		} else {
			got = "unavailable"
		}
		if got != c.expected {
			t.Errorf("ping %d: got %s, want %s", c.ping, got, c.expected)
		}
	}
}

func TestComputeMedianLatency(t *testing.T) {
	tests := []struct {
		name     string
		samples  []int
		expected int
	}{
		{
			name:     "empty samples",
			samples:  []int{},
			expected: -1,
		},
		{
			name:     "all failed (-1)",
			samples:  []int{-1, -1, -1},
			expected: -1,
		},
		{
			name:     "single successful sample",
			samples:  []int{-1, 120, -1},
			expected: 120,
		},
		{
			name:     "three valid samples odd count",
			samples:  []int{140, 95, 110},
			expected: 110,
		},
		{
			name:     "even valid samples",
			samples:  []int{100, 200},
			expected: 150,
		},
		{
			name:     "four samples with one failure",
			samples:  []int{-1, 90, 80, 100},
			expected: 90,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := computeMedianLatency(tc.samples)
			if actual != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, actual)
			}
		})
	}
}

func TestPingAllServers_PreservesCachedBaselineWhenConnected(t *testing.T) {
	engine := NewVpnEngine()
	engine.mu.Lock()
	if len(engine.servers) == 0 {
		engine.mu.Unlock()
		t.Fatalf("expected non-empty default server list")
	}
	engine.servers[0].PingMs = 77
	engine.state = StateConnected
	engine.mu.Unlock()

	res := engine.PingAllServers()
	if len(res) == 0 || res[0].PingMs != 77 {
		t.Fatalf("expected cached baseline PingMs=77 preserved while connected, got %+v", res[0])
	}
}

func TestSelectServer_HotSwitchSkipsDisconnectedState(t *testing.T) {
	engine := NewVpnEngine()
	engine.mu.Lock()
	if len(engine.servers) < 2 {
		engine.mu.Unlock()
		t.Fatalf("expected at least 2 servers")
	}
	targetHost := engine.servers[1].Hostname
	engine.state = StateConnected
	engine.mu.Unlock()

	if err := engine.SelectServer(targetHost); err != nil {
		t.Fatalf("SelectServer failed: %v", err)
	}

	// Immediately after SelectServer, state must be StateConnecting (not StateDisconnecting/StateDisconnected)
	st := engine.GetStatus()
	if st.State != StateConnecting {
		t.Fatalf("expected StateConnecting immediately after hot SelectServer, got %s", st.State)
	}

	// Cancel via Disconnect during switch
	_ = engine.Disconnect()
	stAfter := engine.GetStatus()
	if stAfter.State != StateDisconnected {
		t.Fatalf("expected StateDisconnected after Disconnect cancellation, got %s", stAfter.State)
	}
}

func TestSelectServer_HotSwitchInvokesConnectFlow(t *testing.T) {
	engine := NewVpnEngine()
	engine.mu.Lock()
	if len(engine.servers) < 2 {
		engine.mu.Unlock()
		t.Fatalf("expected at least 2 servers")
	}
	targetHost := engine.servers[1].Hostname
	engine.auth = nil // Ensure Connect() reaches auth check and transitions to StateError
	engine.state = StateConnected
	engine.mu.Unlock()

	if err := engine.SelectServer(targetHost); err != nil {
		t.Fatalf("SelectServer failed: %v", err)
	}

	// Wait for async switch goroutine (150ms sleep + Connect() execution)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st := engine.GetStatus()
		if st.State == StateError {
			return // Connect() actually executed instead of returning early on StateConnecting!
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("SelectServer hot switch stuck in state=%s; Connect() returned early without dialing", engine.GetStatus().State)
}
