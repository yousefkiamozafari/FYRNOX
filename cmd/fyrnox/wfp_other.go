//go:build !windows

package main

import (
	"net"
	"sync"
)

type WfpManager struct {
	mu     sync.Mutex
	active bool
}

func NewWfpManager() *WfpManager {
	return &WfpManager{}
}

func (m *WfpManager) EnableIPv6LeakProtection(tunIfIndex uint32, outerMasqueIPv6 net.IP, outerMasquePort uint16) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = true
	return nil
}

func (m *WfpManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.active = false
	return nil
}

func (m *WfpManager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

func (m *WfpManager) FilterCount() int {
	return 0
}
