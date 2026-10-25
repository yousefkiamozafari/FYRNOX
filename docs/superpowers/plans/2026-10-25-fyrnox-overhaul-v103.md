# FYRNOX v1.0.3 Engineering Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the permanent Disconnect freeze, implement robust Windows L3 Wintun routing with proper adapter metrics, integrate the new high-resolution fiery cyber-fox emblem icon across all Windows binaries and UI, and refine the desktop interface to a clean, professional aesthetic.

**Architecture:** 
1. Fix gVisor teardown order (`closeDevice` before `tunStack.Close()`, bounded `tunStack.Wait()`) and replace PowerShell environment calls with direct Windows Registry API.
2. Query Wintun interface index dynamically and set interface metric to 1 with explicit `IF %IFINDEX%` routes to ensure complete system packet capture and eliminate DNS leaks.
3. Process `media_1792873823001_be8d5a62.jpg` into a multi-resolution rounded badge `.ico` (16 to 256px Lanczos) and recompile Windows `.syso` resources for both x64 and x86 targets.
4. Simplify UI dashboard to a focused, single-state telemetry view with dark title bar, crisp typography, and the official fiery color palette.

**Tech Stack:** Go 1.22+, Windows Win32 API / Registry, gVisor Netstack, Wintun Driver, Python 3 / Pillow (Lanczos icon generator), rsrc COFF linker, HTML5 / Vanilla CSS / JS.

**Spec:** `C:/Users/LouiNoor/.gemini/antigravity/brain/e4a584ef-f1a1-440a-ab06-38e888100028/fyrnox_overhaul_v3_design.md`

## Global Constraints
- Target OS: Windows 10/11 (AMD64 and 386).
- Zero simulated metrics or fake connection states.
- Communication with user must be in Persian (فارسی).
- Native dark mode title bar must be preserved via DWM API.
- All Disconnect actions must complete in under 500ms and restore system routes/proxy/DNS completely.

## Review Focus
1. Disconnect invoked while in `Connecting` state -> cleanly cancels pending token/TLS context and transitions to `Disconnected` without hanging.
2. Disconnect invoked during high active throughput -> force-closes TCP/UDP sockets immediately and restores network stack without gVisor deadlocking.
3. Portable mode on non-elevated user -> reports clear elevation required message rather than failing silently or hanging.
4. Icon rendering in small sizes (16x16, 24x24) -> sharp contrast and legible cyber fox contour without pixel blur.
5. High DPI display scaling (125%, 150%) -> UI dashboard remains sharp, aligned, and without clipping or horizontal scrollbars.

---

### Task 1: Fix Disconnect Engine Lifecycle & Eliminate Netstack Deadlock

**Files:**
- Modify: `cmd/fyrnox/tun_windows.go:320-365`
- Modify: `cmd/fyrnox/engine.go:1120-1170`
- Modify: `cmd/fyrnox/http_proxy.go:70-110`
- Modify: `cmd/fyrnox/socks_proxy.go:70-110`
- Test: `cmd/fyrnox/disconnect_test.go`

**Interfaces:**
- Consumes: `DisableTunMode() error`, `HttpProxyServer.Close() error`, `SocksProxyServer.Close() error`
- Produces: Fast, non-blocking teardown returning in < 200ms with 100% guarantee of restoring direct routing.

- [ ] **Step 1: Write unit test for rapid Disconnect cancellation**
Create `cmd/fyrnox/disconnect_test.go` testing that `Disconnect()` completes within 1000ms even if network endpoints are open.

- [ ] **Step 2: Run test to observe baseline failure or hang**
Run: `go test -v ./cmd/fyrnox -run TestDisconnectSpeed`
Expected: Verification of execution time.

- [ ] **Step 3: Fix gVisor teardown order in `tun_windows.go`**
In `DisableTunMode()`:
1. Close `tunDev` first via `closeDevice(tunDev)` so device read loops immediately hit EOF.
2. Call `tunStack.Close()`.
3. Run `tunStack.Wait()` in a goroutine with a 400ms timeout channel so it can never block the main caller.
4. Replace `powershell ... clearEnvScript` with direct `golang.org/x/sys/windows/registry` calls to `HKCU\Environment` and broadcast `WM_SETTINGCHANGE`.

- [ ] **Step 4: Update proxy connection tracking in `http_proxy.go` and `socks_proxy.go`**
Track active client net.Conn in a synchronized slice/map and close them explicitly upon `Close()`.

- [ ] **Step 5: Run tests and verify fast clean teardown**
Run: `go test -v ./cmd/fyrnox -run TestDisconnect`
Expected: PASS in < 200ms.

---

### Task 2: Robust Windows L3 Wintun Routing, Adapter Metric & DNS Leak Protection

**Files:**
- Modify: `cmd/fyrnox/tun_windows.go:180-260`
- Test: `cmd/fyrnox/routing_test.go`

**Interfaces:**
- Consumes: Wintun interface name "FyrnoxTun"
- Produces: Complete L3 default split routes bound to Wintun interface index with metric 1.

- [ ] **Step 1: Write routing configuration test**
Verify interface index resolution and route command argument builder.

- [ ] **Step 2: Implement dynamic IF index and metric 1 setting**
In `cmd/fyrnox/tun_windows.go`:
1. Find `net.InterfaceByName("FyrnoxTun")` to obtain exact interface index `iface.Index`.
2. Execute `netsh interface ipv4 set interface name="FyrnoxTun" metric=1`.
3. Add split routes specifying `IF %d`:
   `route add 0.0.0.0 mask 128.0.0.0 10.254.1.1 metric 1 if <ifIndex>`
   `route add 128.0.0.0 mask 128.0.0.0 10.254.1.1 metric 1 if <ifIndex>`
