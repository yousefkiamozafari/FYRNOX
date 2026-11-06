# Changelog

All notable changes to the **FYRNOX** project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
