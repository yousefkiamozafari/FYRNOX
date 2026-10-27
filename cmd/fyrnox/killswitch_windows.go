package main

import (
	"sync"
	"time"
)

type KillSwitchManager struct {
	mu     sync.Mutex
	active bool
}

var globalKillSwitch = &KillSwitchManager{}

// CleanupStaleKillSwitchRules removes any lingering FYRNOX KillSwitch firewall rules
func CleanupStaleKillSwitchRules() {
	rules := []string{
		"FYRNOX-KillSwitch-Block",
		"FYRNOX-KillSwitch-AllowLAN",
		"FYRNOX-KillSwitch-AllowDHCP",
		"FYRNOX-KillSwitch-AllowVPN",
	}
	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			_, _ = runSilentCmdWithTimeout(800*time.Millisecond, "netsh", "advfirewall", "firewall", "delete", "rule", "name="+r)
		}(rule)
	}
	wg.Wait()
}

// Arm arms the KillSwitch firewall filter, blocking non-VPN traffic while permitting LAN and VPN server IP
func (k *KillSwitchManager) Arm(serverIP string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.active {
		return nil
	}

	LogInfo("Arming FYRNOX Kill Switch (leak prevention)...")

	// Clean any previous instances
	CleanupStaleKillSwitchRules()

	// 1. Allow local LAN & loopback so LAN sharing, local gateways, and webview IPC work
	_, _ = runSilentCmdWithTimeout(1*time.Second, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=FYRNOX-KillSwitch-AllowLAN",
		"dir=out",
		"action=allow",
		"remoteip=127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12",
	)

	// 2. Allow DHCP so local network adapter doesn't drop IP lease
	_, _ = runSilentCmdWithTimeout(1*time.Second, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=FYRNOX-KillSwitch-AllowDHCP",
		"dir=out",
		"action=allow",
		"protocol=UDP",
		"remoteport=67,68",
	)

	// 3. Allow current VPN server IP so tunnel handshake, ping, and reconnect can succeed
	if serverIP != "" {
		_, _ = runSilentCmdWithTimeout(1*time.Second, "netsh", "advfirewall", "firewall", "add", "rule",
			"name=FYRNOX-KillSwitch-AllowVPN",
			"dir=out",
			"action=allow",
			"remoteip="+serverIP,
		)
	}

	// 4. Block all other outbound traffic (Windows Firewall evaluates ALLOW rules before BLOCK rules)
	_, _ = runSilentCmdWithTimeout(1*time.Second, "netsh", "advfirewall", "firewall", "add", "rule",
		"name=FYRNOX-KillSwitch-Block",
		"dir=out",
		"action=block",
	)

	k.active = true
	LogInfo("FYRNOX Kill Switch armed successfully")
	return nil
}

// Disarm releases all Kill Switch blocking rules
func (k *KillSwitchManager) Disarm() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.active {
		CleanupStaleKillSwitchRules()
		return nil
	}

	LogInfo("Disarming FYRNOX Kill Switch...")
	CleanupStaleKillSwitchRules()
	k.active = false
	LogInfo("FYRNOX Kill Switch disarmed successfully")
	return nil
}

// IsActive returns whether the KillSwitch is currently armed
func (k *KillSwitchManager) IsActive() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.active
}
