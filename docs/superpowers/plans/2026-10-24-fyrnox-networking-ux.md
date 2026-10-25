# FYRNOX Critical Networking Corrections and Native Windows UX Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform FYRNOX into a genuine, native, robust Windows desktop VPN client with verified Anycast exit disclosure, accurate multi-sample TCP latency measurement, zero visible CMD console windows, seamless UAC elevation for Wintun Layer-3 TUN mode, and an original dark-graphite Locations UI.

**Architecture:** 
- Core Go networking layer (`cmd/fyrnox/`) with Fastly MASQUE HTTP/2 duplex CONNECT and optional Wintun L3 packet routing via `tun2socks` + `gVisor`.
- Pure embedded WebView2 desktop window with zero external browser processes or browser chrome.
- Win32 API integration (`CREATE_NO_WINDOW (0x08000000)` and `HideWindow: true`) to permanently eliminate visible console popups.
- Standard Windows UAC elevation handoff via `ShellExecuteW(runas)` passing `--elevated --mode=tun --server=<hostname>` without duplicate instances.
- Multi-sample median TCP SYN latency probing without DNS overhead; separation of lightweight tunnel TCP RTT from heavy exit IP verification.
- Modern dark-graphite UI (`#0B0F17` / `#06B6D4`) with honest metadata: `Selected Target Location` vs `Actual Exit IP` vs `Observed Exit Country`.

**Tech Stack:** Go 1.26, `github.com/jchv/go-webview2`, `github.com/xjasonlyu/tun2socks/v2`, `golang.org/x/sys/windows`, HTML5/CSS3/Vanilla JS (embedded).

**Spec:** User Requirements for FYRNOX Priorities 1 through 7.

---

## Global Constraints

- Work strictly inside the existing repository (`FoxyVPN-Windows`).
- Do not fabricate ping, exit IP, geolocation, traffic statistics, or server availability.
- When Fastly Anycast routes traffic to a European edge node (e.g. NL/DE for US Anycast), display the actual observed exit IP and country honestly; do not claim it is in the US if the egress node is in the Netherlands.
- Maintain compatibility with Go 1.26 and Windows 10/11 x64.
- Zero console window flashing during connect, disconnect, or route updates.
- Keep standard System Proxy mode operational without requiring administrator privileges.

---

## Review Focus

1. **UAC Cancellation:** User denies the UAC prompt when switching to TUN mode -> App must catch `ERROR_CANCELLED (1223)`, preserve state, remain in non-elevated proxy mode, and display a helpful notice without crashing.
2. **Fastly Anycast Geolocation Discrepancy:** Selecting `US Anycast` terminates at an EU POP -> UI displays `Target: United States (Anycast Edge)`, `Exit IP: <Fastly EU IP>`, and `Observed Exit Country: Netherlands (NL)` clearly separated.
3. **Tunnel Ping Latency Overhead:** Telemetry polling must never trigger a full HTTPS TLS handshake with Cloudflare; it must measure pure TCP stream RTT through the established tunnel.
4. **Child Process Window Flashing:** All subprocess calls (`route`, `netsh`, `powershell`) must strictly use `SysProcAttr: {HideWindow: true, CreationFlags: 0x08000000}`.
5. **Clean Elevation Handoff:** Launching the elevated instance must cleanly close the unprivileged instance and pass the user's selected server and TUN mode seamlessly without an infinite elevation loop.

---

## Task Decomposition

### Task 1: Subprocess Concealment & Zero-Console Execution (Priority 4)

**Files:**
- Modify: `cmd/fyrnox/tun_windows.go`
- Modify: `cmd/fyrnox/main.go`
- Test: `cmd/fyrnox/tun_windows_test.go`

**Interfaces:**
- Produces: `runSilentCmd(name string, args ...string) (string, error)` configured with `CREATE_NO_WINDOW (0x08000000)` and `HideWindow: true`.
- Consumes: Standard `os/exec` and `syscall`.

- [ ] **Step 1: Write unit test verifying process creation flags**
  Create `cmd/fyrnox/tun_windows_test.go` asserting that `newSilentCmd` sets `HideWindow == true` and `CreationFlags == 0x08000000`.

- [ ] **Step 2: Run test to confirm behavior**
  `go test -v ./cmd/fyrnox -run TestSilentCmd`

- [ ] **Step 3: Implement `runSilentCmd` in `tun_windows.go`**
  Replace all raw `exec.Command` invocations in `tun_windows.go` (`route`, `netsh`, `powershell`) and `main.go` (`taskkill`) with `runSilentCmd`.

