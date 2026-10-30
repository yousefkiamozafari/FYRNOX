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
	mux.HandleFunc("/api/settings/reset", s.handleSettingsReset)
	mux.HandleFunc("/api/routing-mode", s.handleRoutingMode)
	mux.HandleFunc("/api/apps/running", s.handleRunningApps)
	mux.HandleFunc("/api/apps/installed", s.handleInstalledApps)
	mux.HandleFunc("/api/gaming-mode", s.handleGamingMode)
	mux.HandleFunc("/api/events", s.handleEvents)
	mux.HandleFunc("/api/window/drag", s.handleWindowDrag)
	mux.HandleFunc("/api/window/minimize", s.handleWindowMinimize)
	mux.HandleFunc("/api/window/maximize", s.handleWindowMaximize)
	mux.HandleFunc("/api/window/close", s.handleWindowClose)
	mux.HandleFunc("/api/lan-share", s.handleLANShare)
	mux.HandleFunc("/api/service-matrix", s.handleServiceMatrix)
	mux.HandleFunc("/api/service-matrix/check", s.handleServiceMatrixCheck)
	mux.HandleFunc("/api/panic/toggle", s.handlePanicToggle)
	mux.HandleFunc("/api/panic/restore", s.handlePanicRestore)
	mux.HandleFunc("/api/panic/status", s.handlePanicStatus)
	mux.HandleFunc("/api/game-qos", s.handleGameQoS)

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
		if err == nil && latency > 0 {
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
	}

	// Disconnected state or fallback: measure direct TCP ping to server endpoint
	targetHost := hostname
	if targetHost == "" && status.Server != nil {
		targetHost = status.Server.Hostname
	}
	ping := s.engine.PingServer(targetHost)
	s.jsonResponse(w, map[string]any{
		"ok":        ping > 0,
		"tunnel":    false,
		"hostname":  targetHost,
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
			oldTun := s.engine.GetSettings().EnableTunMode
			s.engine.UpdateSettings(req)
			if oldTun != req.EnableTunMode {
				s.engine.SetTunMode(req.EnableTunMode)
			}
		}
	}
	s.jsonResponse(w, s.engine.GetStatus().Settings)
}

func (s *ApiServer) handleRoutingMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		EnableTun bool `json:"enable_tun"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonResponse(w, map[string]any{"ok": false, "error": "Invalid JSON payload"})
		return
	}
	s.engine.SetTunMode(req.EnableTun)
	s.jsonResponse(w, map[string]any{"ok": true, "enable_tun": req.EnableTun})
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



func (s *ApiServer) handleLANShare(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonResponse(w, map[string]any{"ok": false, "error": "Invalid JSON payload"})
			return
		}
		if req.Enabled {
			status := s.engine.GetStatus()
			if status.State != StateConnected {
				s.jsonResponse(w, map[string]any{"ok": false, "error": "VPN must be connected before enabling LAN sharing"})
				return
			}
			opener := s.engine.GetTunnelOpener()
			if opener == nil {
				s.jsonResponse(w, map[string]any{"ok": false, "error": "VPN tunnel is not ready"})
				return
			}
			if err := globalLANShare.SetEnabled(true, opener, s.engine.meter); err != nil {
				s.jsonResponse(w, map[string]any{"ok": false, "error": err.Error()})
				return
			}
		} else {
			_ = globalLANShare.SetEnabled(false, nil, nil)
		}
	}
	s.jsonResponse(w, globalLANShare.GetStatus())
}

func (s *ApiServer) handleServiceMatrix(w http.ResponseWriter, r *http.Request) {
	matrix := globalServiceMatrix.GetMatrix(s.engine)
	s.jsonResponse(w, matrix)
}

func (s *ApiServer) handleServiceMatrixCheck(w http.ResponseWriter, r *http.Request) {
	matrix := globalServiceMatrix.Refresh(s.engine)
	s.jsonResponse(w, matrix)
}

func (s *ApiServer) handlePanicToggle(w http.ResponseWriter, r *http.Request) {
	globalPanic.TogglePanic()
	s.jsonResponse(w, map[string]interface{}{
		"panic":  globalPanic.IsPanicActive(),
		"hotkey": "Ctrl+Alt+Shift+X",
	})
}

func (s *ApiServer) handlePanicRestore(w http.ResponseWriter, r *http.Request) {
	globalPanic.RestorePanic()
	s.jsonResponse(w, map[string]interface{}{
		"panic":  false,
		"hotkey": "Ctrl+Alt+Shift+X",
	})
}

func (s *ApiServer) handlePanicStatus(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]interface{}{
		"panic":  globalPanic.IsPanicActive(),
		"hotkey": "Ctrl+Alt+Shift+X",
	})
}

func (s *ApiServer) handleGameQoS(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			Add    string `json:"add"`
			Remove string `json:"remove"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Add != "" {
				globalPingShield.AddCustomGameSignature(req.Add)
			}
			if req.Remove != "" {
				globalPingShield.RemoveCustomGameSignature(req.Remove)
			}
		}
	}
	stats := globalGamer.GetStats()
	s.jsonResponse(w, map[string]interface{}{
		"game_detected": stats.GameDetected,
		"detected_game": stats.DetectedGame,
		"signatures":    globalPingShield.GetGameSignatures(),
	})
}

func (s *ApiServer) handleSettingsReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	newSettings := s.engine.ResetSettingsToDefault()
	s.jsonResponse(w, map[string]any{
		"status":   "ok",
		"message":  "Settings restored to defaults",
		"settings": newSettings,
	})
}
