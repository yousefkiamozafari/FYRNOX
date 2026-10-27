package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	vpnclient "firefox-vpn-client"
	"golang.org/x/net/proxy"
)

type VpnState string

const (
	StateDisconnected  VpnState = "disconnected"
	StateConnecting    VpnState = "connecting"
	StateConnected     VpnState = "connected"
	StateDisconnecting VpnState = "disconnecting"
	StateReconnecting  VpnState = "reconnecting"
	StateError         VpnState = "error"
)

type ServerLocation struct {
	Hostname        string `json:"hostname"`
	CountryName     string `json:"country_name"`
	CountryCode     string `json:"country_code"`
	CityName        string `json:"city_name"`
	CityCode        string `json:"city_code"`
	Port            int    `json:"port"`
	PingMs          int    `json:"ping_ms"`
	DnsMs           int    `json:"dns_ms,omitempty"`
	TcpMs           int    `json:"tcp_ms,omitempty"`
	TlsMs           int    `json:"tls_ms,omitempty"`
	IsAnycast       bool   `json:"is_anycast"`
	IsRecommended   bool   `json:"is_recommended"`
	RecommendReason string `json:"recommend_reason,omitempty"`
}

type UserQuotaInfo struct {
	Subscribed      bool    `json:"subscribed"`
	UID             int64   `json:"uid"`
	IsAuthoritative bool    `json:"is_authoritative"`
	Source          string  `json:"source"`
	MaxBytes        string  `json:"max_bytes"`
	MaxBytesNum     int64   `json:"max_bytes_num"`
	QuotaLeft       string  `json:"quota_left"`
	QuotaLeftNum    int64   `json:"quota_left_num"`
	QuotaUsed       string  `json:"quota_used"`
	QuotaUsedNum    int64   `json:"quota_used_num"`
	QuotaReset      string  `json:"quota_reset"`
	QuotaPercent    float64 `json:"quota_percent"`
}

type VpnSettings struct {
	SocksPort          int      `json:"socks_port"`
	HttpPort           int      `json:"http_port"`
	UseSystemProxy     bool     `json:"use_system_proxy"`
	EnableTunMode      bool     `json:"enable_tun_mode"`
	UseH3              bool     `json:"use_h3"`
	AutoConnect        bool     `json:"auto_connect"`
	SelectedServer     string   `json:"selected_server"`
	ApiProxy           string   `json:"api_proxy"`
	TlsFragmentation   bool     `json:"tls_fragmentation"`
	BypassRules        string   `json:"bypass_rules"`
	SplitTunnelEnabled bool     `json:"split_tunnel_enabled"`
	SplitTunnelMode    string   `json:"split_tunnel_mode"` // "bypass" or "proxy_only"
	SplitTunnelApps    []string `json:"split_tunnel_apps"`
	DirectIranTraffic  bool     `json:"direct_iran_traffic"`
	GamingMode         bool     `json:"gaming_mode"`
	IPProtocol         string   `json:"ip_protocol"` // "dual", "ipv4", "ipv6"
}

type VpnStatus struct {
	Version            string          `json:"version"`
	State              VpnState        `json:"state"`
	ConnectStep        string          `json:"connect_step,omitempty"`
	ErrorMessage       string          `json:"error_message,omitempty"`
	Server             *ServerLocation `json:"server,omitempty"`
	Recommended        *ServerLocation `json:"recommended_server,omitempty"`
	CurrentIP          string          `json:"current_ip"`
	CurrentCountry     string          `json:"current_country"`
	IsVPNIP            bool            `json:"is_vpn_ip"`
	ExitIP             string          `json:"exit_ip,omitempty"`
	ExitCountry        string          `json:"exit_country,omitempty"`
	ExitLatency        int64           `json:"exit_latency,omitempty"`
	IsLoggedIn         bool            `json:"is_logged_in"`
	UserEmail          string          `json:"user_email,omitempty"`
	Traffic            TrafficStats    `json:"traffic"`
	Quota              *UserQuotaInfo  `json:"quota,omitempty"`
	Settings           VpnSettings     `json:"settings"`
	IsAdmin            bool            `json:"is_admin"`
	TunActive          bool            `json:"tun_active"`
	WintunActive       bool            `json:"wintun_active"`
	GamingStats        GamingStats     `json:"gaming_stats"`
	SplitTunnelEnabled bool            `json:"split_tunnel_enabled"`
	IPProtocol         string          `json:"ip_protocol"`
	LocalIPv4Available bool            `json:"local_ipv4_available"`
	LocalIPv6Available bool            `json:"local_ipv6_available"`
	TransportProtocol  string          `json:"transport_protocol,omitempty"`
	ExitAddressFamily  string          `json:"exit_address_family,omitempty"`
	FallbackActive     bool            `json:"fallback_active"`
	FallbackState      string          `json:"fallback_state,omitempty"`
}

type VpnEngine struct {
	mu             sync.RWMutex
	state          VpnState
	connectStep    string
	errorMsg       string
	settings       VpnSettings
	servers        []ServerLocation
	selected       *ServerLocation
	auth           *runtimeAuth
	userEmail      string
	quota          *UserQuotaInfo
	exitResult     *ExitProbeResult

	currentIP      string
	currentCountry string
	lastIPFetch    time.Time
	ipFetchMu      sync.Mutex

	meter          *TrafficMeter
	monthlyTracker *MonthlyQuotaTracker
	controller     *proxyController
	socksServer    *SocksProxyServer
	httpServer     *HttpProxyServer

	guardian            string
	pendingSessionToken string

	connectCancel       context.CancelFunc
	connectID           int64
}

func checkLocalIPFamilies() (hasV4 bool, hasV6 bool) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return true, false
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "tun") || strings.Contains(name, "tap") || strings.Contains(name, "fyrnox") {
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
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				hasV4 = true
			} else if ip.To16() != nil {
				hasV6 = true
			}
		}
	}
	return hasV4, hasV6
}

