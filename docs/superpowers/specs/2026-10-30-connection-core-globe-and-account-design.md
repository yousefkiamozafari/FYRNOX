# FYRNOX Atomic Connection Pipeline Rebuild, 3D Globe Spherical True-Scale Geometry, and Settings Account State Synchronization Design Specification

- **Date:** 2026-10-30
- **Author:** Lead Systems Architect & Senior Frontend Engineer
- **Status:** Proposed (Under Review)
- **Target Repository:** `FoxyVPN-Windows` (`yousefkiamozafari/FYRNOX`)

---

## 1. Intent & The "Why"

The FYRNOX Windows desktop client (Go + WebView2 + Three.js) currently exhibits three interrelated issues that block production usability:

1. **VPN Connection Pipeline Failure & UI Stalls:**
   Users clicking "Connect" experience either an indefinite stall or immediate failure. Investigation revealed a fatal JavaScript reference error in `cmd/fyrnox/web/app.js` (`ReferenceError: isTun is not defined` at line 897) which aborts `updateUI()` execution midway on every state update. In the backend (`cmd/fyrnox/engine.go`), connection setup is encumbered by convoluted cancellation contexts, redundant mutex contention, and fragile exit probes. The connection pipeline needs to be rebuilt from the ground up as a deterministic, atomic, linear engine.

2. **3D Globe Geographic Distortion (Country Shapes and Scales):**
   Continents and countries on the Three.js globe appear distorted, squashed, or unnaturally stretched (e.g. South America tapering into a narrow claw, Europe and Africa displaying inaccurate aspect ratios). This stems from rasterizing continent polygons onto a low-resolution (720×360) 2D offscreen equirectangular canvas before sampling points, coupled with perspective camera field-of-view distortions near the sphere's limb. The globe must be upgraded to direct vector-based spherical boundaries with area-uniform Fibonacci lattice particle distribution to render countries in their authentic geographic shapes and true proportions.

3. **Settings Account State Synchronization (Sign In vs Sign Out):**
   Even after successfully logging in with a Firefox Account (verified on disk via `.firefox-vpn-tokens.json` for `yousefkiamozafari@gmail.com`), the Settings page continues to show "Not logged in" and the primary button remains "Sign In". This occurs because static `data-i18n` attributes unconditionally overwrite dynamically assigned text, and the JavaScript crash in `app.js` prevents the account card logic from executing. Once signed in, the interface must display the user's active email and provide a functional "Sign Out" button.

---

## 2. Goals, Non-Goals, and Anti-Goals

### Goals
- **G1 (Atomic Connection Pipeline):** Rebuild the `Connect()` / `Disconnect()` lifecycle in `cmd/fyrnox/engine.go` into a clean, linear, fail-fast sequence:
  1. Validate authentication and obtain/refresh the Guardian proxy-pass token.
  2. Establish the duplex Fastly MASQUE HTTP/2 tunnel via `dialProxyTLS` (utilizing TLS fragmentation to bypass censorship).
  3. Spin up local SOCKS5 (`127.0.0.1:20808`) and HTTP (`127.0.0.1:20809`) forwarding listeners.
  4. Perform an end-to-end parallel exit probe (`VerifyExit`) with an authoritative egress IP returned within 3.5 seconds.
  5. Atomically commit the Windows System Proxy settings via WinINet registry APIs and broadcast `WM_SETTINGCHANGE`.
- **G2 (Frontend Stability & UI Unblocking):** Eliminate all runtime JavaScript errors in `cmd/fyrnox/web/app.js` (including fixing `isTunModeActive` scoping), ensuring `updateUI()` executes without interruption.
- **G3 (Authentic 3D Globe Geometry):** Render country borders and continent points using spherical coordinates with true geographic proportions, high-definition vector line segments from Natural Earth datasets, and an area-corrected point distribution that eliminates equirectangular squashing.
- **G4 (Bidirectional Account State in Settings):** Ensure that when `status.is_logged_in` is true, the Settings panel displays the user's email address and toggles the action button to "Sign Out" (خروج از حساب); clicking Sign Out calls `/api/logout`, revokes local tokens, and returns the button to "Sign In".

### Non-Goals
- **NG1:** Rewriting unrelated application pages (e.g., Diagnostics, Service Matrix, Split Tunneling rules).
- **NG2:** Replacing the WebView2 framework or altering the core Liquid Glass visual styling.
- **NG3:** Forcing Wintun Layer-3 TUN mode as the default routing mode (System Proxy remains default to avoid requiring Windows Administrator elevation on standard startup).

