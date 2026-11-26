package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type killSwitchCmdRunner func(timeout time.Duration, name string, args ...string) (string, error)

type KillSwitchManager struct {
	mu        sync.Mutex
	active    bool
	cmdRunner killSwitchCmdRunner
	isAdminFn func() bool
}

var globalKillSwitch = &KillSwitchManager{}

var killSwitchRuleNames = []string{
	"FYRNOX-KillSwitch-AllowLAN",
	"FYRNOX-KillSwitch-AllowDHCP",
	"FYRNOX-KillSwitch-AllowVPN",
	"FYRNOX-KillSwitch-Block",
}

func (k *KillSwitchManager) runCmd(timeout time.Duration, name string, args ...string) (string, error) {
	if k.cmdRunner != nil {
		return k.cmdRunner(timeout, name, args...)
	}
	return runSilentCmdWithTimeout(timeout, name, args...)
}

func (k *KillSwitchManager) checkAdmin() bool {
	if k.isAdminFn != nil {
		return k.isAdminFn()
	}
	return IsAdmin()
}

func (k *KillSwitchManager) deleteRules(rules []string) error {
	var errs []error
	for _, r := range rules {
		out, err := k.runCmd(800*time.Millisecond, "netsh", "advfirewall", "firewall", "delete", "rule", "name="+r)
		if err != nil {
			low := strings.ToLower(out + " " + err.Error())
			// netsh exits with 1 ("No rules match the specified criteria") when a rule is already absent.
			if strings.Contains(low, "no rules match") {
				continue
			}
			errs = append(errs, fmt.Errorf("delete %s: %w", r, err))
		}
	}
	return errors.Join(errs...)
}

// CleanupStaleKillSwitchRules removes any lingering FYRNOX KillSwitch firewall rules.
func CleanupStaleKillSwitchRules() {
	_ = globalKillSwitch.deleteRules(killSwitchRuleNames)
}

type killSwitchRuleSpec struct {
	name string
	args []string
}

// Arm arms the KillSwitch firewall filter, blocking non-VPN traffic while permitting LAN and VPN server IP.
func (k *KillSwitchManager) Arm(serverIP string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.active {
		return nil
	}
	if !k.checkAdmin() {
		return errors.New("kill switch requires Administrator privileges")
	}

	LogInfo("Arming FYRNOX Kill Switch (leak prevention)...")

	// Clean any previous stale rules before installing fresh rules
	_ = k.deleteRules(killSwitchRuleNames)

	// Permit IPv4 LAN/loopback plus IPv6 loopback (::1) and link-local (fe80::/10) only.
	specs := []killSwitchRuleSpec{
		{
			name: "FYRNOX-KillSwitch-AllowLAN",
			args: []string{
				"advfirewall", "firewall", "add", "rule",
				"name=FYRNOX-KillSwitch-AllowLAN",
				"dir=out",
				"action=allow",
				"remoteip=127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,::1,fe80::/10",
			},
		},
		{
			name: "FYRNOX-KillSwitch-AllowDHCP",
			args: []string{
				"advfirewall", "firewall", "add", "rule",
				"name=FYRNOX-KillSwitch-AllowDHCP",
				"dir=out",
				"action=allow",
				"protocol=UDP",
				"remoteport=67,68",
			},
		},
	}
	if serverIP != "" {
		specs = append(specs, killSwitchRuleSpec{
			name: "FYRNOX-KillSwitch-AllowVPN",
			args: []string{
				"advfirewall", "firewall", "add", "rule",
				"name=FYRNOX-KillSwitch-AllowVPN",
				"dir=out",
				"action=allow",
				"remoteip=" + serverIP,
			},
		})
	}
	specs = append(specs, killSwitchRuleSpec{
		name: "FYRNOX-KillSwitch-Block",
		args: []string{
			"advfirewall", "firewall", "add", "rule",
			"name=FYRNOX-KillSwitch-Block",
			"dir=out",
			"action=block",
		},
	})

	installed := make([]string, 0, len(specs))
	rollback := func() error {
		if len(installed) == 0 {
			return nil
		}
		rev := make([]string, len(installed))
		for i, name := range installed {
			rev[len(installed)-1-i] = name
		}
		return k.deleteRules(rev)
	}

	for _, spec := range specs {
		if out, err := k.runCmd(1*time.Second, "netsh", spec.args...); err != nil {
			rbErr := rollback()
			if rbErr != nil {
				return fmt.Errorf("failed to install kill switch rule %s (%v: %s); rollback failed: %w", spec.name, err, strings.TrimSpace(out), rbErr)
			}
			return fmt.Errorf("failed to install kill switch rule %s: %w (%s)", spec.name, err, strings.TrimSpace(out))
		}
		installed = append(installed, spec.name)

		// Verify the rule is present in the firewall store
		if out, err := k.runCmd(1*time.Second, "netsh", "advfirewall", "firewall", "show", "rule", "name="+spec.name); err != nil {
			rbErr := rollback()
			if rbErr != nil {
				return fmt.Errorf("failed to verify kill switch rule %s (%v: %s); rollback failed: %w", spec.name, err, strings.TrimSpace(out), rbErr)
			}
			return fmt.Errorf("failed to verify kill switch rule %s: %w", spec.name, err)
		}
	}

	k.active = true
	LogInfo("FYRNOX Kill Switch armed successfully (%d rules verified)", len(installed))
	return nil
}

// Disarm releases all Kill Switch blocking rules safely and idempotently.
func (k *KillSwitchManager) Disarm() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.active {
		LogInfo("Disarming FYRNOX Kill Switch...")
	}
	err := k.deleteRules(killSwitchRuleNames)
	if err != nil {
		return fmt.Errorf("disarming kill switch: %w", err)
	}
	k.active = false
	return nil
}

// IsActive returns whether the KillSwitch is currently armed.
func (k *KillSwitchManager) IsActive() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.active
}
