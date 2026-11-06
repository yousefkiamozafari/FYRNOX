# Contributing to FYRNOX

Thank you for your interest in contributing to **FYRNOX**! We welcome contributions that improve reliability, performance, documentation, and user experience.

---

## 🛠️ Development Setup

### Prerequisites
* **Operating System:** Windows 10, Windows 11, or Windows Server 2019+
* **Go Compiler:** [Go 1.22+](https://golang.org/dl/) (Go 1.24+ recommended)
* **Git:** Installed and available in PATH
* **PowerShell:** Version 5.1 or PowerShell 7+

### Getting the Code
```powershell
git clone https://github.com/yousefkiamozafari/FYRNOX.git
cd FYRNOX
go mod download
```

### Running from Source
```powershell
go run ./cmd/fyrnox
```

### Building Distributables
To build all installers, portable archives, and standalone executables:
```powershell
.\scripts\build_dist.ps1
```

---

## 🧱 Architectural Guardrails

When modifying FYRNOX, you MUST respect the following core boundaries:
1. **Never Break Fallback Logic:** If Wintun TUN fails or lacks Administrator elevation, the engine must gracefully fall back to System Proxy mode (`SetSystemProxyWithBypass`) without closing local proxy listeners.
2. **Never Break Disconnect Cleanliness:** When the client disconnects, all routes, DNS caches (`DnsFlushResolverCache`), WinINet proxies, WinHTTP proxies, and user environment variables (`clearEnvVarsDirect`) must be thoroughly restored to direct mode.
3. **Multi-Architecture Support:** Any Windows API or driver integration must compile and function cleanly on both `amd64` (64-bit) and `386` (32-bit). Architecture-specific binaries must use Go build tags (`//go:build`).
4. **No Unapproved Dependencies:** Do not add external third-party dependencies unless strictly necessary and discussed beforehand.

---

## 🧪 Testing Before Submitting

Always run the automated test suite before opening a pull request:
```powershell
go test -v ./...
```

Run static code verification:
```powershell
go vet ./...
```

---

## 📋 Pull Request Guidelines

1. **Branch Naming:** Create a focused feature or fix branch from `main` (e.g. `fix/proxy-cleanup`, `feat/telemetry-latency`).
2. **Commit Messages:** Follow [Conventional Commits](https://www.conventionalcommits.org/) (e.g. `fix(tun): ...`, `docs: ...`, `refactor: ...`). Bilingual commit messages are appreciated.
3. **No Breaking Changes to Published Releases:** Never alter or remove existing release tags or published binary names.
4. **Link Related Issues:** Reference the issue number in your PR description (e.g. `Fixes #17`).