- [ ] **Step 4: Verify that Connect and Disconnect trigger zero visible console windows**

---

### Task 2: Multi-Sample TCP Latency & Separation of Tunnel RTT (Priority 2)

**Files:**
- Modify: `cmd/fyrnox/engine.go`
- Modify: `cmd/fyrnox/api.go`
- Modify: `cmd/fyrnox/core.go`
- Test: `cmd/fyrnox/ping_test.go`

**Interfaces:**
- Produces: `measureServerPing(ctx context.Context, s ServerLocation) int` (3-sample median TCP SYN on port 2499/443 without DNS lookup overhead).
- Produces: `PingTunnelFast(ctrl *VpnController) (int64, error)` (pure TCP stream RTT through existing MASQUE session).
- Modifies: `VerifyExit` to run once on connection or on manual refresh, not inside the 2s telemetry loop.

- [ ] **Step 1: Write unit test for median TCP latency calculation**
  Create `cmd/fyrnox/ping_test.go` testing median computation from multiple TCP latency samples.

- [ ] **Step 2: Run test to verify it passes**
  `go test -v ./cmd/fyrnox -run TestMedianCalculation`

- [ ] **Step 3: Update `measureServerPing` in `engine.go`**
  Perform 3 TCP SYN dial probes to the resolved IP (or fallback hostname), calculate the median, and return accurate RTT.

- [ ] **Step 4: Implement lightweight `PingTunnelFast`**
  Dial a quick TCP connect (`1.1.1.1:80` or `8.8.8.8:53`) via `opener.OpenTunnel()` timing only stream establishment, avoiding full HTTPS body retrieval.

- [ ] **Step 5: Verify via `/api/ping` endpoint**
  Ensure response returns sub-150ms real RTT instead of 1000ms+ HTTPS handshake overhead.

---

### Task 3: Real Windows UAC Elevation & Clean Handoff for TUN Mode (Priority 3)

**Files:**
- Modify: `cmd/fyrnox/tun_windows.go`
- Modify: `cmd/fyrnox/main.go`
- Modify: `cmd/fyrnox/api.go`
- Modify: `cmd/fyrnox/web/app.js`

**Interfaces:**
- Produces: `RestartAsAdminWithParams(serverHostname string, mode string) error` using `windows.ShellExecuteW` with `runas` verb.
- Consumes: Windows `ShellExecuteW` returning `ERROR_CANCELLED (1223)`.
- Produces: Command line flags in `main.go`: `-elevated`, `-server`, `-mode`.

- [ ] **Step 1: Add command line flags in `main.go`**
  Support `-elevated`, `-server <hostname>`, and `-mode <tun|proxy>`. If `-elevated` is set and user is Admin, auto-select server and activate TUN mode.

- [ ] **Step 2: Update `RestartAsAdmin` in `tun_windows.go`**
  Build parameter string `fmt.Sprintf("-elevated -server %s -mode %s", server, mode)`. If `err == windows.ERROR_CANCELLED`, return typed cancellation error.

- [ ] **Step 3: Update `/api/restart-admin` in `api.go`**
  Accept server and mode parameters. Return `{ok: false, cancelled: true}` on user cancellation. On success, schedule clean process shutdown of the unprivileged instance after 500ms.

- [ ] **Step 4: Update frontend `app.js`**
  Display a concise modal when TUN mode is toggled, explaining why administrator rights are needed for Wintun. If user confirms, trigger `/api/restart-admin`. If cancelled, restore switch to Proxy mode gracefully without errors.

---

### Task 4: Server List Anycast & Exit Country Disclosure (Priority 1)

**Files:**
- Modify: `cmd/fyrnox/engine.go`
- Modify: `cmd/fyrnox/serverlist.go`
- Modify: `cmd/fyrnox/api.go`
- Modify: `cmd/fyrnox/web/index.html`
- Modify: `cmd/fyrnox/web/app.js`

**Interfaces:**
- Produces: Extended `ServerLocation` with `IsAnycast: bool` and `TargetRegion: string`.
- Produces: Complete verified Mozilla Remote Settings list of 34 regional servers (including `muc139` DE, `eglc860` GB, `lfpb115` FR, `ehrd229` NL, `rjtf770` JP, `wsat188` SG, etc.).
- Produces: UI display distinguishing:
  1. `Selected Server`: Country + City / Anycast Edge
  2. `Actual Public Exit IP`: e.g. `151.101.x.x`
  3. `Verified Exit Country`: Real geolocated country code and name.

