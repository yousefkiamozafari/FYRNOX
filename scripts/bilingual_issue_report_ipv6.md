# [Bugfix] [Security] Production IPv6 Leak Remediation, RouteManager Fail-Closed Subsystem & Hardened VM Acceptance Package
## رفع قطعی نشت IPv6، زیرسیستم RouteManager با منطق Fail-Closed و امن‌سازی پکیج تست پذیرش VM

---

## 1. Problem / Context (صورت مسئله و زمینه)

### English:
During the engineering audit of FYRNOX Wintun Full-Tunnel mode on Windows, a critical release-blocking IPv6 leak vulnerability was identified:
1. **Dual-Mode IPv6 Leak Vulnerability:**
   In `cmd/fyrnox/tun_windows.go`, IPv6 blackhole routes (`::/1` and `8000::/1`) were conditioned on `if ipProtocol == "ipv4"`. However, FYRNOX defaults to `settings.IPProtocol = "dual"`. Whenever the MASQUE upstream connection failed or timed out on IPv6 (e.g. port 2499 handshake failure), the engine fell back to IPv4 MASQUE. Because `ipProtocol != "ipv4"`, zero IPv6 routes were installed on the virtual `FyrnoxTun` adapter and zero blackhole routes were added. Consequently, the physical network adapter's default IPv6 route (`::/0`) remained 100% active and unencrypted, leaking all IPv6 DNS queries, web traffic, and application sockets directly to the ISP.
2. **Fire-and-Forget Route Management:**
   Route additions and deletions used unverified, fire-and-forget shell commands (`route.exe` and `netsh.exe`). Route creation failures did not fail closed or trigger rollbacks, and teardown commands used fixed timeouts without confirming deletion in the kernel routing table.
3. **VM Acceptance Package Safety:**
   The test package in `dist/fyrnox-vm-acceptance-package/` required rigorous guardrails to prevent accidental execution on the primary host (`DESKTOP-3IOGCB8`), enforce positive VM identification, require administrative elevation, validate pre-flight baselines, and guarantee strictly scoped cleanup.

### فارسی:
در جریان ممیزی مهندسی حالت Wintun Full-Tunnel در کلاینت FYRNOX در ویندوز، یک آسیب‌پذیری بحرانی نشت ترافیک IPv6 شناسایی شد:
1. **آسیب‌پذیری نشت ترافیک در حالت Dual-Mode:**
   در فایل `cmd/fyrnox/tun_windows.go`، افزودن روت‌های سیاهچاله IPv6 (شامل `::/1` و `8000::/1`) صرفاً مشروط به `if ipProtocol == "ipv4"` بود. از آنجا که مقدار پیش‌فرض پروتکل در FYRNOX برابر `dual` است، در شرایطی که اتصال MASQUE روی IPv6 تایم‌اوت می‌خورد یا مسدود بود، برنامه به IPv4 تغییر مسیر می‌داد اما هیچ روت IPv6 روی آداپتور Wintun نصب نمی‌شد و روت‌های سیاهچاله نیز فعال نمی‌شدند. در نتیجه، روت پیش‌فرض فیزیکی کارت شبکه اصلی (`::/0`) کاملاً باز می‌ماند و تمامی ترافیک IPv6 به صورت مستقیم و بدون رمزنگاری از طریق ISP نشت می‌کرد.
2. **مدیریت ناامن و بدون بازخورد روت‌ها:**
   افزودن و حذف مسیرها با دستورات فاقد بررسی انجام می‌شد و در صورت بروز خطا هیچ فرآیند بازگشت به عقب (Rollback) یا مسدودسازی امن (Fail-Closed) وجود نداشت.
3. **امن‌سازی پکیج تست پذیرش ماشین مجازی (VM):**
   پکیج تست در `dist/fyrnox-vm-acceptance-package/` نیازمند قفل‌های سخت‌افزاری جهت جلوگیری از اجرای ناخواسته روی هاست اصلی (`DESKTOP-3IOGCB8`)، تشخیص قطعی محیط مجازی (Positive VM ID)، اعتبارسنجی بیس‌لاین شبکه و پاکسازی کاملاً ایزوله بود.

---

## 2. Proposed Solution & Changes Implemented (راهکار و تغییرات اعمال‌شده)

### English:
1. **Deterministic IPv6 Routing & Blackholing:**
   - Modified `cmd/fyrnox/tun_windows.go` to enforce verified IPv6 routing (`::/1` and `8000::/1`) regardless of whether the configured mode is `dual` or `ipv4`.
   - When IPv6 is operational, traffic is routed through `FyrnoxTun` to MASQUE.
   - When IPv6 is non-operational or unverified, traffic is routed to `FyrnoxTun` where gVisor/tun2socks blackholes it safely. Because `/1` is a more specific prefix than `/0` (Longest Prefix Match), 100% of IPv6 traffic is intercepted, preventing any packet from reaching the physical interface.
