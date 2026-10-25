# FYRNOX Brand Liquid Glass & Live Log Console Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the FYRNOX client UI with a brand-tinted smoky charcoal Liquid Glass aesthetic (`rgba(11, 13, 18, 0.75)`) featuring glowing orange/amber subsurface gradients and ambient card lights, while fixing the in-app log console serialization bug to deliver a real-time, syntax-colored diagnostics terminal.

**Architecture:** Update CSS custom properties with brand-tinted glass variables, multi-point radial background mesh, and ambient glow spots. Refactor frontend `app.js` log parser to iterate over `LogEntry` structures with color-coded log level badges (`INFO`, `WARN`, `ERROR`, `DEBUG`), auto-scrolling, and asynchronous backend synchronization via `/api/logs/clear`.

**Tech Stack:** HTML5, CSS3 Glassmorphism (DWM Acrylic + Backdrop Filter + Radial Meshes), Vanilla ES6+ JavaScript, Go 1.24 HTTP/SSE server.

**Spec:** [`docs/superpowers/specs/2026-10-25-fyrnox-brand-liquid-glass-and-live-logs-design.md`](file:///C:/Users/LouiNoor/.gemini/antigravity/scratch/FoxyVPN-Windows/docs/superpowers/specs/2026-10-25-fyrnox-brand-liquid-glass-and-live-logs-design.md)

## Global Constraints

- Chat communication in Antigravity MUST be strictly in Persian (فارسی).
- Public documentation and GitHub issues MUST be 100% bilingual (English + Persian).
- Base glass tint MUST be smoky charcoal `rgba(11, 13, 18, 0.75)` with subsurface `#FF6A00` (opacity 0.08) and `#FFB000` (opacity 0.04).
- Glass borders MUST use warm glowing translucent tint `1px solid rgba(255, 106, 0, 0.25)`.
- All text and metrics MUST maintain crisp, high-contrast readability against the warm backdrop.
- Log console MUST escape log messages to prevent HTML injection and format as `[timestamp] [level] message`.

## Review Focus

1. **Log message HTML injection:** Unsanitized error messages containing `<` or `>` must be HTML-escaped before rendering in the terminal console.
2. **Empty or null log payload:** Backend returning empty array or nil slice must cleanly display "No logs available" without crashing.
3. **Log console auto-scroll behavior:** Ensure `scrollTop = scrollHeight` executes smoothly without interrupting active user text selection.
4. **Text contrast on glowing glass:** Ensure white text (`#F8FAFC`) remains effortlessly readable over orange/amber radial glow spots.
5. **Clear logs persistence:** Verify clicking Clear executes `POST /api/logs/clear` so logs do not reappear on the next 2-second polling tick.

---

### Task 1: Brand-Tinted Liquid Glass Stylesheet & Radial Ambient Mesh

**Files:**
- Modify: `cmd/fyrnox/web/style.css:15-65, 500-550, 1220-1310`
- Modify: `cmd/fyrnox/web/index.html:270-360`

**Interfaces:**
- Produces: Updated CSS root variables:
  - `--bg-app: rgba(11, 13, 18, 0.75);`
  - `--border-subtle: rgba(255, 106, 0, 0.22);`
  - `--border-medium: rgba(255, 106, 0, 0.38);`
  - `--shadow-ambient-fox: 0 0 35px rgba(255, 106, 0, 0.18);`
- Produces: Multi-point radial mesh background on `.app-content-wrapper`:
  - `radial-gradient(circle at 18% 25%, rgba(255, 106, 0, 0.08) 0%, transparent 45%), radial-gradient(circle at 82% 75%, rgba(255, 176, 0, 0.05) 0%, transparent 50%), rgba(11, 13, 18, 0.75)`
- Produces: Glowing ambient light spot behind `#hero-connect-card` and `.telemetry-grid`.

- [ ] **Step 1: Update CSS root variables and background mesh**
  Edit `cmd/fyrnox/web/style.css` to update `--bg-app`, `--border-subtle`, `--border-medium`, and add the brand subsurface gradient mesh to `body` and `.app-content-wrapper`.

- [ ] **Step 2: Add ambient glow and warm edge highlights to cards**
  Update `.hero-connect-card`, `.telemetry-card`, and `.view-panel` in `cmd/fyrnox/web/style.css` with warm borders and ambient fox glows.

- [ ] **Step 3: Verify text contrast and layout integrity**
  Ensure typography variables (`--text-primary: #F8FAFC`, `--text-secondary: #94A3B8`) remain sharp and unaffected by dark tint.

- [ ] **Step 4: Commit Task 1**
  ```bash
  git add cmd/fyrnox/web/style.css cmd/fyrnox/web/index.html
  git commit -m "feat(ui): implement FYRNOX brand-tinted Liquid Glass and ambient radial glow"
  ```

---

### Task 2: Live In-App Logger Engine Bugfix & Colored Terminal Formatting

**Files:**
- Modify: `cmd/fyrnox/web/app.js:1530-1565`
- Modify: `cmd/fyrnox/web/style.css:1950-1970`

**Interfaces:**
- Consumes: `/api/logs` returning JSON array of `LogEntry{Timestamp string, Level string, Message string}`.
- Consumes: `/api/logs/clear` accepting `POST` request.
- Produces: `renderLogs(logs)` parsing each entry into structured HTML with `.log-entry`, `.log-timestamp`, `.log-badge`, and `.log-msg`.
- Produces: `clearLogs()` dispatching `fetch('/api/logs/clear', { method: 'POST' })`.

- [ ] **Step 1: Add terminal styling and log-level classes in `style.css`**
  Add styles for `.log-entry`, `.log-timestamp` (`#64748B`), `.log-badge.info` (`#22C55E`), `.log-badge.warn` (`#EAB308`), `.log-badge.error` (`#EF4444`), `.log-badge.debug` (`#38BDF8`), and `.log-msg` (`#E2E8F0`).

- [ ] **Step 2: Implement structured log parser in `cmd/fyrnox/web/app.js`**
  Rewrite `renderLogs(logs)` to iterate through `LogEntry` items, escape HTML, build structured elements, update line count, and auto-scroll to the bottom.

- [ ] **Step 3: Update `clearLogs()` with backend API call**
  Update `clearLogs()` in `cmd/fyrnox/web/app.js` to send `POST /api/logs/clear`, clear DOM console, and reset line counter to 0.

- [ ] **Step 4: Commit Task 2**
  ```bash
  git add cmd/fyrnox/web/app.js cmd/fyrnox/web/style.css
  git commit -m "fix(logs): resolve object serialization bug and implement syntax-colored terminal logs"
  ```

---

### Task 3: Build Verification, Packaging & Bilingual GitHub Documentation

**Files:**
- Modify: `docs/issues/ISSUE_v1.0.4_OVERHAUL.md`
- Target: Run `go test` and `scripts/build_dist.ps1`

- [ ] **Step 1: Execute Go regression tests**
  Run `go test -v -skip "TestLiveConnectStageByStage" ./cmd/fyrnox` and ensure 100% pass.

- [ ] **Step 2: Build release distribution binaries**
  Run `powershell -ExecutionPolicy Bypass -File scripts\build_dist.ps1` to rebuild `Fyrnox.exe` and installer packages.

- [ ] **Step 3: Update bilingual documentation and post GitHub comment**
  Update `docs/issues/ISSUE_v1.0.4_OVERHAUL.md` with details of Brand Liquid Glass and Live Terminal Logs, and post bilingual comment to GitHub Issue #1 via `gh` CLI.

- [ ] **Step 4: Commit Task 3**
  ```bash
  git add docs/issues/ISSUE_v1.0.4_OVERHAUL.md
  git commit -m "docs(release): update v1.0.4 issue with brand liquid glass and live log console overhaul"
  ```
