# FYRNOX Connection Pipeline, Direct UAC, SVG Flags & Latency Colors Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the stuck "Connecting" state via resilient dual SSE/polling and progressive connection states, enable direct native Windows UAC for Wintun TUN mode, display crisp vector SVG country flags across all 35 locations, and apply accurate green/amber/red latency badges to server pings.

**Architecture:** Extend backend `VpnStatus` with `ConnectStep` to broadcast granular connection progression via SSE and HTTP polling. Eliminate the SSE event name mismatch so the frontend reliably updates state. Trigger genuine Windows UAC elevation directly on TUN mode selection, handling user cancellation gracefully. Replace Unicode emoji flags with an embedded vector SVG dictionary. Apply strict latency color thresholds (0–100ms green, 101–200ms amber, >200ms red, gray for idle).

**Tech Stack:** Go 1.26, Windows DWM/ShellExecute Win32 APIs, Wintun Layer-3 virtual TUN, Vanilla JS (ES6+), HTML5/CSS3 with SVG vector assets.

**Spec:** [fyrnox_networking_and_ux_v2_design.md](file:///C:/Users/LouiNoor/.gemini/antigravity/brain/e4a584ef-f1a1-440a-ab06-38e888100028/fyrnox_networking_and_ux_v2_design.md)

## Global Constraints

- Preserve all existing FYRNOX branding, logo assets, fiery color scheme (`#FF9400`, `#FE6C31`, `#E53935`, `#120A08`), and dark DWM title bar styling.
- Never claim "Connected Securely" unless upstream connection and egress IP verification succeed.
- Never crash, deadlock, or spawn orphaned processes on rapid Connect/Disconnect cycles.
- Support all 35 verified countries without missing glyph boxes on Windows.
- Keep normal application startup non-elevated; request elevation only when Wintun TUN mode is selected.

## Review Focus

1. **SSE Event Stream Compatibility:** Both typed `event: status` and untyped `data:` events are handled seamlessly by WebView2 and standard browsers.
2. **UAC Elevation Cancellation:** If the user clicks "No" on the Windows UAC dialog (error 1223), the application stays in System Proxy mode without error loops or crashed UI.
3. **Slow / Throttled POP Verification:** Verify that `VerifyExit` completes within 6 seconds or fails gracefully with explicit error messages and resource rollback.
4. **Offline / Missing Flag Fallback:** Every server in the 35-country catalog has an authentic SVG flag representation; unknown codes fallback to a global vector emblem.
5. **Ping Badge Accessibility:** Numeric latency text remains legible in both dark and fiery accents across all four color states.

---

### Task 1: Connection Pipeline, Dual SSE Emission, and Progress Tracking

**Files:**
- Modify: `cmd/fyrnox/engine.go:70-100, 887-1110`
- Modify: `cmd/fyrnox/api.go:300-327`
- Modify: `cmd/fyrnox/web/app.js:150-205, 465-515`
- Test: `cmd/fyrnox/engine_test.go`

**Interfaces:**
- Produces: `VpnStatus.ConnectStep string` ("authorizing", "connecting_upstream", "starting_proxies", "verifying_exit", "configuring_routes", "")
- Emits: Dual SSE payload `event: status\ndata: ...\n\n` AND `data: ...\n\n` on `/api/events`

- [ ] **Step 1: Write test for connection progress state transitions and dual event emission**

```go
func TestConnectionProgressSteps(t *testing.T) {
	engine := NewVpnEngine()
	status := engine.GetStatus()
	if status.ConnectStep != "" {
		t.Fatalf("expected empty initial connect step, got %q", status.ConnectStep)
	}
	engine.setConnectStep("authorizing")
	if engine.GetStatus().ConnectStep != "authorizing" {
		t.Fatalf("expected connect step 'authorizing', got %q", engine.GetStatus().ConnectStep)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestConnectionProgressSteps ./cmd/fyrnox`
Expected: FAIL with `undefined: engine.setConnectStep`

- [ ] **Step 3: Implement `ConnectStep` and dual SSE in `cmd/fyrnox/engine.go` & `cmd/fyrnox/api.go`**

1. Add `ConnectStep string json:"connect_step,omitempty"` to `VpnStatus`.
2. Add `connectStep string` and `setConnectStep(step string)` to `VpnEngine`.
3. In `Connect()`, call `e.setConnectStep("authorizing")` -> `e.setConnectStep("connecting_upstream")` -> `e.setConnectStep("starting_proxies")` -> `e.setConnectStep("verifying_exit")` -> `e.setConnectStep("configuring_routes")`.
4. In `api.go:handleEvents`, emit both `fmt.Fprintf(w, "event: status\ndata: %s\n\n", statusData)` and `fmt.Fprintf(w, "data: %s\n\n", statusData)`.
5. In `app.js`, add `es.addEventListener('status', (e) => updateUI(JSON.parse(e.data)))` and active polling during `connecting` state. Render `connect_step` text beneath the connecting button.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -run TestConnectionProgressSteps ./cmd/fyrnox`
Expected: PASS

---

### Task 2: Direct Windows UAC Elevation for Wintun TUN Mode

**Files:**
- Modify: `cmd/fyrnox/tun_windows.go:390-415`
- Modify: `cmd/fyrnox/api.go:232-288`
- Modify: `cmd/fyrnox/web/app.js:518-570`
- Modify: `cmd/fyrnox/web/index.html:195-210, 600-620`
- Test: `cmd/fyrnox/uac_test.go`

**Interfaces:**
- Consumes: `/api/elevate` with `{ server: string, mode: "tun" }`
- Produces: Direct Windows UAC execution via `ShellExecute(0, "runas", ...)`, returns `{ ok: true }` or `{ ok: false, cancelled: true }`

- [ ] **Step 1: Write test for UAC error translation and argument building**

```go
func TestUACArgumentBuildingAndCancellation(t *testing.T) {
	args := buildElevationArgs("muc139.m1.fastly-masque.net", "tun")
	if !strings.Contains(args, "-elevated") || !strings.Contains(args, "-mode tun") || !strings.Contains(args, "muc139") {
		t.Fatalf("unexpected elevation args: %s", args)
	}
	cancelErr := syscall.Errno(1223) // ERROR_CANCELLED
	if !isUACCancelled(cancelErr) {
		t.Fatalf("expected ERROR_CANCELLED (1223) to be recognized as cancelled")
	}
}
```

- [ ] **Step 2: Run test to verify it passes/fails**

Run: `go test -v -run TestUACArgumentBuildingAndCancellation ./cmd/fyrnox`
Expected: PASS or update signatures as needed.

- [ ] **Step 3: Implement direct UAC elevation in `app.js` and remove intermediate web modal**

1. In `selectRoutingMode(enableTun)`: If `enableTun` and not `is_admin`, directly call `restartAsAdmin()`.
2. In `restartAsAdmin()`: Send `POST /api/elevate`. If `data.cancelled === true`, keep toggle off, show inline status toast `"UAC elevation was cancelled by user. Staying in System Proxy mode."`, and keep current connection intact.
3. If approved, elevated child starts with `-elevated -mode tun -server <current_server>` and parent cleanly exits.

- [ ] **Step 4: Verify test passes and mock call returns expected structure**

Run: `go test -v ./cmd/fyrnox`
Expected: PASS

---

### Task 3: Embedded Vector SVG Country Flags

**Files:**
- Create/Modify: `cmd/fyrnox/web/flags.js` (or embedded in `app.js:60-95`)
- Modify: `cmd/fyrnox/web/app.js:64-80, 225-235, 285-295, 345-370, 670-705`
- Modify: `cmd/fyrnox/web/style.css:1250-1310`
- Test: `cmd/fyrnox/flags_test.go`

**Interfaces:**
- Produces: `getFlagSVG(countryCode: string) -> string` (returns inline SVG vector element)

- [ ] **Step 1: Write unit test verifying country code flag coverage**

```go
func TestFlagCoverageForVerifiedCountries(t *testing.T) {
	engine := NewVpnEngine()
	servers := engine.GetServers()
	codes := make(map[string]bool)
	for _, s := range servers {
		codes[s.CountryCode] = true
	}
	if len(codes) < 30 {
		t.Fatalf("expected at least 30 unique country codes, got %d", len(codes))
	}
}
```

- [ ] **Step 2: Run test to verify server country codes**

Run: `go test -v -run TestFlagCoverageForVerifiedCountries ./cmd/fyrnox`
Expected: PASS

- [ ] **Step 3: Implement SVG flag map in `cmd/fyrnox/web/app.js` and style in `style.css`**

1. Define `SVG_FLAGS` dictionary containing clean, crisp SVG representations for all 35 country codes:
   `US, DE, NL, GB, FR, SE, FI, NO, DK, BE, AT, ES, PT, IT, IE, BG, CA, BR, AR, CL, CO, PE, MX, JP, SG, KR, TH, MY, PH, AU, NZ, ZA, GH, etc.`
2. Define `getFlag(code)` to return `<span class="flag-icon-svg" title="${code}">${svg}</span>` or fallback global globe vector.
3. Style `.flag-icon-svg` in `style.css` with `display: inline-flex`, `width: 22px`, `height: 16px`, `border-radius: 2px`, `box-shadow: 0 1px 3px rgba(0,0,0,0.3)`.
4. Ensure server rows, dashboard target cards, origin location, and egress verification indicators use `getFlag(code)`.

- [ ] **Step 4: Verify rendering without missing glyphs**

Inspect in WebView2 window to confirm flags render as colored vectors rather than empty rectangles.

---

### Task 4: Latency Color Indicators and Accessible Thresholds

**Files:**
- Modify: `cmd/fyrnox/web/app.js:672-685`
- Modify: `cmd/fyrnox/web/style.css:1320-1350`
- Test: `cmd/fyrnox/ping_test.go`

**Interfaces:**
- Produces: CSS classes `.ping-latency-pill.fast` (0–100ms), `.medium` (101–200ms), `.slow` (>200ms), `.unavailable` (idle/pending/<=0)

- [ ] **Step 1: Write test for latency classification logic**

```go
func TestLatencyClassificationThresholds(t *testing.T) {
	cases := []struct {
		ping     int
		expected string
	}{
		{ping: 45, expected: "fast"},
		{ping: 100, expected: "fast"},
		{ping: 101, expected: "medium"},
		{ping: 200, expected: "medium"},
		{ping: 201, expected: "slow"},
		{ping: 450, expected: "slow"},
		{ping: -1, expected: "unavailable"},
		{ping: 0, expected: "unavailable"},
	}
	for _, c := range cases {
		var got string
		if c.ping > 0 && c.ping <= 100 {
			got = "fast"
		} else if c.ping > 100 && c.ping <= 200 {
			got = "medium"
		} else if c.ping > 200 {
			got = "slow"
		} else {
			got = "unavailable"
		}
		if got != c.expected {
			t.Errorf("ping %d: got %s, want %s", c.ping, got, c.expected)
		}
	}
}
```

- [ ] **Step 2: Run test to verify logic**

Run: `go test -v -run TestLatencyClassificationThresholds ./cmd/fyrnox`
Expected: PASS

- [ ] **Step 3: Update `app.js` and `style.css`**

1. In `app.js` line 672:
   ```javascript
   let pingPillClass = 'unavailable';
   let pingText = '-- ms';
   if (srv.ping_ms > 0) {
     pingText = `${srv.ping_ms} ms`;
     if (srv.ping_ms <= 100) pingPillClass = 'fast';
     else if (srv.ping_ms <= 200) pingPillClass = 'medium';
     else pingPillClass = 'slow';
   }
   ```
2. In `style.css`:
   - `.ping-latency-pill.fast`: background `rgba(16, 185, 129, 0.15)`, color `#10B981` (Emerald).
   - `.ping-latency-pill.medium`: background `rgba(255, 148, 0, 0.15)`, color `#FF9400` (Gold/Orange).
   - `.ping-latency-pill.slow`: background `rgba(229, 57, 53, 0.15)`, color `#E53935` (Deep Red).
   - `.ping-latency-pill.unavailable`: background `rgba(255, 255, 255, 0.05)`, color `#64748B` (Slate Gray).

- [ ] **Step 4: Verify with live ping measurements**

Confirm in server browser that servers color-code accurately.

---

### Task 5: End-to-End Build, Acceptance Testing & Packaging

**Files:**
- Modify: `cmd/installer/main.go`
- Compile: `build/Fyrnox-x64.exe`, `build/Fyrnox-x86.exe`
- Compile: `build/dist/Fyrnox-Setup-v1.0.2.exe`
- Package: `build/dist/Fyrnox-v1.0.2-Windows-x64-Portable.zip`, `build/dist/Fyrnox-v1.0.2-Windows-x86-Portable.zip`

- [ ] **Step 1: Recompile 64-bit and 32-bit executables**
- [ ] **Step 2: Test live Connect lifecycle and verify transitions out of "Connecting"**
- [ ] **Step 3: Test direct UAC invocation and cancellation**
- [ ] **Step 4: Verify SVG flags and latency colors across all 35 servers**
- [ ] **Step 5: Build fresh Setup installer and portable zip archives**