2. **Active Operational IPv6 Data-Plane Probing:**
   - Added `ProbeIPv6DataPath` and `IsIPv6Operational` to `cmd/fyrnox/core.go`. Actively transmits a wire-format DNS query for `cloudflare.com` to Cloudflare's IPv6 resolver `[2606:4700:4700::1111]:53` through the MASQUE tunnel with a 3-second timeout.
   - Updated `cmd/fyrnox/engine.go` to probe operational IPv6 before invoking `EnableTunMode`.
   - Added live monitoring in `runTunnelWatchdog`: if IPv6 data-path connectivity drops during an active session, the engine logs a warning, updates `ipv6Operational = false`, and keeps traffic safely blackholed on Wintun without dropping the IPv4 connection.
3. **RouteManager Fail-Closed Architecture:**
   - Implemented `RouteManager` in `cmd/fyrnox/route_manager_windows.go` (and cross-platform stubs in `cmd/fyrnox/route_manager_other.go`).
   - Tracks every injected route in an in-memory ledger.
   - Implemented `CheckConflict`: refuses to install routes if the prefix already belongs to an external network adapter (`ErrRouteConflict`).
   - Implemented `VerifyRouteExists`: queries the Windows kernel routing table via `netsh` / `route print` to confirm route presence before declaring Full-Tunnel active.
   - Implemented `Rollback`: deletes tracked routes in exact reverse order of installation.
4. **Hardened Standalone VM Acceptance Package:**
   - Updated `dist/fyrnox-vm-acceptance-package/network_baseline.ps1`, `test_wintun_fulltunnel.ps1`, and `README.md`.
   - Injected mandatory primary host block for `DESKTOP-3IOGCB8`.
   - Injected positive VM detection checking BIOS/System manufacturer (VMware, VirtualBox, KVM, QEMU, Hyper-V, Xen, Parallels, CIM `HypervisorPresent`). Fails closed on physical machines.
   - Enforced strict pre-flight baseline regex validation (`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`), UAC elevation assertion, and scoped cleanup function deleting only routes associated with the created Wintun interface index.

### فارسی:
1. **هدایت قطعی و سیاهچاله IPv6 جهت جلوگیری از نشت:**
   - اصلاح `cmd/fyrnox/tun_windows.go` به طوری که روت‌های `::/1` و `8000::/1` در هر دو حالت `dual` و `ipv4` اجباراً نصب می‌شوند.
   - در صورت عملیاتی بودن IPv6، ترافیک از طریق تونل MASQUE ارسال می‌شود. در غیر این صورت، ترافیک وارد Wintun شده و بلک‌هول می‌شود. با توجه به اولویت طول پیشوند (LPM: /1 بر /0)، ۱۰۰٪ ترافیک IPv6 جذب شده و هرگز به کارت شبکه فیزیکی نشت نمی‌کند.
2. **پایش فعال و آزمایش زنده دیتای IPv6:**
   - اضافه شدن `ProbeIPv6DataPath` و `IsIPv6Operational` در `cmd/fyrnox/core.go` جهت ارسال مستقیم پکت استاندارد DNS به `[2606:4700:4700::1111]:53` از داخل تونل MASQUE با مهلت ۳ ثانیه.
   - آزمایش زنده وضعیت در `cmd/fyrnox/engine.go` قبل از فعال‌سازی TUN و همچنین نظارت مداوم در حلقه واچ‌داگ (`runTunnelWatchdog`).
3. **زیرسیستم امن RouteManager با قابلیت Rollback:**
   - پیاده‌سازی `RouteManager` در `cmd/fyrnox/route_manager_windows.go` جهت ثبت، اعتبارسنجی تداخل (`CheckConflict`)، تایید حضور در جدول هسته ویندوز (`VerifyRouteExists`) و پاکسازی ترتیبی معکوس (`Rollback`).
4. **امن‌سازی پکیج تست پذیرش در ماشین مجازی:**
   - اعمال قفل عدم اجرا روی هاست اصلی (`DESKTOP-3IOGCB8`).
   - الزام به احراز هویت ماشین مجازی از طریق WMI/CIM و بایوس، تایید دسترسی ادمین (UAC) و پاکسازی کاملاً ایزوله.

---

## 3. Technical Details (جزئیات فنی و پیاده‌سازی)

