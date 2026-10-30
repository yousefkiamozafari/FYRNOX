# FYRNOX Atomic Connection Pipeline, 3D Globe Spherical True-Scale Geometry, and Settings Account State Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the FYRNOX VPN connection pipeline into a fast, atomic linear state machine in Go, eliminate the JavaScript reference error freezing the UI, reconstruct the 3D globe with true-scale geographic geometry via spherical Fibonacci sampling, and dynamically synchronize the Firefox Account state in Settings between Sign In and Sign Out.

**Architecture:** 
1. **Frontend Core:** Eliminate the fatal `ReferenceError: isTun is not defined` in `app.js`, wrap `updateUI()` in defensive handling, dynamically bind the selected server city (e.g. `SPJC · Lima` for Peru), and decouple the Settings account row from static `data-i18n` tags so it displays `yousefkiamozafari@gmail.com` and a red/amber "Sign Out" button when logged in.
2. **Backend Engine:** Refactor `VpnEngine.Connect()` and `Disconnect()` in `cmd/fyrnox/engine.go` into an atomic linear fail-fast pipeline: pre-flight check → proxy-pass token → MASQUE HTTP/2 dial with TLS fragmentation → local SOCKS5/HTTP listeners → fast parallel exit probe (under 3.5s) → atomic WinINet proxy commit and state transition.
3. **WebGL 3D Globe:** Replace the 720×360 equirectangular 2D raster scan in `cmd/fyrnox/web/globe.js` with an equal-area spherical Fibonacci lattice (~18,000 points) and direct vector Cartesian projection of Natural Earth GeoJSON boundaries, guaranteeing undistorted continent contours and true area proportions.

**Tech Stack:** Go 1.26, Three.js r128, Vanilla ES6+ JavaScript, Windows WinINet & Win32 APIs, HTML5/CSS3.