func NewVpnEngine() *VpnEngine {
	meter := NewTrafficMeter()
	monthlyTracker := NewMonthlyQuotaTracker("")
	settings := loadSettings()
	used, left, resetDate, pct := monthlyTracker.GetMonthlyUsage()

	engine := &VpnEngine{
		state:          StateDisconnected,
		guardian:       vpnclient.GuardianEndpointDefault,
		meter:          meter,
		monthlyTracker: monthlyTracker,
		settings:       settings,
		quota: &UserQuotaInfo{
			Subscribed:      false,
			IsAuthoritative: false,
			Source:          "Estimated local usage (may differ from Firefox account)",
			MaxBytes:        "50.0 GB",
			MaxBytesNum:     MonthlyQuotaAllowanceBytes,
			QuotaLeft:       FormatBytes(left),
			QuotaLeftNum:    left,
			QuotaUsed:       FormatBytes(used),
			QuotaUsedNum:    used,
			QuotaReset:      resetDate,
			QuotaPercent:    pct,
		},
	}

	defaultServers := []ServerLocation{
		{Hostname: "us.m1.fastly-masque.net", CountryName: "United States", CountryCode: "US", CityName: "Anycast Edge", CityCode: "US", Port: 2499, PingMs: -1, IsAnycast: true},
		{Hostname: "p.m1.fastly-masque.net", CountryName: "CatchAll Anycast", CountryCode: "US", CityName: "Global Anycast", CityCode: "Q30", Port: 2499, PingMs: -1, IsAnycast: true},
		{Hostname: "muc139.m1.fastly-masque.net", CountryName: "Germany", CountryCode: "DE", CityName: "Munich", CityCode: "MUC", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "eglc860.m1.fastly-masque.net", CountryName: "United Kingdom", CountryCode: "GB", CityName: "London", CityCode: "EGLC", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "lfpb115.m1.fastly-masque.net", CountryName: "France", CountryCode: "FR", CityName: "Paris", CityCode: "LFPB", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "ehrd229.m1.fastly-masque.net", CountryName: "Netherlands", CountryCode: "NL", CityName: "Rotterdam", CityCode: "EHRD", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "ca.m1.fastly-masque.net", CountryName: "Canada", CountryCode: "CA", CityName: "Toronto", CityCode: "CA", Port: 2499, PingMs: -1, IsAnycast: true},
		{Hostname: "rjtf770.m1.fastly-masque.net", CountryName: "Japan", CountryCode: "JP", CityName: "Tokyo", CityCode: "RJTF", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "wsat188.m1.fastly-masque.net", CountryName: "Singapore", CountryCode: "SG", CityName: "Singapore", CityCode: "WSAT", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "ysbk106.m1.fastly-masque.net", CountryName: "Australia", CountryCode: "AU", CityName: "Sydney", CityCode: "YSBK", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "essb127.m1.fastly-masque.net", CountryName: "Sweden", CountryCode: "SE", CityName: "Stockholm", CityCode: "ESSB", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "hel141.m1.fastly-masque.net", CountryName: "Finland", CountryCode: "FI", CityName: "Helsinki", CityCode: "HEL", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "osl65.m1.fastly-masque.net", CountryName: "Norway", CountryCode: "NO", CityName: "Oslo", CityCode: "OSL", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "cph232.m1.fastly-masque.net", CountryName: "Denmark", CountryCode: "DK", CityName: "Copenhagen", CityCode: "CPH", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "dub43.m1.fastly-masque.net", CountryName: "Ireland", CountryCode: "IE", CityName: "Dublin", CityCode: "DUB", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "mxp69.m1.fastly-masque.net", CountryName: "Italy", CountryCode: "IT", CityName: "Milan", CityCode: "MXP", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "leto235.m1.fastly-masque.net", CountryName: "Spain", CountryCode: "ES", CityName: "Madrid", CityCode: "LETO", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "vie63.m1.fastly-masque.net", CountryName: "Austria", CountryCode: "AT", CityName: "Vienna", CityCode: "VIE", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "bru148.m1.fastly-masque.net", CountryName: "Belgium", CountryCode: "BE", CityName: "Brussels", CityCode: "BRU", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "lis149.m1.fastly-masque.net", CountryName: "Portugal", CountryCode: "PT", CityName: "Lisbon", CityCode: "LIS", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "faor73.m1.fastly-masque.net", CountryName: "South Africa", CountryCode: "ZA", CityName: "Johannesburg", CityCode: "FAOR", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "spjc128.m1.fastly-masque.net", CountryName: "Peru", CountryCode: "PE", CityName: "Lima", CityCode: "SPJC", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "sbsp209.m1.fastly-masque.net", CountryName: "Brazil", CountryCode: "BR", CityName: "Sao Paulo", CityCode: "SBSP", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "icn145.m1.fastly-masque.net", CountryName: "Korea", CountryCode: "KR", CityName: "Seoul", CityCode: "ICN", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "acc96.m1.fastly-masque.net", CountryName: "Ghana", CountryCode: "GH", CityName: "Accra", CityCode: "ACC", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "eze223.m1.fastly-masque.net", CountryName: "Argentina", CountryCode: "AR", CityName: "Buenos Aires", CityCode: "EZE", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "mnl215.m1.fastly-masque.net", CountryName: "Philippines", CountryCode: "PH", CityName: "Manila", CityCode: "MNL", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "akl103.m1.fastly-masque.net", CountryName: "New Zealand", CountryCode: "NZ", CityName: "Auckland", CityCode: "AKL", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "scl222.m1.fastly-masque.net", CountryName: "Chile", CountryCode: "CL", CityName: "Santiago", CityCode: "SCL", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "kul98.m1.fastly-masque.net", CountryName: "Malaysia", CountryCode: "MY", CityName: "Kuala Lumpur", CityCode: "KUL", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "mmqt107.m1.fastly-masque.net", CountryName: "Mexico", CountryCode: "MX", CityName: "Queretaro", CityCode: "MMQT", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "skbo234.m1.fastly-masque.net", CountryName: "Colombia", CountryCode: "CO", CityName: "Bogota", CityCode: "SKBO", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "bkk228.m1.fastly-masque.net", CountryName: "Thailand", CountryCode: "TH", CityName: "Bangkok", CityCode: "BKK", Port: 2499, PingMs: -1, IsAnycast: false},
		{Hostname: "sof151.m1.fastly-masque.net", CountryName: "Bulgaria", CountryCode: "BG", CityName: "Sofia", CityCode: "SOF", Port: 2499, PingMs: -1, IsAnycast: false},
	}

	cached := loadCachedServers()
	if len(cached) > 0 {
		engine.servers = cached
	} else {
		engine.servers = defaultServers
	}
	engine.selected = &engine.servers[0]

	// Setup custom DNS / direct transport with TLS fragmentation by default
	_ = configureAPIProxy("", true)

	// Try loading saved auth tokens
	saved, err := vpnclient.LoadTokens()
	if err == nil && saved != nil {
		engine.auth = &runtimeAuth{
			Token: &vpnclient.TokenResponse{
				AccessToken:  saved.AccessToken,
				RefreshToken: saved.RefreshToken,
				ExpiresIn:    saved.ExpiresIn,
				Scope:        saved.Scope,
				TokenType:    "bearer",
			},
			ObtainedAt: saved.ObtainedAt,
		}
		engine.userEmail = saved.UserEmail
		if engine.userEmail != "" {
			LogInfo("Loaded saved Firefox account session for %s", engine.userEmail)
		} else {
			LogInfo("Loaded saved Firefox account session")
		}
	}

	go engine.refreshServers()
	go engine.RefreshUserIP()
	if engine.auth != nil {
		go engine.refreshUserInfo()
	}

	return engine
}

func getServerCachePath() string {
	appData, err := os.UserConfigDir()
	if err == nil {
		dir := filepath.Join(appData, "Fyrnox")
		_ = os.MkdirAll(dir, 0755)
		return filepath.Join(dir, "servers_cache.json")
	}
	return "servers_cache.json"
}

func loadCachedServers() []ServerLocation {
	data, err := os.ReadFile(getServerCachePath())
	if err != nil {
		return nil
	}
	var list []ServerLocation
	if err := json.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list
	}
	return nil
}

