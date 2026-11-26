# [Release] [Feature] [Security] FYRNOX v1.1.0 — Reliability, Security, Telemetry & Single-Source Release Packaging

## انتشار نسخه v1.1.0 — بهبود پایداری، امنیت API محلی، دقت تله‌متری، مدیریت تک‌منبعی نسخه و بسته‌های نصبی و Portable

---

## Release Downloads / لینک‌های دانلود نسخه v1.1.0

All binaries and portable packages for **Windows x64** and **Windows x86** are published on the GitHub Release page:  
تمامی فایل‌های نصبی، اجرایی و نسخه‌های **Portable** برای **ویندوز ۶۴ بیتی (x64)** و **۳۲ بیتی (x86)** در صفحه Release گیت‌هاب منتشر شده‌اند:

- **Release Page / صفحه انتشار:** [FYRNOX v1.1.0](https://github.com/yousefkiamozafari/FYRNOX/releases/tag/v1.1.0)
- **Installer x64 (Windows 10/11):** [`Fyrnox-Setup-v1.1.0.exe`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-Setup-v1.1.0.exe) | [`Fyrnox-Setup-Windows-11-10-x64.exe`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-Setup-Windows-11-10-x64.exe)
- **Installer x86 (Windows 7/8/10):** [`Fyrnox-Setup-Windows-7-8-10-x86.exe`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-Setup-Windows-7-8-10-x86.exe)
- **Portable x64 (ZIP):** [`Fyrnox-v1.1.0-Windows-x64-Portable.zip`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-v1.1.0-Windows-x64-Portable.zip)
- **Portable x86 (ZIP):** [`Fyrnox-v1.1.0-Windows-x86-Portable.zip`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-v1.1.0-Windows-x86-Portable.zip)
- **Standalone Executables:** [`Fyrnox-x64.exe`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-x64.exe) | [`Fyrnox-x86.exe`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/Fyrnox-x86.exe)
- **Driver & Checksums:** [`wintun.dll`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/wintun.dll) | [`checksums-sha256.txt`](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.1.0/checksums-sha256.txt)

---

## 1. Problem / Context (صورت مسئله و زمینه)

### English
Following the v1.0.6 release and the WinHTTP / WFP IPv6 engineering audits, a full product audit of FYRNOX identified seven high-priority engineering defects (D1–D7) impacting safety, correctness, local API security, UI accuracy, and release consistency:
1. **D1 — KillSwitchManager:** `Arm()` silently ignored `netsh advfirewall` failures and lacked Admin checks, atomic rollback, and IPv6 link-local/loopback (`::1`, `fe80::/10`) AllowLAN rules; `Disarm()` cleared `active` state even when rule deletion failed.
2. **D2 — Route Ownership & Cleanup:** `CleanupOrphanedFyrnoxRoutes()` and `VerifyTunCleanup()` needed strict FYRNOX ownership proof for IPv4 and IPv6 routes so third-party VPN routes or dead unrelated interfaces are never deleted.
3. **D3 — Local API Security & Shutdown:** The local control API previously bound `0.0.0.0` via `http.ListenAndServe` with wildcard CORS (`*`) and no graceful shutdown.
4. **D4/D5 — UI Telemetry & Exit IP Accuracy:** Sparklines read mismatched JSON keys, unmeasured latency showed a hardcoded `42 ms` fallback, and `exit_ip` fell back to physical `current_ip` before verification completed.
5. **D6 — Server Switch & Ping Integrity:** Hot server switching flashed `StateDisconnected` due to a race with `Disconnect()`, and `PingAllServers()` ran direct TCP pings while connected.
6. **D7 — Single-Source Version Drift:** `AppVersion` was duplicated across Go sources, Inno Setup (`installer.iss`), and build scripts (`scripts/build_dist.ps1`).

### فارسی
پس از انتشار نسخه `v1.0.6` و ممیزی‌های فنی WinHTTP و WFP، بررسی جامع محصول FYRNOX هفت ایراد کلیدی مهندسی (D1 تا D7) را در حوزه‌های ایمنی، صحت عملکرد، امنیت API محلی، دقت رابط کاربری و یکپارچگی نسخه شناسایی کرد:
1. **D1 — مدیریت KillSwitch:** متد `Arm()` خطاهای `netsh advfirewall` را نادیده می‌گرفت، فاقد بررسی دسترسی Admin و Rollback اتمیک بود و از `::1` و `fe80::/10` در AllowLAN پشتیبانی نمی‌کرد؛ همچنین `Disarm()` حتی در صورت خطا در حذف قوانین، وضعیت `active` را پاک می‌کرد.
2. **D2 — مالکیت روت‌ها و پاک‌سازی ایمن:** متدهای `CleanupOrphanedFyrnoxRoutes()` و `VerifyTunCleanup()` نیاز به اثبات قطعی مالکیت FYRNOX برای روت‌های IPv4/IPv6 داشتند تا روت‌های سایر VPNها یا اینترفیس‌های نامرتبط هرگز حذف نشوند.
3. **D3 — امنیت API محلی و خاموش‌سازی ایمن:** سرور API محلی روی `0.0.0.0` با `CORS: *` گوش می‌داد و فاقد `Shutdown` کنترل‌شده بود.
4. **D4/D5 — دقت تله‌متری UI و IP خروجی:** نمودارهای ترافیک کلید اشتباه می‌خواندند، تاخیر اندازه‌گیری‌نشده مقدار جعلی `42 ms` نشان می‌داد و `exit_ip` پیش از تایید به `current_ip` فیزیکی برمی‌گشت.
5. **D6 — پرش وضعیت در تعویض سرور و پینگ حین اتصال:** تعویض سریع سرور باعث نمایش موقت `StateDisconnected` می‌شد و `PingAllServers()` حین اتصال پینگ مستقیم می‌زد.
6. **D7 — پراکندگی منبع نسخه:** مقدار نسخه در چند فایل مختلف (کد Go، `installer.iss` و اسکریپت‌های بیلد) به صورت دستی تکرار شده بود.

---

## 2. Proposed Solution & Changes Implemented (راهکار و تغییرات اعمال‌شده)

### English
- **D1 (`cmd/fyrnox/killswitch_windows.go`):** Enforced Admin privilege verification, per-rule `netsh` add + `show rule` verification, reverse rollback on partial failure, IPv6 `::1,fe80::/10` in AllowLAN, and error-safe `Disarm()` that preserves `active=true` on failure.
- **D2 (`cmd/fyrnox/route_manager_windows.go`):** Enforced `InterfaceIndex > 0` plus strict ownership (`NextHop == 10.254.1.2` or active `FyrnoxTun` index, correlated into `ownedIfIndices` for IPv6 with explicit `interface=<idx>`). Unresolved routes are logged and left untouched.
- **D3 (`cmd/fyrnox/api.go`, `cmd/fyrnox/main.go`):** Bound listener strictly to `127.0.0.1:<port>`, added `securityMiddleware` validating `Host`, `Origin`, and `Sec-Fetch-Site`, and wired bounded `http.Server.Shutdown(ctx)`.
- **D4/D5 (`cmd/fyrnox/web/app.js`):** Mapped sparklines to `status.traffic.speed_down` / `speed_up`, replaced `42 ms` fallback with `—`, and displayed `Verifying...` until `exit_ip` is confirmed.
- **D6 (`cmd/fyrnox/engine.go`):** Passed `switchConnID int64` atomically through `SelectServer()`, `disconnectInternal()`, and `connectInternal()`, and returned cached latencies in `PingAllServers()` when tunnel is active.
- **D7 (`VERSION`, `version_embed.go`, `cmd/fyrnox/version.go`, `cmd/installer/main.go`, `installer.iss`, `scripts/build_dist.ps1`):** Embedded root `VERSION` (`1.1.0`) at compile time via `//go:embed VERSION` (`vpnclient.Version`) and updated `installer.iss`, `build_dist.ps1`, and `.github/workflows/release.yml`.
- **Control Plane & IPv6 Leak Prevention:** Integrated native WinHTTP transport (`winhttp_transport_windows.go`), Firefox TLS fingerprinting (`firefoxtls/`), and 3-layer WFP IPv6 containment (`cmd/fyrnox/wfp_windows.go`, pending final isolated-VM acceptance).

### فارسی
- **D1 (`cmd/fyrnox/killswitch_windows.go`):** بررسی اجباری دسترسی ادمین، راستی‌آزمایی تک‌تک قوانین با `show rule`، بازگشت معکوس (Rollback) در خطای جزئی، افزودن `::1` و `fe80::/10` به AllowLAN و حفظ وضعیت `active=true` در صورت شکست `Disarm()`.
- **D2 (`cmd/fyrnox/route_manager_windows.go`):** الزام `InterfaceIndex > 0` و مالکیت قطعی (`NextHop == 10.254.1.2` یا ایندکس فعال `FyrnoxTun` و نگاشت `ownedIfIndices` برای IPv6 همراه با `interface=<idx>` صریح). روت‌های نامشخص دست‌نخورده باقی می‌مانند.
- **D3 (`cmd/fyrnox/api.go` و `cmd/fyrnox/main.go`):** محدودسازی لیسنر به `127.0.0.1:<port>`، افزودن `securityMiddleware` برای بررسی `Host`، `Origin` و `Sec-Fetch-Site` و پیاده‌سازی `http.Server.Shutdown(ctx)`.
- **D4/D5 (`cmd/fyrnox/web/app.js`):** اتصال نمودار سرعت به `speed_down` / `speed_up`، جایگزینی `42 ms` با `—` و نمایش `Verifying...` تا تایید قطعی `exit_ip`.
- **D6 (`cmd/fyrnox/engine.go`):** عبور اتمیک `switchConnID int64` در زنجیره `SelectServer()` و حفظ پینگ‌های کش‌شده در `PingAllServers()` هنگام فعال بودن تونل.
- **D7 (`VERSION`، `version_embed.go`، `cmd/fyrnox/version.go`، `cmd/installer/main.go`، `installer.iss`، `scripts/build_dist.ps1`):** تعبیه فایل `VERSION` (`1.1.0`) در زمان کامپایل با `//go:embed VERSION` و یکپارچه‌سازی در نصاب و گردش‌کار انتشار گیت‌هاب.
- **کنترل پلین و پیشگیری از نشت IPv6:** یکپارچه‌سازی ترنسپورت بومی WinHTTP، اثرانگشت TLS فایرفاکس و فیلترینگ ۳ لایه WFP (در انتظار تایید نهایی روی VM ایزوله).

---

## 3. Verification & Testing (آزمایش و راستی‌آزمایی)

All 8 verification gates passed with exit code `0`:
1. `gofmt -l` → `0`
2. `go vet ./...` (`amd64` & `386`) → `0`
3. `go build ./...` (`amd64` & `386`) → `0`
4. `node --check cmd/fyrnox/web/app.js` → `0`
5. `go test -count=1 -timeout 60s ./...` → `0` (all 4 packages PASS, including D1–D7 regression tests)
6. `git diff --check` → `0`
7. `scripts/build_dist.ps1` → `0` (all x64/x86 installers, Portable ZIPs, standalone EXEs, and `checksums-sha256.txt` built & hashed)

### SHA-256 Checksums (`build/dist/checksums-sha256.txt`)
```text
F7463F54ABD7C2E4A574826D880F1C453481516B920EF75EC329CD3C354C9F1F  Fyrnox-Setup-v1.1.0.exe
F7463F54ABD7C2E4A574826D880F1C453481516B920EF75EC329CD3C354C9F1F  Fyrnox-Setup-Windows-11-10-x64.exe
817338490449152A09CA03DD495961CFD445B69E880509DA88F8EA5309BC349A  Fyrnox-Setup-Windows-7-8-10-x86.exe
F607C43E2988063FC5154F2AA3FFC55DE812F30395CCF8D97B1ABEE0AD709797  Fyrnox-v1.1.0-Windows-x64-Portable.zip
8268586DA26EC49941136EC01BC855DD3A16663122897B974621D239E12E97CD  Fyrnox-v1.1.0-Windows-x86-Portable.zip
FD3DE8F2334E3003BA40E8FD0ABF995BF7699693299FAC75E817D8D120BBE1DD  fyrnox.exe
E5DA8447DC2C320EDC0FC52FA01885C103DE8C118481F683643CACC3220DAFCE  wintun.dll
```

---

## 4. Acceptance Criteria & Next Steps (معیارهای پذیرش و اقدامات بعدی)

- [x] D1–D7 implemented, regression-tested, and verified locally (`go test`, `go vet`, `go build`)
- [x] Single-source `VERSION` (`1.1.0`) embedded and packaged into x64/x86 installers and Portable ZIPs
- [x] Git tag `v1.1.0` & GitHub Release published with all platform packages
- [ ] Isolated Windows VM acceptance execution (`test_wfp_integration.exe`, `test_wintun_fulltunnel.ps1`, physical-adapter packet capture)
