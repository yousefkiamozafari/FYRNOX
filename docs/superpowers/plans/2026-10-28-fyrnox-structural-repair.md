# FYRNOX UI/UX Structural Repair & View Ownership Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate view overlapping and layout corruption across the FYRNOX desktop application by relocating App Split Tunneling into the Settings view, restoring standard deterministic SPA panel visibility (`display: none` / `.active { display: flex }`), and establishing proper viewport scroll clearance above the floating Smoked Glass dock without touching working Go backend logic or regressing approved visual styling.

**Architecture:** 
1. The DOM view hierarchy is restored to exactly four top-level `<section class="view-panel">` elements (`#view-dashboard`, `#view-servers`, `#view-diagnostics`, `#view-settings`), perfectly matching `NAV_TABS` and the 4 navigation dock buttons.
2. The App Split Tunneling panel markup is moved inside `#view-settings` as a native `.settings-group-card.split-tunnel-card`.
3. Conflicting CSS rules forcing `position: absolute !important; display: block !important; opacity: 1 !important;` across all panels are purged. Only `.view-panel.active` is rendered (`display: flex !important;`); inactive panels are reliably hidden (`display: none !important;`).
4. Floating navigation dock (`.nav-dock-pill` at `bottom: 32px; left: 32px;`) and telemetry (`.bottom-telemetry-row` at `bottom: 32px; right: 32px;`) remain floating on top, with `110px` bottom padding applied to scrollable views to guarantee full visibility of scrolled content.

**Tech Stack:** Go (Fastly MASQUE / Mozilla Guardian / Wintun), WebView2, Vanilla JavaScript (ES6+), Three.js (WebGL Holographic Dot-Matrix Globe), Smoked Glass CSS Architecture.

**Spec:** Root-cause forensics report completed on 2026-10-28 based on baseline commit `e455e26` and regression commit `81296d8`.

---

## Global Constraints

- **STRICTLY NO GIT PUSH:** User explicitly ordered: *"این یکبار رو لازم نیست داخل گیت هاب منتشر کنی"*. Do not execute `git push`.
- **Zero Go Test Regressions:** All 28 tests in `./cmd/fyrnox/...` must continue to pass cleanly.
- **Preserve Approved Visual Design:** Maintain the Smoked Glass design tokens (`--fyrnox-bg: #17191D`, `--brand-amber: #D9954B`, `--brand-blue: #83AFCB`, `--text-primary: #E5E3DF`), dense holographic dot-matrix 3D globe with Fresnel rim glow, and floating oval dock.
- **Preserve All Element IDs:** Retain all `split-*`, `server-*`, `diag-*`, and `dash-*` element IDs for 100% backward compatibility with event listeners and Go IPC.
- **Ponytail Mode:** Minimal surgical diffs, no unnecessary abstractions or speculative code.
- **Language Requirement:** All user-facing explanations must be communicated in Persian (فارسی).

---

## Review Focus

1. **Split Tunneling controls inside Settings:** Checkbox toggle, preset chips, and process lists must remain fully interactive and trigger Go IPC when accessed within Settings.
2. **Tab switching cleanly isolates views:** Switching from Dashboard to Servers, Diagnostics, or Settings must never show ghost elements or overlapping content from any other tab.
3. **Scroll container clearance:** In Servers and Settings, scrolling to the very bottom must reveal all items cleanly above the 32px floating dock and telemetry row.
4. **Three.js WebGL canvas lifecycle:** Switching to other tabs and back to Dashboard must trigger `window.fyrnoxGlobe.handleResize()` and maintain smooth rendering without crashing or collapsing.
5. **Real-time Latency chip:** The Latency chip in the floating telemetry row must remain clickable and trigger a targeted single-server ping check without breaking the layout.

---

## Implementation Tasks

### Task 1: Relocate App Split Tunneling into Settings View (`cmd/fyrnox/web/index.html`)