4. Set Wintun DNS to `1.1.1.1` and `1.0.0.1` to prevent DNS leakage.

- [ ] **Step 3: Verify routing table generation**
Run: `go test -v ./cmd/fyrnox -run TestRouting`
Expected: PASS.

---

### Task 3: High-Resolution Rounded Badge Custom Icon Processing & Windows Resource Compilation

**Files:**
- Create: `scripts/build_custom_icon.py`
- Source: `C:/Users/LouiNoor/.gemini/antigravity/brain/e4a584ef-f1a1-440a-ab06-38e888100028/.user_uploaded/media_1792873823001_be8d5a62.jpg`
- Generate: `icon.ico`, `icon.png`, `cmd/fyrnox/web/favicon.png`, `cmd/fyrnox/web/logo.svg`
- Update: `cmd/fyrnox/rsrc_windows_amd64.syso`, `cmd/fyrnox/rsrc_windows_386.syso`, `cmd/installer/rsrc_windows_*.syso`

**Interfaces:**
- Consumes: 1024x1024 JPEG emblem image.
- Produces: Professional rounded-badge ICO containing 256, 128, 64, 48, 32, 24, 16px layers, plus updated web icons and binary .syso object files.

- [ ] **Step 1: Create `scripts/build_custom_icon.py`**
Use Pillow to:
1. Open `media_1792873823001_be8d5a62.jpg`.
2. Apply high-precision rounded-corner mask on the dark badge border.
3. Resample using `Image.Resampling.LANCZOS` into:
   - `icon.ico` (256, 128, 64, 48, 32, 24, 16)
   - `icon.png` (512x512)
   - `cmd/fyrnox/web/favicon.png` (512x512)
   - `cmd/installer/payload/icon.ico`

- [ ] **Step 2: Run Python script to generate icon assets**
Run: `python scripts/build_custom_icon.py`
Expected: Creation of all icon files with verified byte sizes and resolutions.

- [ ] **Step 3: Recompile Windows .syso resources with `rsrc`**
Run:
`& "$env:GOPATH\bin\rsrc.exe" -ico icon.ico -arch amd64 -o cmd/fyrnox/rsrc_windows_amd64.syso`
`& "$env:GOPATH\bin\rsrc.exe" -ico icon.ico -arch 386 -o cmd/fyrnox/rsrc_windows_386.syso`
`& "$env:GOPATH\bin\rsrc.exe" -ico icon.ico -arch amd64 -o cmd/installer/rsrc_windows_amd64.syso`
`& "$env:GOPATH\bin\rsrc.exe" -ico icon.ico -arch 386 -o cmd/installer/rsrc_windows_386.syso`
`& "$env:GOPATH\bin\rsrc.exe" -ico icon.ico -arch amd64 -o cmd/uninstaller/rsrc_windows_amd64.syso`
Expected: Successful creation of all .syso files.

---

### Task 4: Professional UI Refinement & Trustworthy Telemetry Dashboard

**Files:**
- Modify: `cmd/fyrnox/web/index.html`
- Modify: `cmd/fyrnox/web/style.css`
- Modify: `cmd/fyrnox/web/app.js`

**Interfaces:**
- Consumes: Status SSE stream and REST endpoints (`/api/status`, `/api/servers`, `/api/connect`, `/api/disconnect`).
- Produces: Polished, clutter-free desktop UI adhering strictly to the official FYRNOX fiery obsidian palette.

- [ ] **Step 1: Simplify HTML structure in `index.html`**
- Replace legacy decorative containers with a streamlined dashboard:
  1. Header with custom fox emblem and clean version tag.
  2. Hero Connection Widget with state badge and clear action button.
  3. Active Endpoint Card (Flag, Country, City, Server IP, Latency with exact color threshold).
  4. Real Telemetry Grid: Actual Sent / Received bytes, Session Duration, Protocol (Fastly MASQUE / HTTP3).
- Ensure modal dialogs and server picker are clean and modern.

- [ ] **Step 2: Refine styling in `style.css`**
- Obsidian dark background `#120A08`, border `#2A1810`, primary accent `#FF9400`, warm gradient `#FE6C31`, error/disconnect `#E53935`.
- Remove unnecessary glowing borders and blurry shadows.
- Apply crisp typography with Segoe UI Variable / Segoe UI.

- [ ] **Step 3: Update `app.js` UI binding & fast Disconnect reaction**
- Update Disconnect handler to provide immediate visual feedback ("Disconnecting...") while allowing graceful timeout.
- Ensure state updates from SSE instantaneously transition UI to Disconnected without waiting for manual refresh.

---

### Task 5: End-to-End Build, Packaging & Multi-Stage Acceptance Verification

**Files:**
- Generate: `build/portable/Fyrnox-Portable/Fyrnox.exe` (x64)
- Generate: `build/portable-x86/Fyrnox-Portable/Fyrnox.exe` (x86)
- Generate: `build/installer/Fyrnox-Setup-v1.0.3.exe`
- Generate: `build/Fyrnox-v1.0.3-Windows-x64-Portable.zip`

- [ ] **Step 1: Compile x64 and x86 Fyrnox binaries**
Verify that the new `.syso` resources are properly linked and manifest is valid.

- [ ] **Step 2: Build Native Go Windows Installer**
Compile `cmd/installer` to produce `Fyrnox-Setup-v1.0.3.exe` containing x64 and x86 payloads.

- [ ] **Step 3: Run comprehensive verification suite**
1. Test Disconnect teardown speed (< 200ms).
2. Test Cancel while connecting.
3. Test Icon resource presence in generated binaries via PowerShell.
4. Test package integrity and executable version information (v1.0.3).