- [ ] **Step 1: Update `ServerLocation` struct and server definitions**
  Mark `us.m1.fastly-masque.net` and `p.m1.fastly-masque.net` as `IsAnycast: true, CityName: "Anycast Edge"`.

- [ ] **Step 2: Synchronize official Mozilla Remote Settings server definitions**
  Include all 34 verified Fastly MASQUE endpoints from Mozilla's official collection in `defaultServers`.

- [ ] **Step 3: Update `VpnStatus` and frontend UI**
  Show separate clear badges for `Selected Target`, `Actual Exit IP`, and `Observed Exit Country`.

---

### Task 5: Premium Dark Graphite UI & Locations Redesign (Priority 5 & 6)

**Files:**
- Modify: `cmd/fyrnox/web/style.css`
- Modify: `cmd/fyrnox/web/index.html`
- Modify: `cmd/fyrnox/web/app.js`

**Interfaces:**
- Modern dark graphite theme: `--bg-base: #0B0F17`, `--bg-surface: #101522`, `--border-subtle: rgba(255, 255, 255, 0.08)`, `--accent-cyan: #06B6D4`.
- Reorganized Locations screen:
  - Search & filter bar (instant typing filter).
  - Country grouping with accurate SVG flags and city names.
  - Latency badges with real-time median ping.
  - Dedicated Anycast vs POP indicator.
  - Accurate, minimalist world map with exact coordinates for each server location.
  - Zero fake sparklines, fake graphs, or fabricated stats.

- [ ] **Step 1: Refactor CSS color variables and typography in `style.css`**
  Apply `#0B0F17` / `#101522` / `#06B6D4` palette with refined typography and smooth transitions.

- [ ] **Step 2: Redesign the Locations Grid in `index.html` and `app.js`**
  Implement clean card layout with flag, country name, city code, server hostname, and real-time ping pill.

- [ ] **Step 3: Update SVG World Map with precise geographic coordinates**
  Map ICAO airport codes to their verified latitudes/longitudes.

---

### Task 6: Build, Verification & Acceptance Testing (Priority 7)

**Files:**
- Build: `build/portable/Fyrnox-Portable/Fyrnox.exe`
- Artifact: `walkthrough.md`

- [ ] **Step 1: Compile portable executable**
  Build with Go 1.26: `go build -ldflags "-H windowsgui -s -w" -o build/portable/Fyrnox-Portable/Fyrnox.exe ./cmd/fyrnox`

- [ ] **Step 2: Execute all 10 acceptance tests**
  1. US Endpoint selection & actual exit IP/country observation.
  2. Alternative endpoint selection (e.g. Germany/UK/Japan) and egress verification.
  3. Direct TCP RTT vs tunnel RTT multi-sample verification.
  4. TUN activation with UAC approval.
  5. UAC cancellation handling & non-admin proxy mode fallback.
  6. Multi-application traffic routing.
  7. DNS and endpoint route protection without routing loop.
  8. Repeated connect/disconnect with zero visible CMD windows.
  9. Native embedded WebView2 launch without external browser.
  10. Testing standalone portable distribution.

- [ ] **Step 3: Document test matrix with PASS/FAIL/NOT TESTED in Walkthrough**

---

## Verification Plan

### Automated Tests
```powershell
$env:GOCACHE='C:\Users\LouiNoor\.gemini\antigravity\scratch\gocache'
$env:GOPATH='C:\Users\LouiNoor\.gemini\antigravity\scratch\gopath'
cd C:\Users\LouiNoor\.gemini\antigravity\scratch\FoxyVPN-Windows
go test -v ./cmd/fyrnox
```

### Manual Acceptance Testing
1. Launch `build/portable/Fyrnox-Portable/Fyrnox.exe`.
2. Confirm window opens directly as embedded WebView2 without external Edge/Chrome.
3. Select "United States": Verify UI clearly states `Selected: United States (Anycast Edge)`, `Exit IP: <Fastly IP>`, and `Observed Country: Netherlands/Germany`.
4. Select "Germany (Munich)" or "United Kingdom (London)": Verify exit IP and country match Germany / UK.
5. Click "Ping All": Verify measured pings are real, responsive (under 200ms where reachable), and multi-sampled.
6. Toggle TUN Mode: Confirm UAC dialog appears with application name and path; upon approval, elevated instance launches and connects in TUN mode.
7. Test Connect / Disconnect 5 times in succession: Verify ZERO black console windows flash on screen.