func saveCachedServers(list []ServerLocation) {
	if len(list) == 0 {
		return
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(getServerCachePath(), data, 0644)
}

func (e *VpnEngine) setConnectStep(step string) {
	e.mu.Lock()
	e.connectStep = step
	e.mu.Unlock()
}

func (e *VpnEngine) GetStatus() VpnStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ipInfo := e.resolveCurrentIPLocked()

	hasV4, hasV6 := checkLocalIPFamilies()
	ipProto := e.settings.IPProtocol
	if ipProto == "" {
		ipProto = "dual"
	}

	quotaObj := e.quota
	if (quotaObj == nil || !quotaObj.IsAuthoritative) && e.monthlyTracker != nil {
		used, left, resetDate, pct := e.monthlyTracker.GetMonthlyUsage()
		quotaObj = &UserQuotaInfo{
			Subscribed:      false,
			IsAuthoritative: false,
			Source:          "Estimated local usage (may differ from Firefox account)",
			MaxBytes:        "50.0 GB",
			MaxBytesNum:     MonthlyQuotaAllowanceBytes,
			QuotaLeft:       FormatBytes(left),
			QuotaLeftNum:    left,
			QuotaUsed:       FormatBytes(used),
			QuotaUsedNum:    used,
			QuotaReset:      resetDate,
			QuotaPercent:    pct,
		}
	}

	return VpnStatus{
		Version:        "1.0.4",
		State:          e.state,
		ConnectStep:    e.connectStep,
		ErrorMessage:   e.errorMsg,
		Server:         e.selected,
		Recommended:    e.getRecommendedLocked(),
		CurrentIP:      ipInfo.IP,
		CurrentCountry: func() string {
			if e.currentCountry != "" {
				return e.currentCountry
			}
			return ipInfo.CountryCode
		}(),
		IsVPNIP:        ipInfo.IsVPN,
		ExitIP:         func() string { if e.exitResult != nil { return e.exitResult.IP }; return "" }(),
		ExitCountry:    func() string { if e.exitResult != nil { return e.exitResult.CountryCode }; return "" }(),
		ExitLatency:    func() int64 { if e.exitResult != nil { return e.exitResult.LatencyMs }; return 0 }(),
		IsLoggedIn:     e.auth != nil,
		UserEmail:      e.userEmail,
		Traffic:        e.meter.GetStats(),
		Quota:          quotaObj,
		Settings:       e.settings,
		IsAdmin:            IsAdmin(),
		TunActive:          IsTunModeActive(),
		WintunActive:       IsWintunActive(),
		GamingStats:        globalGamer.GetStats(),
		SplitTunnelEnabled: e.settings.SplitTunnelEnabled,
		IPProtocol:         ipProto,
		LocalIPv4Available: hasV4,
		LocalIPv6Available: hasV6,
		TransportProtocol: func() string {
			if e.state == StateConnected {
				switch ipProto {
				case "ipv4":
					return "IPv4 (tcp4)"
				case "ipv6":
					return "IPv6 (tcp6)"
				default:
					return "Dual Stack (tcp)"
				}
			}
			return ""
		}(),
		ExitAddressFamily: func() string {
			if e.exitResult != nil && e.exitResult.IP != "" {
				if parsed := net.ParseIP(e.exitResult.IP); parsed != nil {
					if parsed.To4() != nil {
						return "IPv4"
					}
					return "IPv6"
				}
			}
			return ""
		}(),
		FallbackActive: func() bool {
			if e.controller != nil {
				active, _, _ := e.controller.GetFallbackStatus()
				return active
			}
			return false
		}(),
		FallbackState: func() string {
			if e.controller != nil {
				_, state, _ := e.controller.GetFallbackStatus()
				return state
			}
			return ""
		}(),
	}
}

func (e *VpnEngine) GetTunnelOpener() tunnelOpener {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.controller
}

func (e *VpnEngine) getRecommendedLocked() *ServerLocation {
	for i := range e.servers {
		if e.servers[i].IsRecommended {
			rec := e.servers[i]
			return &rec
		}
	}
	var best *ServerLocation
	minPing := 99999
	for i := range e.servers {
		if e.servers[i].PingMs > 0 && e.servers[i].PingMs < minPing {
			minPing = e.servers[i].PingMs
			rec := e.servers[i]
			best = &rec
		}
	}
	return best
}

func (e *VpnEngine) GetServers() []ServerLocation {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]ServerLocation, len(e.servers))
	copy(res, e.servers)
	return res
}

func (e *VpnEngine) refreshServers() {
	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()

	countries, err := vpnclient.FetchServerList(ctx)
	if err != nil {
		LogWarn("Failed to fetch server list: %v (using cached/default servers)", err)
		return
	}

	var list []ServerLocation
	for _, c := range countries {
		for _, city := range c.Cities {
			for _, srv := range city.Servers {
				if srv.Quarantined {
					continue
				}
				port := 443
				host := ""
				// Prioritize connect or masque protocols
				for _, p := range srv.Protocols {
					if (p.Name == "connect" || p.Name == "masque") && p.Host != "" {
						host = p.Host
						if p.Port > 0 {
							port = p.Port
						}
						break
					}
				}
				if host == "" && len(srv.Protocols) == 0 && srv.Hostname != "" {
					host = srv.Hostname
					if srv.Port > 0 {
						port = srv.Port
					}
				}
				if host != "" {
					isAny := strings.Contains(host, "us.m1") || strings.Contains(host, "p.m1") || strings.Contains(host, "ca.m1")
					list = append(list, ServerLocation{
						Hostname:    host,
						CountryName: c.Name,
						CountryCode: c.Code,
						CityName:    city.Name,
						CityCode:    city.Code,
						Port:        port,
						PingMs:      -1,
						IsAnycast:   isAny,
					})
				}
			}
		}
	}

	if len(list) > 0 {
		saveCachedServers(list)
		e.mu.Lock()
		e.servers = list
		// Validate and preserve selection
		found := false
		if e.selected != nil {
			for i := range list {
				if list[i].Hostname == e.selected.Hostname {
					e.selected = &list[i]
					found = true
					break
				}
			}
		}
		if !found {
			e.selected = &list[0]
		}
		e.mu.Unlock()

		LogInfo("Loaded and cached %d Mozilla VPN servers across %d countries", len(list), len(countries))
		go e.PingAllServers()
	}
}

var (
	dnsCacheMu sync.RWMutex
	dnsCache   = make(map[string]net.IP)
)

func resolveServerIP(ctx context.Context, hostname string) (net.IP, error) {
	dnsCacheMu.RLock()
	cachedIP, found := dnsCache[hostname]
	dnsCacheMu.RUnlock()
	if found && cachedIP != nil {
		return cachedIP, nil
	}

	dialer := createProxyDialer(2 * time.Second)
	ips, err := dialer.Resolver.LookupIP(ctx, "ip4", hostname)
	if err == nil && len(ips) > 0 {
		dnsCacheMu.Lock()
		dnsCache[hostname] = ips[0]
		dnsCacheMu.Unlock()
		return ips[0], nil
	}

	stdIPs, stdErr := net.LookupIP(hostname)
	if stdErr == nil && len(stdIPs) > 0 {
		dnsCacheMu.Lock()
		dnsCache[hostname] = stdIPs[0]
		dnsCacheMu.Unlock()
		return stdIPs[0], nil
	}

	return nil, fmt.Errorf("could not resolve %s", hostname)
}