**Spec:** [`docs/superpowers/specs/2026-10-30-connection-core-globe-and-account-design.md`](file:///C:/Users/LouiNoor/.gemini/antigravity/scratch/FoxyVPN-Windows/docs/superpowers/specs/2026-10-30-connection-core-globe-and-account-design.md)

---

## Global Constraints

- Never commit or force-push to git (work strictly within working copy).
- Never report the UI status as "Connected" before the MASQUE tunnel and exit IP are verified.
- Zero console window popups or unhandled runtime exceptions in WebView2.
- Preserve all existing Liquid Glass visual tokens, Acrylic desktop blur, and custom titlebar styling.
- Standard System Proxy mode remains default; no mandatory Windows UAC administrator elevation on normal launch.

## Review Focus

1. **JavaScript Error Shielding:** Verify that even if unexpected null properties arrive from status, `updateUI()` never throws an unhandled exception that halts the poller or SSE loop.
2. **Cancellation and Reconnection Race:** If a user rapidly toggles Connect and Disconnect, verify that previous sockets and listeners are closed immediately without port collision (`20808`/`20809` in use).
3. **Censorship-Resistant MASQUE Dial:** In filtered networks, verify that `dialProxyTLS` directly establishes HTTP/2 to Fastly edge port 2499 with TLS fragmentation without triggering ISP TCP resets.
4. **Offline Auth Recovery:** If `.firefox-vpn-tokens.json` is missing or invalid, verify the engine transitions cleanly to `StateError` with prompt "Please sign in with your Firefox account first" and opens the login dialog.
5. **Spherical Proportions at the Limb:** As the globe rotates, verify that South America, Africa, and Australia do not distort or taper into thin spikes when passing near the camera horizon.

---

### Task 1: Fix Frontend Uncaught `ReferenceError: isTun` & Wrap `updateUI()` Defensively

**Files:**
- Modify: `cmd/fyrnox/web/app.js:860-900`
- Test: Node.js headless DOM evaluation of `app.js`

**Interfaces:**
- Consumes: `status.wintun_active`, `status.tun_active`, `status.is_admin`
- Produces: Clean, uninterrupted execution of `updateUI(status)` across all DOM subtrees

- [ ] **Step 1: Write a headless Node.js test script reproducing the crash**

Create `scratch/test_update_ui.js`:
```javascript
const fs = require('fs');
const vm = require('vm');
const code = fs.readFileSync('cmd/fyrnox/web/app.js', 'utf8');
const window = { addEventListener: () => {}, EventSource: null };
const localStorage = { getItem: () => 'fa', setItem: () => {} };
const document = {
  getElementById: (id) => ({ style: {}, classList: { add: ()=>{}, remove: ()=>{}, toggle: ()=>{} } }),
  querySelectorAll: () => [],
  querySelector: () => ({ classList: { toggle: ()=>{} } }),
  body: { classList: { toggle: ()=>{} } },
  addEventListener: () => {}
};
const ctx = { window, document, localStorage, console, setTimeout, setInterval, clearInterval, navigator: {} };
vm.createContext(ctx);
vm.runInContext(code, ctx);
try {
  ctx.updateUI({ is_logged_in: true, server: { country_name: 'Peru', country_code: 'PE', city_code: 'SPJC', city_name: 'Lima' } });
  console.log('PASS');
} catch (e) {
  console.error('FAIL:', e.message);
  process.exit(1);
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node scratch/test_update_ui.js`
Expected: FAIL with `ReferenceError: isTun is not defined`

- [ ] **Step 3: Define `isTun` and wrap `updateUI` in defensive error handling in `cmd/fyrnox/web/app.js`**

In `cmd/fyrnox/web/app.js`:
1. Define `const isTun = isTunModeActive;` right after line 861.
2. In lines 1157, 1166, 1176 ensure consistent usage of `isTun`.
3. Wrap the interior of `function updateUI(status)` in a top-level `try { ... } catch (err) { console.error('Error in updateUI:', err); }`.

- [ ] **Step 4: Run test to verify it passes**

Run: `node scratch/test_update_ui.js`
Expected: PASS

---

### Task 2: Selected Server Subtitle & City Binding

**Files:**
- Modify: `cmd/fyrnox/web/index.html:130-145`
- Modify: `cmd/fyrnox/web/app.js:835-860`, `1054-1065`
- Test: Node.js headless DOM check for `dash-target-city` content

**Interfaces:**
- Consumes: `status.server.city_code`, `status.server.city_name`, `status.server.country_name`
- Produces: Correct formatted subtitle `#dash-target-city` (e.g. `SPJC · Lima` for Peru)

- [ ] **Step 1: Write test verifying server city display**

In `scratch/test_update_ui.js`, assert that `dash-target-city` receives `SPJC · Lima` when Peru is selected.

- [ ] **Step 2: Run test to verify current output**

Run: `node scratch/test_update_ui.js`
Expected: Text content missing or fallback `FAOR · Johannesburg`.

- [ ] **Step 3: Update `cmd/fyrnox/web/app.js` to format target city directly in Card 2**

In Section 6 of `updateUI` in `cmd/fyrnox/web/app.js`:
```javascript
const targetCity = document.getElementById('dash-target-city');
if (targetCity) {
  const cCode = status.server.city_code || 'POP';
  const cName = status.server.city_name || status.server.country_name || 'Selected';
  targetCity.textContent = `${cCode} · ${cName}`;
}
```
In `cmd/fyrnox/web/index.html`, remove hardcoded `FAOR · Johannesburg` placeholder and default to `Ready to Connect`.

- [ ] **Step 4: Run test to verify it passes**

Run: `node scratch/test_update_ui.js`
Expected: PASS with `dash-target-city` set to `SPJC · Lima`.

---

### Task 3: Settings Dynamic Account Sign In / Sign Out Sync & Token Logout

**Files:**
- Modify: `cmd/fyrnox/web/index.html:780-800`
- Modify: `cmd/fyrnox/web/app.js:1315-1335`, `2540-2560`, `2720-2740`
- Modify: `cmd/fyrnox/web/style.css` (add `.btn-danger` / sign-out button styles)
- Test: Headless auth state test verifying button text and class toggle

**Interfaces:**
- Consumes: `status.is_logged_in`, `status.user_email`, `currentLang`
- Produces: Dynamic `#settings-user-email` text, `#settings-auth-btn` text ("Sign Out" / "خروج از حساب") and red accent class

- [ ] **Step 1: Write test verifying Sign In vs Sign Out toggle**

Extend `scratch/test_update_ui.js`:
- Test A: When `is_logged_in: true`, `user_email: "user@test.com"`, button text must be "Sign Out" (or "خروج از حساب" in fa), and email must be "user@test.com".
- Test B: When `is_logged_in: false`, button text must be "Sign In" ("ورود به حساب") and email must be "Not logged in" ("وارد نشده‌اید").

- [ ] **Step 2: Run test to verify it fails**

Run: `node scratch/test_update_ui.js`
Expected: FAIL due to static `data-i18n` or hardcoded English strings.

- [ ] **Step 3: Implement dynamic Account Controller in `index.html`, `app.js`, and `style.css`**

1. In `cmd/fyrnox/web/index.html`:
   Remove `data-i18n="not_logged_in"` from `#settings-user-email` and `data-i18n="sign_in"` from `#settings-auth-btn`.
2. In `cmd/fyrnox/web/app.js`:
   Update section 12 in `updateUI(status)`:
   ```javascript
   if (setUserEmail && setAuthBtn) {
     if (status.is_logged_in) {
       setUserEmail.textContent = status.user_email || 'Firefox Account (Active)';
       setAuthBtn.textContent = (currentLang === 'fa') ? 'خروج از حساب' : 'Sign Out';
       setAuthBtn.classList.add('btn-signout');
     } else {
       setUserEmail.textContent = (currentLang === 'fa') ? 'وارد نشده‌اید' : 'Not logged in';
       setAuthBtn.textContent = (currentLang === 'fa') ? 'ورود به حساب' : 'Sign In';
       setAuthBtn.classList.remove('btn-signout');
     }
   }
   ```
3. Update `applyLanguage(lang)` in `app.js` to skip overwriting `#settings-auth-btn` and `#settings-user-email` if already rendered.
4. Update `handleAuthClick()`:
   If `currentStatus && currentStatus.is_logged_in`:
   Show confirm dialog; upon confirmation, call `doLogout()`.
5. In `cmd/fyrnox/web/style.css`:
   Add `.btn-signout` styling with subtle dark crimson/amber background (`rgba(220, 38, 38, 0.18)`), warm red border (`rgba(239, 68, 68, 0.4)`), and hover state.

- [ ] **Step 4: Run test to verify it passes**

Run: `node scratch/test_update_ui.js`
Expected: PASS

---

### Task 4: Rebuild Go Connection Engine (`cmd/fyrnox/engine.go` & `core.go`)

**Files:**
- Modify: `cmd/fyrnox/engine.go:1220-1470`
- Modify: `cmd/fyrnox/core.go`
- Test: `cmd/fyrnox/engine_test.go` (`TestLiveConnectStageByStage`, `TestConnectAtomicRollback`)

**Interfaces:**
- Consumes: `e.auth`, `e.selected`, `e.settings`, Fastly MASQUE edge endpoints
- Produces: Atomically verified VPN connection with WinINet proxy enabled and verified public exit IP

- [ ] **Step 1: Write an atomic rollback and cancellation test in `engine_test.go`**

Add `TestConnectAtomicRollback(t *testing.T)` to verify that:
1. `Connect()` succeeds on a valid server and transitions state cleanly to `StateConnected`.
2. Calling `Disconnect()` rolls back all listeners and resets proxy settings within 500ms.
3. If an invalid server is chosen, `Connect()` fails fast, returns an explicit error, and leaves no zombie listeners on ports 20808/20809.

- [ ] **Step 2: Run test to verify baseline**

Run: `go test -v -count=1 ./cmd/fyrnox -run TestLiveConnectStageByStage`
Expected: PASS (as established earlier)

- [ ] **Step 3: Refactor `VpnEngine.Connect()` into an atomic linear pipeline**

In `cmd/fyrnox/engine.go`:
1. Streamline state guards: ensure single atomic check under `e.mu`.
2. Encapsulate stages in a clean linear structure:
   - Stage 1: Validate session & proxy-pass token (`getOrRenewProxyPass`).
   - Stage 2: `newProxyController` with Fastly MASQUE HTTP/2 edge and TLS fragmentation.
   - Stage 3: Start local forwarding servers (`NewSocksProxyServer` & `NewHttpProxyServer`).
   - Stage 4: Run parallel `VerifyExit` probe with 5-second deadline. If verification fails, immediately invoke rollback and return actionable error.
   - Stage 5: Activate Windows System Proxy via `SetSystemProxyWithBypass` and broadcast `WM_SETTINGCHANGE`.
   - Stage 6: Atomically commit `e.state = StateConnected`, store `e.exitResult`, notify UI.
3. Ensure all stages respect `connectCtx` cancellation so any user-triggered disconnect halts background routines cleanly.

- [ ] **Step 4: Run test suite to verify implementation**

Run: `go test -v -count=1 ./cmd/fyrnox -run "TestLiveConnectStageByStage|TestConnectAtomicRollback"`
Expected: PASS

---

### Task 5: 3D Globe True Spherical Fibonacci Sampling & High-Definition Vector Borders

**Files:**
- Modify: `cmd/fyrnox/web/globe.js:360-450`, `460-520`
- Test: Headless Edge screenshot & particle distribution verification script

**Interfaces:**
- Consumes: `window.WORLD_GEOJSON.features` from `world_geo.js`
- Produces: Area-uniform spherical Fibonacci particles + crisp direct vector border line segments

- [ ] **Step 1: Write a mathematical validation test for spherical projection**

Create `scratch/test_globe_math.js`:
- Verify that `latLonToVector3(0, 0, R)` maps to `(0, 0, R)`.
- Verify that polar points (`lat = 90`) map to `(0, R, 0)`.
- Verify that Fibonacci point generation produces uniform distance distribution across latitudes -90 to +90.

- [ ] **Step 2: Run test to verify math correctness**

Run: `node scratch/test_globe_math.js`
Expected: Mathematical formulas verified.

- [ ] **Step 3: Implement Spherical Fibonacci Lattice and Direct Vector Borders in `cmd/fyrnox/web/globe.js`**

1. Replace `loadLandmassParticles()`:
   - Use a spherical Fibonacci spiral algorithm generating $N = 18,000$ points uniformly over the unit sphere.
   - Convert each candidate point $(x, y, z)$ back to geographic $(\text{lat}, \text{lon})$.
   - Test containment against the high-resolution polygon rings from `WORLD_GEOJSON`.
   - Result: Uniform particle density per $\text{km}^2$ worldwide—eliminating the equator-vs-polar squashing that made South America and Europe look distorted.
2. In `createCountryBoundaries()`:
   - Ensure `latLonToVector3` projects boundary vertices directly to sphere surface ($R \times 1.002$).
   - Configure `LineBasicMaterial` with `depthWrite: false` and `depthTest: true` to prevent backside phantom borders.

- [ ] **Step 4: Visual verification via screenshot capture**

Capture live screenshot of the updated dashboard using headless Edge and inspect globe proportions.
Expected: Continents display true geographic contours and relative areas.

---

### Task 6: Full Integration Verification & Windows Production Build

**Files:**
- Verify: `build/dist/fyrnox.exe`
- Test: Full Go test suite & live application launch

- [ ] **Step 1: Run the full Go test suite across all modules**

Run: `go test -v ./...`
Expected: PASS (all packages ok)

- [ ] **Step 2: Compile the Windows production binary**

Run: `go build -o build/dist/fyrnox.exe ./cmd/fyrnox`
Expected: Clean compilation, executable created.

- [ ] **Step 3: Verify binary file size and timestamp**

Run: `Get-Item build/dist/fyrnox.exe | Select-Object Name, Length, LastWriteTime`
Expected: ~21 MB executable created.

---

## Plan Review & Handoff

Please review this plan. Which execution approach would you prefer?

- **Native (Recommended)** — I execute the plan inline in this session, step by step with continuous verification under `/ponytail` mode, until all 6 tasks are complete and verified. Fastest and most direct.
- **Subagent-driven** — I dispatch separate subagents for each task with fresh review gates.