**Files:**
- Modify: `cmd/fyrnox/web/index.html:324-432, 750-755, 815-820`

**Interfaces:**
- Consumes: Existing Split Tunneling DOM structure and IDs (`split-master-toggle`, `split-status-label`, `btn-split-bypass`, `btn-split-proxy`, `split-search-input`, `split-custom-exe-input`, `split-configured-count`, `split-configured-container`, `split-empty-msg`, `split-detected-count`, `split-process-rows`).
- Produces: A 4-panel top-level view structure (`#view-dashboard`, `#view-servers`, `#view-diagnostics`, `#view-settings`) directly under `<main class="app-view-container">`, with Split Tunneling integrated inside `.settings-sections-wrap`.

- [ ] **Step 1: Inspect and extract `#view-split-tunnel` markup**
  Extract lines 324-432 of `cmd/fyrnox/web/index.html`. Remove the `<section id="view-split-tunnel" class="view-panel">` element from between `#view-servers` and `#view-diagnostics`.

- [ ] **Step 2: Embed Split Tunneling as a `.settings-group-card` in `#view-settings`**
  Insert the extracted content into `<section id="view-settings">` inside `<div class="settings-sections-wrap">`, positioned right beneath the "Routing Modes & Tunnel Protocol" card. Wrap it in `<div class="settings-group-card split-tunnel-settings-card" id="settings-split-tunnel-card">`.

- [ ] **Step 3: Remove unnecessary `<div class="view-slider-track" id="view-slider-track">` wrapper**
  Unwrap the top-level sections so that `#view-dashboard`, `#view-servers`, `#view-diagnostics`, and `#view-settings` are direct child elements of `<main class="app-view-container">`, removing the unnecessary carousel track div.

- [ ] **Step 4: Verify HTML syntax and DOM structure**
  Verify matching tags and validate that all element IDs are preserved.

---

### Task 2: Restore SPA View Visibility Architecture & Scroll Clearance in CSS (`cmd/fyrnox/web/style.css`)

**Files:**
- Modify: `cmd/fyrnox/web/style.css:5480-5520, 5880-5935`

**Interfaces:**
- Consumes: Smoked Glass tokens (`--fyrnox-bg`, `--brand-amber`, `--brand-blue`, `--bg-card`, `--border-subtle`).
- Produces: Deterministic display rules where `.view-panel.active` is visible (`display: flex !important;`), inactive panels are hidden (`display: none !important;`), and scroll containers have 110px bottom clearance.

- [ ] **Step 1: Purge conflicting absolute stacking rules on `.view-panel`**
  Remove lines 5886-5915 where `.view-panel` is forced to `position: absolute !important; display: block !important; opacity: 1 !important; pointer-events: auto !important;` and `.view-slider-track` is forced to `display: block !important;`.

- [ ] **Step 2: Re-establish clean SPA visibility architecture**
  Define strict, deterministic view panel rules:
  ```css
  .app-view-container {
    flex: 1 1 0;
    min-height: 0;
    height: calc(100vh - 38px);
    max-height: calc(100vh - 38px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 0;
    min-width: 0;
    box-sizing: border-box;
    position: relative;
  }

  .view-panel {
    display: none !important;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .view-panel.active {
    display: flex !important;
    flex-direction: column;
  }

  #view-dashboard.active {
    display: flex !important;
    overflow: hidden !important;
    padding: 0 !important;
    max-width: none !important;
  }

  #view-servers.active,
  #view-diagnostics.active,
  #view-settings.active {
    padding: 20px 28px 110px 28px !important;
    max-width: 1240px !important;
    margin: 0 auto !important;
  }
  ```

- [ ] **Step 3: Preserve Floating Smoked Glass Dock & Telemetry styles**
  Confirm `.nav-dock-pill` is fixed at `bottom: 32px; left: 32px; z-index: 9999;` with blur 28px and saturated dark glass, and `.bottom-telemetry-row` is fixed at `bottom: 32px; right: 32px; z-index: 9999;`.