### Anti-Goals
- **AG1 (No False Positives):** Never report the UI status as "Connected" before the MASQUE tunnel and egress IP verification have genuinely succeeded.
- **AG2 (No Silent Failures):** Never swallow errors or leave the user in an infinite "Connecting..." spinner without actionable feedback and automatic resource cleanup.

---

## 3. Constraints

- **Platform:** Windows 10 and Windows 11 (x64 / x86).
- **Backend Stack:** Go 1.26, Windows WinINet & Win32 APIs, standard library HTTP/2 and TLS.
- **Frontend Stack:** Vanilla ES6+ JavaScript, Three.js (r128), HTML5, CSS3 with Acrylic backdrop support.
- **Networking Environment:** Host networks subject to aggressive censorship (Iran SNI/DNS filtering, port 53 blocking); upstream proxy `127.0.0.1:10808` supported for Mozilla API control plane.
- **No Destructive Operations:** Never perform git force-pushes or delete existing documentation/issues.

---

## 4. Approaches Considered

### Approach 1 (Selected): Atomic Linear Connection Engine + High-Fidelity Spherical Vector Globe + Dynamic Auth Bridge
- **Why Selected:** Fully solves the verified root causes with minimal architectural bloat. It eliminates the JS crash, restructures the Go connection engine into an atomic linear state machine, corrects the globe's coordinate math, and binds the account state seamlessly. It operates smoothly for all Windows users without requiring Administrator elevation.

### Approach 2 (Rejected): Kernel-Level Wintun-Exclusive TUN Engine
- **Why Rejected:** While routing all packets at Layer 3 through Wintun avoids WinINet registry proxy settings, it strictly requires Windows Administrator privileges (UAC prompt) on every launch. If the user runs the app normally or declines UAC, the entire application fails to connect.

---

## 5. Detailed Architecture & Component Specifications

### 5.1 Component A: Go Connection Engine Rebuild (`cmd/fyrnox/engine.go` & `core.go`)

```
+-----------------------------------------------------------------------------------+
|                            VpnEngine.Connect() Pipeline                           |
+-----------------------------------------------------------------------------------+
  |
  +--> [1. Pre-Flight] Validate auth (e.auth != nil) & selected server (e.selected != nil)
  |
  +--> [2. Proxy Pass] Fetch Guardian proxy-pass token (timeout: 6s) via apiProxy if set
  |
  +--> [3. MASQUE Dial] Duplex HTTP/2 CONNECT to targetURL:2499 with TLS fragmentation
  |
  +--> [4. Local Proxies] Start SOCKS5 (20808) & HTTP (20809) listeners on 127.0.0.1
  |
  +--> [5. VerifyExit] Multi-target parallel racing (api.ipify.org, icanhazip.com, ipify HTTPS)
  |                     Timeout: 5s. On failure -> rollback proxies & controller.
  |
  +--> [6. Route Commit] If UseSystemProxy -> SetSystemProxyWithBypass() & broadcast WinINet.
  |                     If EnableTunMode (and admin) -> EnableTunMode().
  |
  +--> [7. State Transition] Atomically set e.state = StateConnected, store exitResult, notify UI.
```

- **Cancellation & Rollback:** A single `context.WithCancel` manages the connection attempt. If the user clicks Disconnect during any stage, all open sockets, listeners, and controllers are closed immediately and cleanly.
- **Auto-Recovery on Stalls:** If any stage fails, `e.setError(err.Error())` is invoked, the state reverts to `StateError` or `StateDisconnected`, and no zombie proxy listeners remain running.

### 5.2 Component B: Frontend Error Elimination & DOM Unblocking (`cmd/fyrnox/web/app.js`)

- **Root Cause Fix:** Line 897 currently references an undeclared identifier `isTun`. This will be unified using `const isTun = isTunModeActive;` at the beginning of the routing section.
- **Defensive Error Handling:** Wrap `updateUI()` in a comprehensive `try/catch` block that logs errors to `console.error` rather than crashing the SSE/polling loop.
- **Server Card City Binding:** Ensure `dash-target-city` is updated directly in section 6 of `updateUI()` using:
  ```javascript
  const targetCity = document.getElementById('dash-target-city');
  if (targetCity) {
    targetCity.textContent = `${status.server.city_code || 'POP'} · ${status.server.city_name || status.server.country_name}`;
  }
  ```

### 5.3 Component C: High-Fidelity 3D Globe Geometry & Shapes (`cmd/fyrnox/web/globe.js`)