func computeMedianLatency(samples []int) int {
	var valid []int
	for _, s := range samples {
		if s > 0 {
			valid = append(valid, s)
		}
	}
	if len(valid) == 0 {
		return -1
	}
	sort.Ints(valid)
	n := len(valid)
	if n%2 == 1 {
		return valid[n/2]
	}
	return (valid[n/2-1] + valid[n/2]) / 2
}

func measureServerPing(ctx context.Context, s ServerLocation) (pingMs, dnsMs, tcpMs, tlsMs int) {
	port := s.Port
	if port <= 0 {
		port = 2499
	}

	// 1. DNS Resolution latency
	t0 := time.Now()
	ip, err := resolveServerIP(ctx, s.Hostname)
	dnsMs = int(time.Since(t0).Milliseconds())
	if dnsMs <= 0 {
		dnsMs = 1
	}

	target := fmt.Sprintf("%s:%d", s.Hostname, port)
	if err == nil && ip != nil {
		target = net.JoinHostPort(ip.String(), strconv.Itoa(port))
	}

	// 2. Real TCP Connect (SYN/ACK) Round-Trip Time with strict 2000ms ceiling
	t1 := time.Now()
	dialer := &net.Dialer{Timeout: 2000 * time.Millisecond}
	rawConn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return -1, dnsMs, -1, -1
	}
	tcpMs = int(time.Since(t1).Milliseconds())
	if tcpMs <= 0 {
		tcpMs = 1
	}

	// 3. Transport Handshake (TLS Handshake)
	t2 := time.Now()
	tlsCfg := createProxyTLSConfig(s.Hostname, []string{"h2"})
	tlsConn := tls.Client(rawConn, tlsCfg)
	hErr := tlsConn.HandshakeContext(ctx)
	if hErr != nil {
		_ = rawConn.Close()
		return -1, dnsMs, tcpMs, -1
	}
	tlsMs = int(time.Since(t2).Milliseconds())
	if tlsMs <= 0 {
		tlsMs = 1
	}
	_ = tlsConn.Close()

	pingMs = tcpMs + tlsMs
	return pingMs, dnsMs, tcpMs, tlsMs
}

func (e *VpnEngine) PingAllServers() []ServerLocation {
	e.mu.RLock()
	servers := make([]ServerLocation, len(e.servers))
	copy(servers, e.servers)
	e.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	// Worker pool with concurrency limit of 8 parallel pings to avoid socket exhaustion
	limit := make(chan struct{}, 8)

	for i := range servers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()

			p, d, tc, tl := measureServerPing(ctx, servers[idx])
			servers[idx].PingMs = p
			servers[idx].DnsMs = d
			servers[idx].TcpMs = tc
			servers[idx].TlsMs = tl
		}(i)
	}

	wg.Wait()

	// Sort so the lowest latency servers appear at the top
	sort.SliceStable(servers, func(i, j int) bool {
		pi := servers[i].PingMs
		pj := servers[j].PingMs
		if pi > 0 && pj > 0 {
			return pi < pj
		}
		if pi > 0 {
			return true
		}
		return false
	})

	// Compute recommendation based on actual measurements (never auto-switch user selection)
	for i := range servers {
		servers[i].IsRecommended = false
		servers[i].RecommendReason = ""
	}
	for i := range servers {
		if servers[i].PingMs > 0 && servers[i].PingMs < 350 {
			servers[i].IsRecommended = true
			servers[i].RecommendReason = fmt.Sprintf("Fastest verified edge (%d ms)", servers[i].PingMs)
			break
		}
	}

	e.mu.Lock()
	e.servers = servers
	if e.selected != nil {
		for _, s := range servers {
			if s.Hostname == e.selected.Hostname {
				e.selected.PingMs = s.PingMs
				break
			}
		}
	}
	res := make([]ServerLocation, len(servers))
	copy(res, servers)
	e.mu.Unlock()
	return res
}

func (e *VpnEngine) PingServer(hostname string) int {
	e.mu.RLock()
	var target ServerLocation
	found := false
	for _, s := range e.servers {
		if s.Hostname == hostname {
			target = s
			found = true
			break
		}
	}
	if !found && e.selected != nil {
		target = *e.selected
	}
	e.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p, d, tc, tl := measureServerPing(ctx, target)

	e.mu.Lock()
	for i := range e.servers {
		if e.servers[i].Hostname == target.Hostname {
			e.servers[i].PingMs = p
			e.servers[i].DnsMs = d
			e.servers[i].TcpMs = tc
			e.servers[i].TlsMs = tl
			break
		}
	}
	if e.selected != nil && e.selected.Hostname == target.Hostname {
		e.selected.PingMs = p
		e.selected.DnsMs = d
		e.selected.TcpMs = tc
		e.selected.TlsMs = tl
	}
	e.mu.Unlock()
	ping := p

	LogInfo("Pinged server %s: %d ms", target.Hostname, ping)
	return ping
}

// PingTunnel measures the live round-trip latency through the active tunnel
func (e *VpnEngine) PingTunnel() (int64, error) {
	e.mu.RLock()
	state := e.state
	ctrl := e.controller
	selected := e.selected
	e.mu.RUnlock()

	if state == StateConnected && ctrl != nil {
		start := time.Now()
		conn, err := ctrl.OpenTunnel("1.1.1.1:80")
		if err == nil && conn != nil {
			rtt := time.Since(start).Milliseconds()
			_ = conn.Close()
			if rtt <= 0 {
				rtt = 1
			}
			e.mu.Lock()
			if e.exitResult != nil {
				e.exitResult.LatencyMs = rtt
			}
			if e.selected != nil {
				e.selected.PingMs = int(rtt)
			}
			e.mu.Unlock()
			return rtt, nil
		}
	}

	if selected != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
		defer cancel()
		p, _, _, _ := measureServerPing(ctx, *selected)
		if p > 0 {
			e.mu.Lock()
			if e.exitResult != nil {
				e.exitResult.LatencyMs = int64(p)
			}
			if e.selected != nil {
				e.selected.PingMs = p
			}
			e.mu.Unlock()
			return int64(p), nil
		}
	}

	return 0, fmt.Errorf("latency measurement unavailable")
}