- [ ] **Step 4: Style `.split-tunnel-settings-card` inside Settings**
  Ensure `.split-tunnel-settings-card` seamlessly inherits Settings styling:
  - Transparent backgrounds on nested cards or matching acrylic borders
  - Clean layout for bypass/tunnel mode buttons and search/add binary controls

---

### Task 3: Simplify JavaScript Navigation & Lifecycle Handlers (`cmd/fyrnox/web/app.js`)

**Files:**
- Modify: `cmd/fyrnox/web/app.js:525-545, 546-612, 3300-3432`

**Interfaces:**
- Consumes: `NAV_TABS = ['dashboard', 'servers', 'diagnostics', 'settings']`, active nav state, DOM elements.
- Produces: Instant, robust view switching via `switchNav(tab)` and automatic split tunnel data loading when opening Settings.

- [ ] **Step 1: Update `switchNav(tab)` in `app.js`**
  - Set `activeNav = tab`.
  - Update dock indicator using `updateDockIndicator(tab)`.
  - Toggle `.active` on dock buttons (`.dock-item`).
  - Toggle `.active` class exclusively on `.view-panel` elements:
    `document.querySelectorAll('.view-panel').forEach(p => p.classList.toggle('active', p.id === 'view-' + tab));`
  - When `tab === 'settings'`, trigger both `fetchLANShareStatus()` and `loadSplitTunnelView()`.
  - When `tab === 'dashboard'`, trigger `window.fyrnoxGlobe.handleResize()`.
  - When `tab === 'servers'`, call `renderServerList()` or `fetchServers()`.
  - When `tab === 'diagnostics'`, call `fetchLogs()` and `fetchServiceMatrix()`.

- [ ] **Step 2: Remove conflicting `updateViewPanels` and drag translations**
  Remove or replace the carousel translate loop in `updateViewPanels` and disable the background pointer tracking that attempted to translate hidden panels horizontally.

- [ ] **Step 3: Ensure Split Tunnel actions trigger properly**
  Verify `loadSplitTunnelView()`, `toggleSplitTunnel()`, `setSplitMode()`, `filterProcessList()`, and `addCustomSplitApp()` work seamlessly when invoked in the Settings tab.

---

### Task 4: Comprehensive Verification, Visual Validation, and Dist Build

**Files:**
- Test execution: Go tests, headless Edge screenshots, PowerShell build script.

**Interfaces:**
- Consumes: Updated `index.html`, `style.css`, `app.js`.
- Produces: PASS on all 28 Go tests, 4 screenshot artifacts confirming zero overlapping, successful build output in `build/dist/`.

- [ ] **Step 1: Execute Go Test Suite**
  Run: `go test -v ./cmd/fyrnox/...`
  Expected: All 28 tests PASS (including live connect, fallback, quota, ping, lan share, and split tunnel rules).

- [ ] **Step 2: Headless Edge Screenshot Validation for All 4 Views**
  Capture high-res screenshots for:
  - `file:///.../cmd/fyrnox/web/index.html` (Dashboard)
  - Dashboard with tab switched to `servers`
  - Dashboard with tab switched to `diagnostics`
  - Dashboard with tab switched to `settings`
  Expected:
  - Zero view overlapping or bleeding elements.
  - Split tunneling card is visible and styled properly within Settings.
  - Floating dock is isolated at bottom left, telemetry at bottom right.
  - Globe is clearly visible on Dashboard without gray disk or web wireframe lines.

- [ ] **Step 3: Run Distribution Build Script**
  Run: `powershell -ExecutionPolicy Bypass -File .\scripts\build_dist.ps1`
  Expected: Successful compilation of `fyrnox.exe` without compiler or asset packaging errors.

- [ ] **Step 4: Verify Local Git Integrity (NO GIT PUSH)**
  Confirm that `git status` reflects clean local files and that NO `git push` is executed.
