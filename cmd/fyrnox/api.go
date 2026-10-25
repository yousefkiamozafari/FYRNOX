package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed web/*
var embeddedWebFS embed.FS

type ApiServer struct {
	engine *VpnEngine
	port   int
}

func NewApiServer(engine *VpnEngine, port int) *ApiServer {
	return &ApiServer{
		engine: engine,
		port:   port,
	}
}

func (s *ApiServer) Start() error {
	mux := http.NewServeMux()

	// REST API Endpoints
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/servers", s.handleServers)
	mux.HandleFunc("/api/select-server", s.handleSelectServer)
	mux.HandleFunc("/api/connect", s.handleConnect)
	mux.HandleFunc("/api/disconnect", s.handleDisconnect)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/verify", s.handleVerify)
	mux.HandleFunc("/api/verify-2fa", s.handleVerify)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/api/logs", s.handleLogs)
	mux.HandleFunc("/api/logs/clear", s.handleClearLogs)
	mux.HandleFunc("/api/ping-all", s.handlePingAll)
	mux.HandleFunc("/api/ping", s.handlePingServer)
	mux.HandleFunc("/api/restart-admin", s.handleRestartAdmin)
	mux.HandleFunc("/api/elevate", s.handleRestartAdmin)
	mux.HandleFunc("/api/ip", s.handleIP)
	mux.HandleFunc("/api/ip/refresh", s.handleIPRefresh)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/apps/running", s.handleRunningApps)
	mux.HandleFunc("/api/apps/installed", s.handleInstalledApps)
	mux.HandleFunc("/api/gaming-mode", s.handleGamingMode)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/window/drag", s.handleWindowDrag)
	mux.HandleFunc("/api/window/minimize", s.handleWindowMinimize)
	mux.HandleFunc("/api/window/maximize", s.handleWindowMaximize)
	mux.HandleFunc("/api/window/close", s.handleWindowClose)

	// Static UI
	subFS, err := fs.Sub(embeddedWebFS, "web")
	if err != nil {
		return fmt.Errorf("embedded web fs error: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	listenAddr := fmt.Sprintf("127.0.0.1:%d", s.port)
	LogInfo("Web dashboard running on http://%s", listenAddr)

	go func() {
		if err := http.ListenAndServe(listenAddr, mux); err != nil {
			LogError("HTTP dashboard error: %v", err)
		}
	}()

	return nil
}

func (s *ApiServer) jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *ApiServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, s.engine.GetStatus())
}

func (s *ApiServer) handleServers(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, s.engine.GetServers())
}

func (s *ApiServer) handleSelectServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Hostname string `json:"hostname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}
	if err := s.engine.SelectServer(req.Hostname); err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}
	s.jsonResponse(w, map[string]any{"ok": true, "status": s.engine.GetStatus()})
}

func (s *ApiServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.engine.Connect(); err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}
	s.jsonResponse(w, map[string]any{"ok": true, "status": s.engine.GetStatus()})
}

func (s *ApiServer) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = s.engine.Disconnect()
	s.jsonResponse(w, map[string]any{"ok": true, "status": s.engine.GetStatus()})
}

func (s *ApiServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Proxy    string `json:"proxy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}

	need2FA, err := s.engine.Login(strings.TrimSpace(req.Email), req.Password, strings.TrimSpace(req.Proxy))
	if err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}

	s.jsonResponse(w, map[string]any{
		"ok":       true,
		"need_2fa": need2FA,
		"status":   s.engine.GetStatus(),
	})
}

func (s *ApiServer) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}

	if err := s.engine.Verify2FA(strings.TrimSpace(req.Code)); err != nil {
		s.jsonResponse(w, map[string]any{"error": err.Error()})
		return
	}

	s.jsonResponse(w, map[string]any{"ok": true, "status": s.engine.GetStatus()})
}

func (s *ApiServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	_ = s.engine.Logout()
	s.jsonResponse(w, map[string]any{"ok": true, "status": s.engine.GetStatus()})
}

func (s *ApiServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, GetRecentLogs())
}

func (s *ApiServer) handleClearLogs(w http.ResponseWriter, r *http.Request) {
	ClearLogs()
	s.jsonResponse(w, map[string]any{"ok": true})
}

func (s *ApiServer) handlePingAll(w http.ResponseWriter, r *http.Request) {
	servers := s.engine.PingAllServers()
	s.jsonResponse(w, servers)
}

func (s *ApiServer) handlePingServer(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")
	status := s.engine.GetStatus()

	// If connected, measure verified end-to-end round-trip latency through the active tunnel
	if status.State == StateConnected && (hostname == "" || r.URL.Query().Get("tunnel") == "true" || (status.Server != nil && hostname == status.Server.Hostname)) {
		latency, err := s.engine.PingTunnel()
		if err != nil {
			s.jsonResponse(w, map[string]any{
				"ok":        false,
				"tunnel":    true,
				"error":     err.Error(),
				"ping":      -1,
				"timestamp": time.Now().Format("15:04:05"),
			})
			return
		}
		s.jsonResponse(w, map[string]any{
			"ok":        true,
			"tunnel":    true,
			"hostname":  func() string { if status.Server != nil { return status.Server.Hostname }; return "" }(),
			"ping":      latency,
			"ping_ms":   latency,
			"timestamp": time.Now().Format("15:04:05"),
		})
		return
	}

	// Disconnected state: measure direct handshake endpoint latency
	ping := s.engine.PingServer(hostname)
	s.jsonResponse(w, map[string]any{
		"ok":        ping > 0,
		"tunnel":    false,
		"hostname":  hostname,
		"ping":      ping,
		"ping_ms":   ping,
		"timestamp": time.Now().Format("15:04:05"),
	})
}

func (s *ApiServer) handleRestartAdmin(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Query().Get("server")
	mode := r.URL.Query().Get("mode")
	if r.Method == http.MethodPost && r.Body != nil {
		var req struct {
			Server string `json:"server"`
			Mode   string `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Server != "" {
				server = req.Server
			}
			if req.Mode != "" {
				mode = req.Mode
			}
		}
	}

	if server == "" {
		st := s.engine.GetStatus()
		if st.Server != nil {
			server = st.Server.Hostname
		}
	}
	if mode == "" {
		mode = "tun"
	}

	err := RestartAsAdminWithParams(server, mode)
	if err != nil {
		if isUACCancelled(err) {
			LogInfo("UAC elevation prompt was cancelled by user")
			s.jsonResponse(w, map[string]any{
				"ok":        false,
				"cancelled": true,
				"error":     "UAC elevation prompt was cancelled by user",
			})
			return
		}
		LogError("Failed to restart as admin: %v", err)
		s.jsonResponse(w, map[string]any{
			"ok":        false,
			"cancelled": false,
			"error":     err.Error(),
		})
		return
	}

	LogInfo("Successfully spawned elevated admin instance; exiting unprivileged instance...")
	s.jsonResponse(w, map[string]any{"ok": true})

	// Graceful shutdown after letting response flush
	go func() {
		time.Sleep(600 * time.Millisecond)
		os.Exit(0)
	}()
}

func (s *ApiServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req VpnSettings
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			s.engine.UpdateSettings(req)
		}
	}
	s.jsonResponse(w, s.engine.GetStatus().Settings)
}

// handleEvents implements Server-Sent Events (SSE) for live, instant UI updates
func (s *ApiServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			statusData, _ := json.Marshal(s.engine.GetStatus())
			fmt.Fprintf(w, "event: status\ndata: %s\n\n", statusData)
			fmt.Fprintf(w, "data: %s\n\n", statusData)
			flusher.Flush()
		}
	}
}

func (s *ApiServer) handleIP(w http.ResponseWriter, r *http.Request) {
	info := s.engine.RefreshUserIP()
	s.jsonResponse(w, info)
}

func (s *ApiServer) handleIPRefresh(w http.ResponseWriter, r *http.Request) {
	info := s.engine.RefreshUserIP()
	s.jsonResponse(w, info)
}

func (s *ApiServer) handleWindowDrag(w http.ResponseWriter, r *http.Request) {
	handleWindowDrag(0)
	s.jsonResponse(w, map[string]any{"ok": true})
}

func (s *ApiServer) handleWindowMinimize(w http.ResponseWriter, r *http.Request) {
	handleWindowMinimize(0)
	s.jsonResponse(w, map[string]any{"ok": true})
}

func (s *ApiServer) handleWindowMaximize(w http.ResponseWriter, r *http.Request) {
	handleWindowMaximize(0)
	s.jsonResponse(w, map[string]any{"ok": true})
}

func (s *ApiServer) handleWindowClose(w http.ResponseWriter, r *http.Request) {
	handleWindowClose(0)
	s.jsonResponse(w, map[string]any{"ok": true})
}

func (s *ApiServer) handleRunningApps(w http.ResponseWriter, r *http.Request) {
	apps := GetRunningProcesses()
	s.jsonResponse(w, apps)
}

func (s *ApiServer) handleInstalledApps(w http.ResponseWriter, r *http.Request) {
	apps := GetInstalledApplications()
	s.jsonResponse(w, apps)
}

func (s *ApiServer) handleGamingMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			Active *bool `json:"active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Active != nil {
			globalGamer.SetGamingMode(*req.Active)
			settings := s.engine.GetSettings()
			settings.GamingMode = *req.Active
			s.engine.UpdateSettings(settings)
		} else {
			// Toggle if no explicit flag specified
			stats := globalGamer.GetStats()
			newVal := !stats.Active
			globalGamer.SetGamingMode(newVal)
			settings := s.engine.GetSettings()
			settings.GamingMode = newVal
			s.engine.UpdateSettings(settings)
		}
	}
	s.jsonResponse(w, globalGamer.GetStats())
}