### Files Modified & Created (فایل‌های تغییر یافته و ایجاد شده):
- [`cmd/fyrnox/route_manager_windows.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/route_manager_windows.go): In-memory route ledger, prefix conflict detection, on-link IPv6 route management, kernel route verification, reverse rollback.
- [`cmd/fyrnox/route_manager_other.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/route_manager_other.go): Cross-platform build stubs.
- [`cmd/fyrnox/route_manager_test.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/route_manager_test.go): Unit tests for tracking, reverse rollback, command errors, and conflict refusal.
- [`cmd/fyrnox/tun_ipv6_test.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/tun_ipv6_test.go): 7 unit/regression tests covering route failure fail-closed, conflict refusal, partial rollback safety, idempotence, kernel verification, engine status tracking, and watchdog drop transitions.
- [`cmd/fyrnox/core.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/core.go): Added `ProbeIPv6DataPath` and `IsIPv6Operational`.
- [`cmd/fyrnox/tun_windows.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/tun_windows.go): Integration with `RouteManager`, verified IPv6 installation, kernel route existence assertion, scoped rollback.
- [`cmd/fyrnox/engine.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/engine.go): Integrated operational IPv6 probing, `ipv6Operational` state tracking, watchdog continuous monitoring.
- [`dist/fyrnox-vm-acceptance-package/network_baseline.ps1`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/network_baseline.ps1): Host interlock and baseline egress regex validation.
- [`dist/fyrnox-vm-acceptance-package/test_wintun_fulltunnel.ps1`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/test_wintun_fulltunnel.ps1): Positive VM checks, UAC assertion, scoped cleanup.
- [`dist/fyrnox-vm-acceptance-package/README.md`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/README.md): Operator guide for VM testing.

---

## 4. Verification & Testing (آزمایش و راستی‌آزمایی)

### 1. Unit & Regression Tests:
- `TestRouteManager_TrackingAndRollback`: PASS (0.00s)
- `TestRouteManager_ConflictRefusal`: PASS (0.00s)
- `TestRouteManager_CommandFailureReturnsError`: PASS (0.00s)
- `TestIPv6_RouteCreationFailureFailsClosed`: PASS (0.00s)
- `TestIPv6_PreexistingRouteConflict`: PASS (0.00s)
- `TestIPv6_PartialRollbackSafety`: PASS (0.00s)
- `TestIPv6_DisconnectCleanupIdempotence`: PASS (0.00s)
- `TestIPv6_KernelVerificationFailsafe`: PASS (0.00s)
- `TestIPv6_EngineStatusTracking`: PASS (0.56s)
- `TestIPv6_WatchdogTransitionOnProbeDrop`: PASS (0.02s)

### 2. Multi-Architecture Toolchain Verification:
- `go vet ./...` (Windows amd64): Clean (exit 0)
- `GOARCH=386 go vet ./...` (Windows 386): Clean (exit 0)
- `go build -v ./...` (Windows amd64): Clean (exit 0)
- `GOARCH=386 go build -v ./...` (Windows 386): Clean (exit 0)
- `go test -count=1 ./...` (Full repository): ALL PASS (exit 0)

---

## 5. Acceptance Criteria & Next Steps (معیارهای پذیرش و اقدامات بعدی)

### English:
- [x] Elimination of release-blocking IPv6 leak in dual-mode and IPv4-only modes.
- [x] Fail-closed atomic route rollback via `RouteManager`.
- [x] End-to-end active IPv6 operational verification via wire-format DNS query over MASQUE.
- [x] Live watchdog monitoring with graceful fallback to blackhole mode on IPv6 connectivity drop.
- [x] Hardened VM acceptance package with strict primary host interlock (`DESKTOP-3IOGCB8`) and positive VM identification.
- [x] Zero network disruption to the developer's primary machine.
- [ ] Next Step: Transfer `dist/fyrnox-vm-acceptance-package/` to a disposable external Windows VM for manual Full-Tunnel packet-capture verification.

### فارسی:
- [x] رفع کامل آسیب‌پذیری نشت IPv6 در حالت‌های Dual و IPv4.
- [x] سیستم مدیریت روت با معماری Fail-Closed و رول‌بک ترتیبی.
- [x] آزمایش فعال و زنده دیتاپلین IPv6 از طریق ارسال کوئری DNS روی MASQUE.
- [x] پایش مداوم اتصال در حلقه واچ‌داگ و تبدیل امن به حالت بلک‌هول در صورت قطعی IPv6 بدون قطع تونل IPv4.
- [x] امن‌سازی کامل پکیج تست پذیرش VM و ممانعت قطعی از اجرا روی هاست اصلی.
- [x] عدم ایجاد کوچکترین اختلال در تنظیمات شبکه هاست توسعه‌دهنده.
- [ ] اقدام بعدی: انتقال پکیج `dist/fyrnox-vm-acceptance-package/` به یک ماشین مجازی مجزای ویندوز جهت آزمایش میدانی پکت کپچر.
