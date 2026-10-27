package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultLANHttpPort  = 10809
	defaultLANSocksPort = 10808
	lanFirewallRuleName = "FYRNOX LAN Proxy"
)

type LANShareStatus struct {
	Enabled   bool   `json:"enabled"`
	LanIP     string `json:"lan_ip"`
	HttpPort  int    `json:"http_port"`
	SocksPort int    `json:"socks_port"`
	ProxyURL  string `json:"proxy_url"`
}

type LANShareManager struct {
	mu          sync.Mutex
	enabled     bool
	lanIP       string
	httpPort    int
	socksPort   int
	httpServer  *HttpProxyServer
	socksServer *SocksProxyServer
}

var globalLANShare = &LANShareManager{
	httpPort:  defaultLANHttpPort,
	socksPort: defaultLANSocksPort,
}

// GetLocalSubnetIP finds the active private IPv4 address for LAN sharing
func GetLocalSubnetIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "tun") || strings.Contains(name, "tap") || strings.Contains(name, "fyrnox") || strings.Contains(name, "virtual") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip4 := ip.To4()
			if ip4 == nil {
				continue
			}

			// Check RFC 1918 private ranges
			if (ip4[0] == 192 && ip4[1] == 168) ||
				(ip4[0] == 10) ||
				(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) {
				return ip4.String()
			}
		}
	}
	return "127.0.0.1"
}

func (m *LANShareManager) SetEnabled(enable bool, opener tunnelOpener, meter *TrafficMeter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.enabled == enable {
		return nil
	}

	if enable {
		lanIP := GetLocalSubnetIP()
		m.lanIP = lanIP

		httpAddr := fmt.Sprintf("0.0.0.0:%d", m.httpPort)
		socksAddr := fmt.Sprintf("0.0.0.0:%d", m.socksPort)

		httpSrv := NewHttpProxyServer(httpAddr, opener, meter)
		if err := httpSrv.Start(); err != nil {
			return fmt.Errorf("failed to start LAN HTTP proxy: %w", err)
		}

		socksSrv := NewSocksProxyServer(socksAddr, opener, meter)
		if err := socksSrv.Start(); err != nil {
			_ = httpSrv.Close()
			return fmt.Errorf("failed to start LAN SOCKS5 proxy: %w", err)
		}

		m.httpServer = httpSrv
		m.socksServer = socksSrv
		m.enabled = true

		// Add temporary inbound Windows Firewall rule if elevated
		// shortcut: non-admin ignores firewall failure and falls back to standard user popup, upgrade via service helper if needed
		go func() {
			ports := fmt.Sprintf("%d,%d", m.httpPort, m.socksPort)
			_, _ = runSilentCmdWithTimeout(2*time.Second, "netsh", "advfirewall", "firewall", "add", "rule",
				fmt.Sprintf("name=%s", lanFirewallRuleName), "dir=in", "action=allow", "protocol=TCP",
				fmt.Sprintf("localport=%s", ports), "profile=private,public")
		}()

		LogInfo("LAN Proxy Sharing activated: http://%s:%d (SOCKS5: %s:%d)", lanIP, m.httpPort, lanIP, m.socksPort)
	} else {
		if m.httpServer != nil {
			_ = m.httpServer.Close()
			m.httpServer = nil
		}
		if m.socksServer != nil {
			_ = m.socksServer.Close()
			m.socksServer = nil
		}
		m.enabled = false

		// Remove Windows Firewall rule
		go func() {
			_, _ = runSilentCmdWithTimeout(2*time.Second, "netsh", "advfirewall", "firewall", "delete", "rule",
				fmt.Sprintf("name=%s", lanFirewallRuleName))
		}()

		LogInfo("LAN Proxy Sharing deactivated")
	}

	return nil
}

func (m *LANShareManager) GetStatus() LANShareStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	lanIP := m.lanIP
	if lanIP == "" || lanIP == "127.0.0.1" {
		lanIP = GetLocalSubnetIP()
	}

	return LANShareStatus{
		Enabled:   m.enabled,
		LanIP:     lanIP,
		HttpPort:  m.httpPort,
		SocksPort: m.socksPort,
		ProxyURL:  fmt.Sprintf("http://%s:%d", lanIP, m.httpPort),
	}
}
