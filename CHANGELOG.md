# Changelog

All notable changes to the **FYRNOX** project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-11-26

### Added
- **Native WinHTTP Control-Plane Transport & Firefox TLS Fingerprinting:** Added `winhttp_transport_windows.go`, `firefoxtls/`, and `tls_frag.go` for resilient FxA/Guardian/Fastly API transport.
- **3-Layer WFP IPv6 Leak Prevention & `RouteManager` (Pending VM Verification):** Added `cmd/fyrnox/wfp_windows.go`, `cmd/fyrnox/route_manager_windows.go`, and standalone `cmd/diagnostic-wfp` (`test_wfp_integration.exe`) for dynamic BFE filtering across `ALE_AUTH_CONNECT_V6`, `OUTBOUND_TRANSPORT_V6`, and `OUTBOUND_IPPACKET_V6` alongside `GetIpForwardTable2`/`GetBestRoute2` route tracking. *(Note: Full-Tunnel IPv6 leak prevention awaits final isolated-VM acceptance testing and packet capture before production sign-off).*
- **Safe Startup Route Recovery:** Added ownership-verified `RecoverOrphanedTunRoutes()` (`CleanupOrphanedFyrnoxRoutes`) inspecting IPv4 (`0.0.0.0/1`, `128.0.0.0/1`) and IPv6 (`::/1`, `8000::/1`) split routes via `FyrnoxTun` interface index, FYRNOX next-hop (`10.254.1.1`), or dead interface indices while preserving unrelated VPNs.
- **Single-Source Version Management:** Added `cmd/fyrnox/version.go` (`AppVersion = "1.1.0"`) with `-ldflags` injection across `cmd/fyrnox`, `cmd/installer`, `installer.iss`, and `scripts/build_dist.ps1`.

### Fixed
- **KillSwitchManager Error Propagation & IPv6 LAN Rules (D1):** `KillSwitchManager.Arm()` now enforces Admin privileges, verifies every `netsh advfirewall` rule, rolls back partial installations on error, and permits IPv6 loopback (`::1`) and link-local (`fe80::/10`) in `AllowLAN`.
- **IPv6 `VerifyTunCleanup()` Coverage (D2):** Extended `VerifyTunCleanup()` to inspect and clean lingering FYRNOX-owned IPv6 routes in addition to IPv4 routes.
- **Local API Security & Graceful Shutdown (D3):** Replaced unmanaged `http.ListenAndServe` with an owned `http.Server` and `Shutdown(ctx)` on exit; replaced wildcard CORS (`*`) with strict `127.0.0.1:<port>` / `localhost:<port>` `Host`, `Origin`, and `Sec-Fetch-Site` validation.
- **UI Telemetry & Exit IP Accuracy (D4/D5):** Fixed sparkline mappings to read `speed_down` / `speed_up`, removed the fabricated `42 ms` latency fallback in favor of `—`, and removed the fallback to physical `current_ip` when `exit_ip` is still verifying.
- **Server-Switch State Flicker & Connected Ping Integrity (D6):** Hot server switching now preserves `StateConnecting` without flashing `StateDisconnected`; `PingAllServers()` skips direct TCP sweeps while a tunnel is active so baseline latencies are not distorted.

---

## [1.0.6] - 2026-11-22

### Added
- **Proactive Token Expiration Validation:** Pre-flight validation (`freshRuntimeAuth`) in `cmd/fyrnox/engine.go` checking token validity and clock-skew safety margin ($\ge 120s$) before dispatching requests to Mozilla Guardian (`/api/v1/fpn/token`).
- **Decoupled Error Classification (HTTP 401 vs 403):** Introduced sentinel errors `ErrReauthRequired` and `ErrNotEntitled` in `fxa.go` and `guardian.go`.
- **Forced Token Refresh Primitive:** Added `ForceRefreshAndSaveTokens(ctx)` in `tokens.go` allowing one-shot forced token renewal after server-side Bearer token rejection.
- **Hermetic Token Lifecycle Test Suite:** Added 12 automated unit tests in `cmd/fyrnox/engine_auth_test.go` covering proactive renewal, controlled retry, error bubbling, 403 activation branching, concurrent thread safety, and rollback integrity.
- **Investigation Artifact:** Comprehensive technical report in `docs/reports/GUARDIAN_HTTP_401_INVESTIGATION.md`.