type IPInfo struct {
	IP          string `json:"ip"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	IsVPN       bool   `json:"is_vpn"`
}

func getCountryName(code string) string {
	names := map[string]string{
		"IR": "Iran",
		"DE": "Germany",
		"NL": "Netherlands",
		"GB": "United Kingdom",
		"US": "United States",
		"FR": "France",
		"TR": "Turkey",
		"AE": "United Arab Emirates",
		"CA": "Canada",
		"FI": "Finland",
		"SE": "Sweden",
		"NO": "Norway",
		"CH": "Switzerland",
		"IT": "Italy",
		"ES": "Spain",
		"JP": "Japan",
		"SG": "Singapore",
		"AU": "Australia",
		"AT": "Austria",
		"BE": "Belgium",
		"PL": "Poland",
		"DK": "Denmark",
		"IE": "Ireland",
		"CZ": "Czech Republic",
	}
	if n, ok := names[strings.ToUpper(code)]; ok {
		return n
	}
	return code
}

func (e *VpnEngine) resolveCurrentIPLocked() IPInfo {
	if e.state == StateConnected {
		if e.exitResult != nil && e.exitResult.IP != "" {
			return IPInfo{
				IP:          e.exitResult.IP,
				CountryCode: e.exitResult.CountryCode,
				CountryName: getCountryName(e.exitResult.CountryCode),
				IsVPN:       true,
			}
		}
		if e.selected != nil {
			return IPInfo{
				IP:          e.selected.Hostname,
				CountryCode: e.selected.CountryCode,
				CountryName: e.selected.CountryName,
				IsVPN:       true,
			}
		}
	}

	if e.currentIP != "" {
		return IPInfo{
			IP:          e.currentIP,
			CountryCode: e.currentCountry,
			CountryName: getCountryName(e.currentCountry),
			IsVPN:       false,
		}
	}

	go e.RefreshUserIP()

	return IPInfo{
		IP:          "Checking IP...",
		CountryCode: "",
		CountryName: "",
		IsVPN:       false,
	}
}

func (e *VpnEngine) RefreshUserIP() IPInfo {
	e.ipFetchMu.Lock()
	defer e.ipFetchMu.Unlock()

	e.mu.RLock()
	state := e.state
	exit := e.exitResult
	ctrl := e.controller
	e.mu.RUnlock()

	if state == StateConnected && ctrl != nil {
		probe, err := VerifyExit(ctrl, 6*time.Second)
		if err == nil && probe.IP != "" {
			e.mu.Lock()
			e.exitResult = &probe
			e.mu.Unlock()
			return IPInfo{
				IP:          probe.IP,
				CountryCode: probe.CountryCode,
				CountryName: getCountryName(probe.CountryCode),
				IsVPN:       true,
			}
		}
		if exit != nil && exit.IP != "" {
			return IPInfo{
				IP:          exit.IP,
				CountryCode: exit.CountryCode,
				CountryName: getCountryName(exit.CountryCode),
				IsVPN:       true,
			}
		}
	}

	// Fetch public IP via anti-censorship direct transport
	dialer := createProxyDialer(6 * time.Second)
	transport := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	client := &http.Client{Transport: transport, Timeout: 6 * time.Second}

	// 1. Try Cloudflare trace (provides IP & Country)
	resp, err := client.Get("https://www.cloudflare.com/cdn-cgi/trace")
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		ip := ""
		loc := ""
		for _, line := range strings.Split(string(body), "\n") {
			k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
			if ok {
				if strings.ToLower(k) == "ip" {
					ip = v
				} else if strings.ToLower(k) == "loc" {
					loc = v
				}
			}
		}
		if ip != "" {
			e.mu.Lock()
			e.currentIP = ip
			e.currentCountry = loc
			e.lastIPFetch = time.Now()
			e.mu.Unlock()
			LogInfo("Detected user public IP: %s (%s)", ip, loc)
			return IPInfo{
				IP:          ip,
				CountryCode: loc,
				CountryName: getCountryName(loc),
				IsVPN:       false,
			}
		}
	}

	// 2. Fallback to api.ipify.org
	resp2, err2 := client.Get("https://api.ipify.org?format=text")
	if err2 == nil {
		defer resp2.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp2.Body, 100))
		ip := strings.TrimSpace(string(body))
		if ip != "" {
			e.mu.Lock()
			e.currentIP = ip
			e.lastIPFetch = time.Now()
			e.mu.Unlock()
			LogInfo("Detected user public IP: %s", ip)
			return IPInfo{
				IP:          ip,
				CountryCode: "",
				CountryName: "",
				IsVPN:       false,
			}
		}
	}

	e.mu.RLock()
	cur := e.currentIP
	loc := e.currentCountry
	e.mu.RUnlock()

	if cur != "" {
		return IPInfo{
			IP:          cur,
			CountryCode: loc,
			CountryName: getCountryName(loc),
			IsVPN:       false,
		}
	}

	return IPInfo{IP: "Unavailable", CountryCode: "", CountryName: "", IsVPN: false}
}

func (e *VpnEngine) SelectServer(hostname string) error {
	e.mu.Lock()
	var found *ServerLocation
	for _, s := range e.servers {
		if s.Hostname == hostname {
			sel := s
			e.selected = &sel
			e.settings.SelectedServer = hostname
			found = &sel
			break
		}
	}
	if found == nil {
		e.mu.Unlock()
		return fmt.Errorf("server not found: %s", hostname)
	}

	currentState := e.state
	LogInfo("Selected server: %s (%s, %s) [current state: %s]", found.Hostname, found.CityName, found.CountryName, currentState)

	shouldReconnect := (currentState == StateConnected || currentState == StateConnecting)
	if shouldReconnect {
		e.state = StateConnecting
		e.connectStep = "authorizing"
	}
	e.mu.Unlock()

	if shouldReconnect {
		LogInfo("Seamless server switch triggered for %s. Teardown previous route and hot reconnecting...", hostname)
		go func() {
			_ = e.Disconnect()
			time.Sleep(150 * time.Millisecond)
			_ = e.Connect()
		}()
	}

	return nil
}

func (e *VpnEngine) getOrRenewProxyPass(ctx context.Context) (*vpnclient.ProxyPassInfo, error) {
	e.mu.RLock()
	auth := e.auth
	e.mu.RUnlock()
	if auth == nil || auth.Token == nil {
		return nil, fmt.Errorf("not logged in")
	}

	// 1. Try with current access token
	pass, err := vpnclient.FetchProxyPass(ctx, e.guardian, auth.Token.AccessToken)
	if err == nil {
		return pass, nil
	}

	// 2. If token rejected (401/403), first try to ACTIVATE Guardian (new accounts or inactive FPN tier)
	if errors.Is(err, vpnclient.ErrTokenInvalid) {
		LogInfo("Proxy pass returned 401/403. Attempting to activate Guardian account...")
		actCtx, actCancel := context.WithTimeout(context.Background(), apiRequestTimeout)
		_, actErr := vpnclient.ActivateGuardian(actCtx, e.guardian, auth.Token.AccessToken)
		actCancel()
		if actErr == nil {
			LogInfo("Guardian account activated successfully! Retrying proxy pass...")
			pass2, err2 := vpnclient.FetchProxyPass(ctx, e.guardian, auth.Token.AccessToken)
			if err2 == nil {
				return pass2, nil
			}
			err = err2
		} else {
			LogWarn("Activation with current token failed: %v", actErr)
		}

		// 3. Next try refreshing the OAuth token with FxaRefreshToken
		LogInfo("Attempting to refresh OAuth token via RefreshToken...")
		rCtx, rCancel := context.WithTimeout(context.Background(), apiRequestTimeout)
		newToken, rErr := vpnclient.FxaRefreshToken(rCtx, auth.Token.RefreshToken)
		rCancel()
		if rErr == nil && newToken != nil {
			_ = vpnclient.SaveTokens(newToken)
			e.mu.Lock()
			e.auth.Token = newToken
			auth = e.auth
			e.mu.Unlock()

			// Try fetch with new token
			pass3, err3 := vpnclient.FetchProxyPass(ctx, e.guardian, newToken.AccessToken)
			if err3 == nil {
				return pass3, nil
			}

			// If still 403, try activating with the new token
			if errors.Is(err3, vpnclient.ErrTokenInvalid) {
				LogInfo("Activating Guardian with refreshed token...")
				actCtx2, actCancel2 := context.WithTimeout(context.Background(), apiRequestTimeout)
				_, actErr2 := vpnclient.ActivateGuardian(actCtx2, e.guardian, newToken.AccessToken)
				actCancel2()
				if actErr2 == nil {
					LogInfo("Guardian activated with refreshed token!")
					pass4, err4 := vpnclient.FetchProxyPass(ctx, e.guardian, newToken.AccessToken)
					if err4 == nil {
						return pass4, nil
					}
					err = err4
				} else {
					LogWarn("Activation with refreshed token failed: %v", actErr2)
					err = err3
				}
			} else {
				err = err3
			}
		} else {
			LogWarn("RefreshToken failed: %v", rErr)
		}
	}

	return nil, err
}

func (e *VpnEngine) refreshUserInfo() {
	e.mu.RLock()
	auth := e.auth
	e.mu.RUnlock()
	if auth == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	pass, err := e.getOrRenewProxyPass(ctx)
	cancel()
	if err != nil {
		LogWarn("Failed to fetch proxy pass for quota info: %v", err)
		return
	}

	maxBytesNum, _ := strconv.ParseInt(pass.QuotaMax, 10, 64)
	leftBytesNum, _ := strconv.ParseInt(pass.QuotaLeft, 10, 64)

	e.mu.Lock()
	if maxBytesNum > 0 && leftBytesNum >= 0 {
		usedBytesNum := maxBytesNum - leftBytesNum
		if usedBytesNum < 0 {
			usedBytesNum = 0
		}
		var pct float64
		if maxBytesNum > 0 {
			pct = float64(usedBytesNum) / float64(maxBytesNum) * 100
		}
		e.quota = &UserQuotaInfo{
			Subscribed:      true,
			IsAuthoritative: true,
			Source:          "Mozilla Account",
			MaxBytes:        FormatBytes(maxBytesNum),
			MaxBytesNum:     maxBytesNum,
			QuotaLeft:       FormatBytes(leftBytesNum),
			QuotaLeftNum:    leftBytesNum,
			QuotaUsed:       FormatBytes(usedBytesNum),
			QuotaUsedNum:    usedBytesNum,
			QuotaReset:      pass.QuotaReset,
			QuotaPercent:    pct,
		}
		LogInfo("Authoritative Quota updated: %s left of %s (used %s, resets %s)", FormatBytes(leftBytesNum), FormatBytes(maxBytesNum), FormatBytes(usedBytesNum), pass.QuotaReset)
	} else if e.monthlyTracker != nil {
		used, left, resetDate, pct := e.monthlyTracker.GetMonthlyUsage()
		e.quota = &UserQuotaInfo{
			Subscribed:      false,
			IsAuthoritative: false,
			Source:          "Estimated local usage (may differ from Firefox account)",
			MaxBytes:        "50.0 GB",
			MaxBytesNum:     MonthlyQuotaAllowanceBytes,
			QuotaLeft:       FormatBytes(left),
			QuotaLeftNum:    left,
			QuotaUsed:       FormatBytes(used),
			QuotaUsedNum:    used,
			QuotaReset:      resetDate,
			QuotaPercent:    pct,
		}
		LogInfo("Local Estimated Quota updated: %s left of 50.0 GB (used %s, resets %s)", FormatBytes(left), FormatBytes(used), resetDate)
	}
	e.mu.Unlock()
}

func (e *VpnEngine) Connect() error {
	e.mu.Lock()
	if e.state == StateConnected || e.state == StateConnecting {
		e.mu.Unlock()
		return nil
	}
	if e.auth == nil {
		e.state = StateError
		e.errorMsg = "Please sign in with your Firefox account first"
		e.mu.Unlock()
		return fmt.Errorf("not logged in")
	}
	if e.selected == nil {
		e.state = StateError
		e.errorMsg = "No server location selected"
		e.mu.Unlock()
		return fmt.Errorf("no server selected")
	}

	currentConnID := time.Now().UnixNano()
	connectCtx, cancel := context.WithCancel(context.Background())
	e.connectCancel = cancel
	e.connectID = currentConnID

	e.state = StateConnecting
	e.connectStep = "authorizing"
	e.errorMsg = ""
	e.exitResult = nil
	e.meter.StartSession()
	auth := e.auth
	selected := *e.selected
	settings := e.settings
	e.mu.Unlock()
	e.notifyTray()

	LogInfo("Starting VPN connection to %s (%s, %s)...", selected.Hostname, selected.CityName, selected.CountryName)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				LogError("Panic recovered in VPN connect goroutine: %v", r)
				e.setError(fmt.Sprintf("Internal connection error: %v", r))
			}
		}()

		if connectCtx.Err() != nil {
			LogInfo("Connection cancelled before initiation")
			return
		}

		// 1. Fetch Proxy Pass Token (with automatic activation & refresh)
		e.setConnectStep("authorizing")
		ctx, passCancel := context.WithTimeout(connectCtx, apiRequestTimeout)
		pass, err := e.getOrRenewProxyPass(ctx)
		passCancel()
		if err != nil {
			if connectCtx.Err() != nil {
				LogInfo("Connection aborted by user during pass fetch")
				return
			}
			e.setError(fmt.Sprintf("Failed to get proxy pass: %v", err))
			return
		}

		if connectCtx.Err() != nil {
			LogInfo("Connection aborted by user before upstream connect")
			return
		}

		// 2. Setup Upstream Controller
		e.setConnectStep("connecting_upstream")
		targetURL, _ := url.Parse(fmt.Sprintf("https://%s:%d", selected.Hostname, selected.Port))
		controller, err := newProxyController(e.guardian, targetURL, 30*time.Second, auth, pass, settings.UseH3, settings.ApiProxy, settings.TlsFragmentation, settings.IPProtocol)
		if err != nil {
			if connectCtx.Err() != nil {
				LogInfo("Connection aborted by user during controller setup")
				return
			}
			e.setError(fmt.Sprintf("Connection failed: %v", err))
			return
		}

		// Check cancellation before launching local proxies
		e.mu.RLock()
		cancelled := (e.state != StateConnecting || e.connectID != currentConnID)
		e.mu.RUnlock()
		if cancelled {
			LogInfo("Connection aborted by user; releasing controller...")
			_ = controller.Close()
			return
		}

		// Ensure safe, non-conflicting default ports (migrate away from 10808/10809 if set)
		socksPort := settings.SocksPort
		if socksPort <= 0 || socksPort == 10808 {
			socksPort = 20808
		}
		httpPort := settings.HttpPort
		if httpPort <= 0 || httpPort == 10809 {
			httpPort = 20809
		}

		// 3. Start Local SOCKS5 Proxy
		e.setConnectStep("starting_proxies")
		socksAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
		socksSrv := NewSocksProxyServer(socksAddr, controller, e.meter)
		socksSrv.SetSettings(settings)
		if err := socksSrv.Start(); err != nil {
			controller.Close()
			e.setError(fmt.Sprintf("Failed to start SOCKS5 proxy on %s: %v", socksAddr, err))
			return
		}

		// 4. Start Local HTTP Proxy
		httpAddr := fmt.Sprintf("127.0.0.1:%d", httpPort)
		httpSrv := NewHttpProxyServer(httpAddr, controller, e.meter)
		httpSrv.SetSettings(settings)
		if err := httpSrv.Start(); err != nil {
			socksSrv.Close()
			controller.Close()
			e.setError(fmt.Sprintf("Failed to start HTTP proxy on %s: %v", httpAddr, err))
			return
		}

		actualHttpAddr := httpSrv.listenAddr
		actualSocksAddr := socksSrv.listenAddr

		// Check cancellation before verification probe
		e.mu.RLock()
		cancelled = (e.state != StateConnecting || e.connectID != currentConnID)
		e.mu.RUnlock()
		if cancelled {
			LogInfo("Connection aborted by user; closing local proxies...")
			_ = socksSrv.Close()
			_ = httpSrv.Close()
			_ = controller.Close()
			return
		}

		// 5. Test real end-to-end connectivity through the tunnel before declaring Connected!
		e.setConnectStep("verifying_exit")
		LogInfo("Verifying tunnel connectivity and exit IP...")
		probe, err := VerifyExit(controller, 12*time.Second)
		if err != nil {
			if connectCtx.Err() != nil {
				_ = socksSrv.Close()
				_ = httpSrv.Close()
				_ = controller.Close()
				return
			}
			LogWarn("First exit probe failed (%v); retrying once...", err)
			time.Sleep(500 * time.Millisecond)
			probe, err = VerifyExit(controller, 8*time.Second)
		}
		if err != nil {
			// Clean up and report actual error - NEVER report connected when verification fails!
			_ = socksSrv.Close()
			_ = httpSrv.Close()
			_ = controller.Close()
			if connectCtx.Err() == nil {
				e.setError(fmt.Sprintf("Tunnel established but exit verification failed: %v", err))
			}
			return
		}

		LogInfo("Exit verified: IP %s, Country %s (latency: %dms)", probe.IP, probe.CountryCode, probe.LatencyMs)

		// Check cancellation before configuring system proxy or routing
		e.mu.RLock()
		cancelled = (e.state != StateConnecting || e.connectID != currentConnID)
		e.mu.RUnlock()
		if cancelled {
			LogInfo("Connection aborted by user before routing changes; rolling back...")
			_ = socksSrv.Close()
			_ = httpSrv.Close()
			_ = controller.Close()
			return
		}

		// 6. Configure Windows System Proxy and/or TUN Mode
		e.setConnectStep("configuring_routes")
		if settings.UseSystemProxy {
			if err := SetSystemProxyWithBypass(true, actualHttpAddr, actualSocksAddr, settings.BypassRules); err != nil {
				LogWarn("Could not set Windows system proxy: %v", err)
			} else {
				LogInfo("Windows system proxy activated (HTTP: %s, SOCKS5: %s)", actualHttpAddr, actualSocksAddr)
			}
		}

		if settings.EnableTunMode {
			if err := EnableTunMode(actualSocksAddr, actualHttpAddr, selected.Hostname, settings.IPProtocol); err != nil {
				LogWarn("Could not activate TUN mode: %v", err)
			} else {
				LogInfo("System TUN mode activated")
			}
		}

		// Final check and atomic state commit
		e.mu.Lock()
		if e.state != StateConnecting || e.connectID != currentConnID {
			e.mu.Unlock()
			LogInfo("Connection cancelled by user during finalization; rolling back everything...")
			_ = socksSrv.Close()
			_ = httpSrv.Close()
			_ = controller.Close()
			_ = DisableTunMode()
			_ = RestoreSystemProxy()
			return
		}

		e.controller = controller
		e.socksServer = socksSrv
		e.httpServer = httpSrv
		e.exitResult = &probe
		e.state = StateConnected
		e.connectStep = ""
		e.errorMsg = ""
		e.connectCancel = nil
		e.mu.Unlock()
		e.notifyTray()

		LogInfo("VPN Connected and verified successfully! Public IP: %s (%s)", probe.IP, probe.CountryCode)

		// Update quota info in background
		go e.refreshUserInfo()
	}()

	return nil
}

func (e *VpnEngine) Disconnect() error {
	e.mu.Lock()
	if e.state == StateDisconnecting {
		e.mu.Unlock()
		return nil
	}
	if e.state == StateDisconnected && e.socksServer == nil && e.httpServer == nil && e.controller == nil && !IsTunModeActive() {
		e.mu.Unlock()
		return nil
	}

	if e.connectCancel != nil {
		e.connectCancel()
		e.connectCancel = nil
	}
	e.state = StateDisconnecting
	e.connectStep = ""
	e.connectID = time.Now().UnixNano()
	socksSrv := e.socksServer
	httpSrv := e.httpServer
	ctrl := e.controller
	e.socksServer = nil
	e.httpServer = nil
	e.controller = nil
	useSysProxy := e.settings.UseSystemProxy
	e.mu.Unlock()
	e.notifyTray()

	LogInfo("Disconnecting VPN...")

	// Ensure LAN proxy listeners are stopped on VPN disconnect
	_ = globalLANShare.SetEnabled(false, nil, nil)

	// 1. Immediately close local proxy listeners and active upstream sessions to unblock all reads/writes
	if socksSrv != nil {
		_ = socksSrv.Close()
	}
	if httpSrv != nil {
		_ = httpSrv.Close()
	}
	if ctrl != nil {
		_ = ctrl.Close()
	}

	// 2. Tear down TUN mode routes, interface and netstack with bounded timeouts
	if IsTunModeActive() {
		_ = DisableTunMode()
		LogInfo("System TUN mode deactivated")
	}

	// 3. Restore Windows system proxy to direct mode
	if useSysProxy {
		_ = RestoreSystemProxy()
		LogInfo("Windows system proxy restored to direct")
	}

	// 4. Verify system routes and adapter state
	VerifyTunCleanup()

	e.mu.Lock()
	e.state = StateDisconnected
	e.connectStep = ""
	e.errorMsg = ""
	e.exitResult = nil
	if e.meter != nil {
		e.meter.StopSession()
		stats := e.meter.GetStats()
		if e.monthlyTracker != nil && (stats.BytesDown > 0 || stats.BytesUp > 0) {
			e.monthlyTracker.AddBytes(stats.BytesDown + stats.BytesUp)
		}
	}
	e.mu.Unlock()
	e.notifyTray()
	LogInfo("VPN Disconnected cleanly")
	return nil
}

func (e *VpnEngine) setError(msg string) {
	e.mu.Lock()
	e.state = StateError
	e.connectStep = ""
	e.errorMsg = msg
	e.mu.Unlock()
	e.notifyTray()
	LogError("%s", msg)
}

func (e *VpnEngine) Login(email, password, proxyAddr string) (need2FA bool, err error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if proxyAddr != "" {
		if err := configureAPIProxy(proxyAddr, e.settings.TlsFragmentation); err != nil {
			LogWarn("Setting API proxy failed: %v", err)
		} else {
			e.mu.Lock()
			e.settings.ApiProxy = proxyAddr
			e.mu.Unlock()
		}
	} else if e.settings.ApiProxy != "" {
		_ = configureAPIProxy(e.settings.ApiProxy, e.settings.TlsFragmentation)
	}

	LogInfo("Signing in as %s...", cleanEmail)
	// Give 120s budget for Fastly challenge solving & verification!
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	loginResp, err := vpnclient.FxaLogin(ctx, cleanEmail, password)
	if err != nil {
		LogError("Sign-in failed: %v", err)
		return false, err
	}

	if !loginResp.Verified {
		e.mu.Lock()
		e.pendingSessionToken = loginResp.SessionToken
		e.userEmail = cleanEmail
		e.mu.Unlock()
		LogInfo("Verification code sent to email %s", cleanEmail)
		return true, nil
	}

	// Session is verified, get OAuth token
	return false, e.completeSessionToken(loginResp.SessionToken, cleanEmail)
}

func (e *VpnEngine) Verify2FA(code string) error {
	e.mu.Lock()
	sessionToken := e.pendingSessionToken
	email := e.userEmail
	e.mu.Unlock()

	if sessionToken == "" {
		return fmt.Errorf("no pending 2FA session")
	}

	LogInfo("Submitting 2FA verification code...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := vpnclient.FxaVerifySession(ctx, sessionToken, strings.TrimSpace(code))
	if err != nil {
		LogError("2FA verification failed: %v", err)
		return err
	}

	return e.completeSessionToken(sessionToken, email)
}

func (e *VpnEngine) completeSessionToken(sessionToken, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	token, err := vpnclient.FxaOAuthToken(ctx, sessionToken)
	if err != nil {
		LogError("Failed to obtain OAuth token: %v", err)
		return err
	}

	_ = vpnclient.SaveTokensWithEmail(token, email)

	e.mu.Lock()
	e.auth = &runtimeAuth{Token: token, ObtainedAt: time.Now()}
	e.userEmail = email
	e.pendingSessionToken = ""
	e.mu.Unlock()

	LogInfo("Successfully signed in as %s", email)
	go func() {
		// Immediately activate Guardian for new or unactivated accounts
		actCtx, actCancel := context.WithTimeout(context.Background(), apiRequestTimeout)
		_, _ = vpnclient.ActivateGuardian(actCtx, e.guardian, token.AccessToken)
		actCancel()
		e.refreshUserInfo()
	}()
	return nil
}

func (e *VpnEngine) Logout() error {
	_ = e.Disconnect()

	home, err := os.UserHomeDir()
	if err == nil {
		_ = os.Remove(filepath.Join(home, ".firefox-vpn-tokens.json"))
	}

	e.mu.Lock()
	e.auth = nil
	e.userEmail = ""
	e.quota = nil
	e.mu.Unlock()

	LogInfo("Signed out")
	return nil
}

func (e *VpnEngine) GetSettings() VpnSettings {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.settings
}

func (e *VpnEngine) notifyTray() {
	if globalTray != nil {
		serverName := ""
		if e.selected != nil {
			serverName = e.selected.Hostname
		}
		globalTray.UpdateTrayStatus(e.state, serverName)
	}
}

func (e *VpnEngine) Reconnect() {
	LogInfo("Reconnection triggered...")
	_ = e.Disconnect()
	time.Sleep(150 * time.Millisecond)
	_ = e.Connect()
}

func (e *VpnEngine) SetTunMode(enable bool) {
	e.mu.Lock()
	e.settings.EnableTunMode = enable
	e.settings.UseSystemProxy = !enable
	s := e.settings
	state := e.state
	e.mu.Unlock()

	e.UpdateSettings(s)
	if state == StateConnected {
		go e.Reconnect()
	}
}

func (e *VpnEngine) UpdateSettings(s VpnSettings) {
	e.mu.Lock()
	oldProxy := e.settings.ApiProxy
	oldFrag := e.settings.TlsFragmentation
	e.settings = s
	httpSrv := e.httpServer
	socksSrv := e.socksServer
	e.mu.Unlock()

	_ = saveSettings(s)

	if s.ApiProxy != oldProxy || s.TlsFragmentation != oldFrag {
		_ = configureAPIProxy(s.ApiProxy, s.TlsFragmentation)
	}

	globalGamer.SetGamingMode(s.GamingMode)

	if httpSrv != nil {
		httpSrv.SetSettings(s)
	}
	if socksSrv != nil {
		socksSrv.SetSettings(s)
	}

	LogInfo("Settings updated (System Proxy: %v, TUN: %v, H3: %v, GamingMode: %v, SplitTunnel: %v, IranDirect: %v)",
		s.UseSystemProxy, s.EnableTunMode, s.UseH3, s.GamingMode, s.SplitTunnelEnabled, s.DirectIranTraffic)
}

func configureAPIProxy(raw string, tlsFrag bool) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		dialer := createCustomDialer()
		vpnclient.SetControlPlaneTransport(&http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				c, err := dialer.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				if tlsFrag && strings.HasPrefix(network, "tcp") {
					return newFragmentingConn(c, "sni"), nil
				}
				return c, nil
			},
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		})
		return nil
	}

	formatted := raw
	if !strings.Contains(formatted, "://") {
		if strings.Contains(formatted, ":1080") || strings.Contains(formatted, ":10808") {
			formatted = "socks5://" + formatted
		} else {
			formatted = "http://" + formatted
		}
	}

	parsed, err := url.Parse(formatted)
	if err != nil {
		return err
	}

	dialer := &net.Dialer{Timeout: 20 * time.Second}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		transport.Proxy = http.ProxyURL(parsed)
		transport.DialContext = dialer.DialContext
	case "socks5", "socks5h":
		var auth *proxy.Auth
		if parsed.User != nil {
			pwd, _ := parsed.User.Password()
			auth = &proxy.Auth{User: parsed.User.Username(), Password: pwd}
		}
		socksDialer, err := proxy.SOCKS5("tcp", parsed.Host, auth, dialer)
		if err != nil {
			return fmt.Errorf("creating socks5 dialer: %w", err)
		}
		contextDialer, ok := socksDialer.(proxy.ContextDialer)
		if !ok {
			return fmt.Errorf("socks5 dialer does not support context")
		}
		transport.DialContext = contextDialer.DialContext
	default:
		return fmt.Errorf("unsupported proxy scheme: %s", parsed.Scheme)
	}

	vpnclient.SetControlPlaneTransport(transport)
	LogInfo("API requests (Firefox Accounts / Guardian) now routed through proxy: %s", formatted)
	return nil
}

func createCustomDialer() *net.Dialer {
	return createProxyDialer(20 * time.Second)
}