- **Direct Vector Geometry for Country Borders:**
  - Utilize high-resolution Natural Earth GeoJSON features from `world_geo.js`.
  - For each country feature, project boundary coordinates `(lat, lon)` directly to 3D Cartesian coordinates on sphere radius $R = 60 \times 1.002$:
    $$x = R \cdot \cos(\text{lat}) \cdot \sin(\text{lon})$$
    $$y = R \cdot \sin(\text{lat})$$
    $$z = R \cdot \cos(\text{lat}) \cdot \cos(\text{lon})$$
  - Store border lines in indexed `THREE.LineSegments` with `depthWrite: false` and render with an amber-tinted border material (`#EEA050` / `#4B6E88`).
- **Spherical Fibonacci Lattice for Landmass Particles:**
  - Replace the 720×360 equirectangular 2D rasterization with equal-area spherical sampling.
  - Distribute $\sim 18,000$ points uniformly over the sphere surface using the golden spiral (Fibonacci sphere algorithm):
    $$y_i = 1 - \frac{2i}{N - 1}, \quad \text{radius}_i = \sqrt{1 - y_i^2}, \quad \theta_i = \phi \cdot i$$
  - Filter points using polygon point-in-polygon checks against continent geometries so that particle density is strictly constant per square kilometer across all latitudes. This guarantees that Greenland, Scandinavia, South America, and Africa maintain their true relative geographical sizes without polar stretching.

### 5.4 Component D: Settings Account State & Dynamic Button Toggle (`cmd/fyrnox/web/`)

- **HTML Restructuring (`index.html`):**
  - Remove hardcoded `data-i18n="not_logged_in"` from `#settings-user-email` and `data-i18n="sign_in"` from `#settings-auth-btn`.
  - Provide base structure:
    ```html
    <div class="settings-item-title" id="settings-user-email">Not logged in</div>
    <button class="settings-btn-primary" id="settings-auth-btn" onclick="handleAuthClick()">Sign In</button>
    ```
- **Dynamic JavaScript Sync (`app.js`):**
  - In `updateUI(status)`:
    ```javascript
    if (setUserEmail && setAuthBtn) {
      if (status.is_logged_in) {
        setUserEmail.textContent = status.user_email || 'Firefox Account (Active)';
        setAuthBtn.textContent = (currentLang === 'fa') ? 'خروج از حساب' : 'Sign Out';
        setAuthBtn.classList.add('btn-danger');
      } else {
        setUserEmail.textContent = (currentLang === 'fa') ? 'وارد نشده‌اید' : 'Not logged in';
        setAuthBtn.textContent = (currentLang === 'fa') ? 'ورود به حساب' : 'Sign In';
        setAuthBtn.classList.remove('btn-danger');
      }
    }
    ```
  - In `handleAuthClick()`:
    - If `status.is_logged_in`: Show confirmation dialog; upon confirmation, invoke `POST /api/logout`, clear stored tokens, and refresh UI.
    - If not logged in: Open `#login-modal`.

---

## 6. Verification and Acceptance Criteria

| ID | Test Scenario | Acceptance Criteria |
| :---: | :--- | :--- |
| **V1** | **JS Error-Free Execution** | Opening DevTools / evaluating `updateUI()` results in 0 uncaught exceptions; all elements update without stopping. |
| **V2** | **Account State in Settings** | When valid tokens exist in `.firefox-vpn-tokens.json`, Settings displays `yousefkiamozafari@gmail.com` and button reads "Sign Out" / "خروج از حساب". |
| **V3** | **Sign Out Execution** | Clicking "Sign Out" deletes local token file, clears session in Go backend, and flips button back to "Sign In" / "ورود به حساب". |
| **V4** | **Selected Server City Display** | Selecting Peru displays "SPJC · Lima" (or Peru's city) rather than "FAOR · Johannesburg". |
| **V5** | **Live Connection Pipeline** | Clicking "Connect" transitions through `Authorizing` -> `Connecting` -> `Verifying` -> `Connected` within 4 seconds; browser traffic adopts VPN egress IP. |
| **V6** | **3D Globe Proportions** | Globe displays undistorted continents; South America, Africa, and Europe show accurate shapes matching physical globes. |
| **V7** | **Regression Build** | `go test -v ./...` passes 100%; `go build ./cmd/fyrnox` succeeds cleanly. |

---

## 7. What is Left to the Builder

- Exact styling details for `.btn-danger` (e.g. subtle dark glass red accent `#7F1D1D` vs vibrant brand red `#EF4444`).
- Minor tuning of the Fibonacci sphere particle count (15,000 to 22,000 points) based on GPU frame rate budget on target machine.
- Localized phrasing refinements in `I18N.fa` dictionary.
