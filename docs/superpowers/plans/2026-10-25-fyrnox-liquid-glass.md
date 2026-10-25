# FYRNOX v1.0.4 Liquid Glass & Frameless Window Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform FYRNOX into a frameless, unified Liquid Glass desktop application on Windows, featuring a custom translucent title bar, fluid 60fps GPU-accelerated motion, smart server recommendation engine, and reliable system routing.

**Architecture:**
1. Win32 frameless style transformation (`WS_CAPTION` stripped, `WS_THICKFRAME` retained for native edge resizing and Aero Snap) with JS bridge and REST fallbacks for window dragging and control actions.
2. Complete Liquid Glass design system in CSS/HTML: multi-layered glass surfaces, specular refraction borders, warm fiery obsidian palette (`#0E0705`, `#FF9400`), and fluid easing curves (`cubic-bezier(0.16, 1, 0.3, 1)`).
3. Smart server measurement worker pool with median calculation, candidate recommendation badge, and strict user confirmation rule.
4. Native installer and portable packaging for v1.0.4 with regression testing.

**Tech Stack:** Go 1.22+, Win32 User32 / DWM APIs, WebView2, HTML5 / Vanilla CSS / ES6 JavaScript.

**Spec:** `C:/Users/LouiNoor/.gemini/antigravity/brain/e4a584ef-f1a1-440a-ab06-38e888100028/fyrnox_liquid_glass_design.md`

## Global Constraints
- Operating System: Windows 10/11 (x64 and x86).
- Preserves native edge-resizing, Aero Snap, and multi-monitor dragging.
- Zero fake metrics or simulated connections.
- Native dark mode title bar styling with custom glass overlay.
- Communication with user in Persian (فارسی).

---

### Task 1: Frameless Win32 Window Integration & Window Management APIs

**Files:**
- Create: `cmd/fyrnox/window_windows.go`
- Modify: `cmd/fyrnox/main.go:160-225`
- Modify: `cmd/fyrnox/api.go:35-70`
- Test: `cmd/fyrnox/window_test.go`

- [ ] **Step 1: Create `cmd/fyrnox/window_windows.go`**
Implement Win32 window management:
- `applyFramelessStyle(hwnd uintptr)`
- `handleWindowDrag(hwnd uintptr)`
- `handleWindowMinimize(hwnd uintptr)`
- `handleWindowMaximize(hwnd uintptr)`
- `handleWindowClose(hwnd uintptr)`

- [ ] **Step 2: Bind window control functions in `main.go` and `api.go`**
Register `w.Bind` callbacks and HTTP REST endpoints `/api/window/minimize`, `/api/window/maximize`, `/api/window/close`, `/api/window/drag`.

- [ ] **Step 3: Run unit tests to verify window API structures**
Run: `go test -v ./cmd/fyrnox -run TestWindow`

---

### Task 2: Liquid Glass Design System & Custom Glass Title Bar

**Files:**
- Modify: `cmd/fyrnox/web/index.html`
- Modify: `cmd/fyrnox/web/style.css`
- Modify: `cmd/fyrnox/web/app.js`

- [ ] **Step 1: Add Custom Liquid Glass Title Bar in `index.html`**
Include:
- App icon (`favicon.png`), brand title (`FYRNOX`), version badge (`v1.0.4`), and live connection pill.
- Center drag region (`titlebar-drag-region`).
- Window control buttons: Minimize (`—`), Maximize/Restore (`□`), and Close (`✕`).

- [ ] **Step 2: Implement Liquid Glass styling and motion system in `style.css`**
- Set dark obsidian root canvas with warm mesh lighting.
- Layered translucent glass panels with `backdrop-filter: blur(24px)` and specular refraction borders.
- Fluid 60fps CSS transitions using GPU-accelerated transforms and `cubic-bezier(0.16, 1, 0.3, 1)`.
- Custom translucent scrollbars and hover effects.

- [ ] **Step 3: Connect Title Bar handlers in `app.js`**
Bind click and double-click handlers with fallback to both native JS bindings and REST API.

---

### Task 3: Smart Server Recommendation Engine & Bounded Ping Scanner

**Files:**
- Modify: `cmd/fyrnox/serverlist.go`
- Modify: `cmd/fyrnox/engine.go`
- Modify: `cmd/fyrnox/web/app.js`

- [ ] **Step 1: Implement bounded concurrent ping scanner and recommendation logic**
- Run max 5 concurrent ping workers with 2.5s deadline.
- Evaluate candidate with lowest latency (< 150ms).
- Mark server with `is_recommended: true` and recommendation reason.
- Ensure recommendation never switches user's selected server automatically.

- [ ] **Step 2: Add scanning progress bar and recommendation UI card**
- Animate progress bar during Ping All.
- Display "Recommended" gold badge on server card with "Use Recommended" button.

---

### Task 4: Packaging & End-to-End Build Verification

**Files:**
- Generate: `build/portable/Fyrnox-Portable/Fyrnox.exe` (x64)
- Generate: `build/portable-x86/Fyrnox-Portable/Fyrnox.exe` (x86)
- Generate: `build/installer/Fyrnox-Setup-v1.0.4.exe`
- Generate: `build/Fyrnox-v1.0.4-Windows-x64-Portable.zip`

- [ ] **Step 1: Compile all binaries for v1.0.4**
- [ ] **Step 2: Build installer and zip archives**
- [ ] **Step 3: Execute regression test suite**