### Fixed
- **Mozilla Guardian HTTP 401 Rejections (#22):** Eliminated token expiration rejections by proactively renewing expired or nearing-expiration access tokens.
- **Invalid Activation Calls on 401:** Prevented erroneous Guardian activation attempts using already-rejected Bearer credentials.
- **Shadowed Refresh Token Failures:** Surfaced true underlying FxA session expiration (`ErrReauthRequired`) to the UI instead of generic Guardian 401 errors.
- **GitHub Issue Automation:** Integrated TLS 1.2/1.3 security protocol enforcement in `scripts/publish_github_issue.ps1`.

---

## [1.0.5] - 2026-11-17

### Added
- **Adaptive TLS ClientHello Fragmentation:** 2-stage fallback connection strategy in `cmd/fyrnox/core.go`. Automatically attempts standard TLS with bounded timeout $T_1 = \min(4.5s, R/2, R-3s)$ and transparently falls back to fragmented ClientHello when transport DPI interference or connection resets are detected, reserving $\ge 3s$ for Stage 2.
- **Strict TLS Error Classification:** `isEligibleTlsHandshakeFailure` filter distinguishes DPI handshake stalls/resets from terminal errors (certificate validation failure, hostname mismatch, context cancellation) to eliminate pointless retries.
- **Thread-Safe OAuth Token Lifecycle:** Concurrency-guarded token persistence (`tokenMu`, `refreshMu`) in `tokens.go` preserving existing `refresh_token` when Mozilla OAuth refresh responses omit it, guaranteeing zero session drops.
- **Hermetic Connection & Rollback Unit Tests:** In-memory mock doubles in `engine_test.go` allowing complete connection lifecycle and error-rollback testing without live network or external credentials.
- **Deterministic Exit Probe Failure Rollback Test:** Dedicated unit test (`TestConnectExitVerificationFailureRollback`) verifying state rollback to `StateError`, clearance of `ExitIP`, release of local SOCKS5/HTTP listeners, and non-blocking retry connectivity.
- **Aligned Multi-Architecture CI Toolchain:** Updated GitHub Actions CI and Release pipelines to Go `1.27.4`, matching repository `go.mod` specifications.

### Fixed
- **DPI Handshake Stalls on Anycast Endpoints:** Fixed connection hangs during initial TLS negotiation on Fastly edge servers across restrictive networks.
- **Silent Refresh Token Eviction:** Resolved defect where renewing an access token caused `refresh_token` to be overwritten with an empty string.
- **IPv6 Address Formatting in UDP Unit Tests:** Replaced `fmt.Sprintf("%s:%d", ...)` with `net.JoinHostPort` in `socks_udp_test.go` to eliminate `go vet` IPv6 diagnostics.

---

## [1.0.4] - 2026-11-06

### Added
- **Multi-Architecture Wintun Embedding:** Dedicated build-tagged embedding files (`wintun_embed_amd64.go`, `wintun_embed_386.go`, `wintun_embed_arm64.go`) bundling official 64-bit (427 KB), 32-bit (550 KB), and ARM64 (222 KB) Wintun drivers.
- **Automated Driver PE Validation Tests:** `wintun_embed_test.go` verifying DOS MZ magic, PE signature, machine type (`0x8664` vs `0x014C`), and runtime disk extraction.
- **Dynamic Build Script:** `scripts/build_dist.ps1` dynamically resolves repository root paths, eliminating hardcoded user profile dependencies.
- **100% Complete Persian (fa-IR) RTL Localization:** Full right-to-left layout, standard Persian typography, mirrored chevrons, and comprehensive dictionary covering all states and settings.
- **Liquid Glass Dark Theme:** Obsidian canvas (`#0B0D12`), frosted glass panels with `backdrop-filter: blur(20px)`, and neon orange `#FF6A00` / amber gold accents.
- **3D Fibonacci Dot-Matrix Globe:** Reconstructed spherical lattice with 26,000 points and Natural Earth vector borders without continent distortion; pauses rendering in background for zero GPU usage.
- **Parallel Ping Pool:** Concurrency-limited (6 workers) real-time TCP latency checks across all 35 edge nodes.

### Fixed
- **Proxy Deadlock on TUN Failure:** In `cmd/fyrnox/engine.go`, when `EnableTunMode` fails or lacks Administrator elevation, the engine now falls back gracefully to `SetSystemProxyWithBypass` without closing active local listeners, preventing internet outages.
- **Explicit Non-Admin TUN Error:** Converted silent `nil` return in `EnableTunMode` when `!IsAdmin()` into an explicit error so `engine.go` detects the failure and seamlessly activates System Proxy mode.
- **Explicit Wintun Error Propagation:** Converted silent `nil` returns in `tun_windows.go` (physical gateway detection, SOCKS5 upstream initialization, gVisor netstack creation) into explicit errors.
- **DNS Leak Prevention via DoH & SOCKS5 UDP Associate:** Full DNS tunneling using RFC 8484 and SOCKS5 UDP Associate to prevent ISP hijacking.
- **Hot-Reconnect Route Collision:** Clean teardown of existing tunnels, route loop prevention cleanup, and Windows DNS resolver cache flushing (`DnsFlushResolverCache`) during active server switches.

---

## [1.0.3] - 2026-10-25

### Added
- Monthly bandwidth quota tracking and session meter.
- Flagship differentiating features: LAN Proxy Sharing, Ping Shield QoS, panic key disconnect, and external service matrix test probes.
- Continuous 3D Slerp great-circle flight path routing on the globe.

### Fixed
- Idle jitter elimination on globe camera rotation.
- Instant disconnect logic ensuring complete listener teardown.

---

## [1.0.2] - 2026-10-22

### Added
- Live telemetry bar displaying download/upload speeds and connection duration.
- Kill switch toggle with automatic traffic blocking on unexpected disconnect.
- Auto-reconnect worker with exponential backoff.

### Fixed
- Frameless Win32 window drag, double-click maximize, and Aero Snap integration.
- Custom AppUserModelID registration to prevent generic taskbar grouping.

---

## [1.0.1] - 2026-10-19

### Added
- System-wide WinINet and WinHTTP proxy configuration.
- Local LAN bypass rules for `127.*`, `10.*`, `192.168.*`, and intranet domains.
- Native Inno Setup and custom Go-based self-contained installers.

### Fixed
- Console CMD window suppression using `-H windowsgui` subsystem flags.

---

## [1.0.0] - 2026-10-17

### Added
- Initial desktop release of FYRNOX for Windows.
- Direct authentication with Mozilla Firefox Accounts (FxA).
- Token retrieval from Mozilla Guardian API.
- Fastly MASQUE edge network proxy client supporting HTTP/2 and HTTP/3 (QUIC).
- Anti-censorship TLS ClientHello fragmentation (`sni` and `first_byte`).
- Basic local SOCKS5 (`127.0.0.1:20808`) and HTTP (`127.0.0.1:20809`) proxy listeners.
