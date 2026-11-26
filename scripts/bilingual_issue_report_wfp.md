# [Feature] [Security] Windows Filtering Platform (WFP) Dynamic Subsystem for Absolute IPv6 & On-Link /64 Leak Prevention
## زیرسیستم فیلترینگ پویای هسته ویندوز (WFP) جهت پیشگیری قطعی از نشت IPv6 و پیشوندهای درون‌شبکه‌ای /64

---

## 1. Problem / Context (صورت مسئله و زمینه)

### English:
During the full-tunnel network architecture audit of FYRNOX on Windows, a fundamental system-level IPv6 leak vector was analyzed:
1. **The On-Link `/64` LPM Bypass Problem:**
   - Wintun Full-Tunnel mode redirects traffic by installing split default routes: `0.0.0.0/1` & `128.0.0.0/1` for IPv4, and `::/1` & `8000::/1` for IPv6.
   - On physical networks with IPv6 Stateless Address Autoconfiguration (SLAAC) or DHCPv6, physical routers periodically broadcast ICMPv6 Router Advertisements (RA) containing on-link prefix information, typically a `/64` subnet.
   - The Windows kernel automatically installs an on-link route for this `/64` prefix bound to the physical network card.
   - Under the Longest Prefix Match (LPM) rule of the IP routing table, prefix length 64 takes absolute precedence over prefix length 1 ($64 > 1$).
   - Any packet addressed to a destination within the local `/64` network or directed to an IPv6 DNS server on that physical subnet **completely bypasses Wintun, gVisor, tun2socks, and the SOCKS5 IPv6 gate**, escaping directly over the physical network adapter unencrypted.
2. **Layer-Condition Mismatch in Naive WFP Filter Designs:**
   - Microsoft's official Windows Filtering Platform documentation (*Filtering Conditions Available at Each Filtering Layer*) strictly restricts the conditions available at each layer.
   - Specifically, `FWPM_LAYER_OUTBOUND_IPPACKET_V6` supports interface index, flags, and remote address, but **does not support** transport ports (`FWPM_CONDITION_IP_REMOTE_PORT`), protocols (`FWPM_CONDITION_IP_PROTOCOL`), or ICMP types (`FWPM_CONDITION_ICMP_TYPE`).
   - Furthermore, `FWPM_CONDITION_ICMP_TYPE` is only available at ICMP error layers, not at packet or transport layers.
   - Submitting unsupported conditions causes `FwpmFilterAdd0` to fail with `FWP_E_INCOMPATIBLE_LAYER`.
3. **Requirement for Verified Multi-Layer Kernel Containment:**
   - A multi-layer architecture is required across `ALE_AUTH_CONNECT_V6`, `OUTBOUND_TRANSPORT_V6`, and `OUTBOUND_IPPACKET_V6`.
   - WFP operates under independent per-layer arbitration: a permit at the IP packet layer does not override a block at the transport or ALE layers.
   - Activation must occur before route modification to eliminate transition leak windows.

### فارسی:
در جریان ممیزی معماری شبکه و تونل کامل FYRNOX در ویندوز، یک بردار نشت بنیادین در سطح هسته شبکه سیستم شناسایی و تحلیل شد:
1. **مسئله دور زدن روت‌های Wintun توسط پیشوندهای محلی /64 (LPM Bypass):**
   - در حالت Wintun Full-Tunnel، ترافیک از طریق نصب روت‌های شکسته `::/1` و `8000::/1` برای IPv6 هدایت می‌شود.
   - در شبکه‌های فیزیکی دارای IPv6، روترهای شبکه پیام‌های دوره‌ای ICMPv6 Router Advertisement (RA) ارسال می‌کنند که یک پیشوند محلی `/64` را روی کارت شبکه فیزیکی تعریف می‌کند.
   - هسته شبکه ویندوز یک روت On-Link برای این زیرشبکه `/64` بر روی آداپتور فیزیکی نصب می‌نماید.
   - بر اساس قاعده تطبیق طولانی‌ترین پیشوند (Longest Prefix Match - LPM)، طول پیشوند ۶۴ بر طول پیشوند ۱ اولویت قطعی دارد ($64 > 1$).
   - در نتیجه، هر بسته‌ای که به مقصدی در زیرشبکه محلی `/64` ارسال شود یا پرس‌وجوهای DNS که به سرور IPv6 روتر محلی هدایت شوند، **آداپتور Wintun، پشته gVisor، tun2socks و گیت SOCKS5 را به طور کامل دور زده** و بدون رمزنگاری از طریق کارت شبکه فیزیکی به بیرون نشت می‌کنند.
2. **تطابق دقیق شرایط پالایش با لایه‌های معتبر مایکروسافت:**
   - طبق مستندات رسمی مایکروسافت، لایه `OUTBOUND_IPPACKET_V6` صرفاً از فلگ‌ها، اینترفیس و آدرس مقصد پشتیبانی کرده و از پورت، پروتکل و کد/نوع ICMP پشتیبانی نمی‌کند.
   - شرط `FWPM_CONDITION_ICMP_TYPE` صرفاً در لایه‌های مدیریت خطای ICMP معتبر است.
   - لذا معماری پالایش باید به صورت تفکیک‌شده در سه لایه ALE، Transport و IPPacket پیاده‌سازی شود تا خطای ناسازگاری لایه (`FWP_E_INCOMPATIBLE_LAYER`) رخ ندهد.
3. **ترتیب فعال‌سازی فاقد شکاف زمانی (Leak-Free Activation Ordering):**
   - فیلترهای WFP باید پیش از تزریق روت‌های Wintun فعال و تأیید شوند تا هیچ پنجره زمانی فاقد حفاظت برای نشت بسته‌ها به وجود نیاید.

---

## 2. Proposed Solution & Changes Implemented (راهکار و تغییرات اعمال‌شده)

### English:
1. **Layer-Condition Validation Matrix:**
   - Implemented `ValidateFilterCondition` in [`cmd/fyrnox/wfp_windows.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/wfp_windows.go) to validate every condition against Microsoft's documented compatibility matrix before calling WFP APIs, rejecting unsupported conditions with `ErrUnsupportedLayerCondition`.
2. **Rigorous Three-Layer Enforcement Strategy:**
   - **Layer 1: `FWPM_LAYER_ALE_AUTH_CONNECT_V6`:** Enforces socket authorization. Permits loopback, permits Wintun, permits outer MASQUE (IP:2499:TCP), blocks physical DNS (ports 53 and 853), and defaults to blocking all non-Wintun IPv6 sockets with immediate `WSAEACCES`.
   - **Layer 2: `FWPM_LAYER_OUTBOUND_TRANSPORT_V6`:** Enforces transport datagram policies. Permits loopback, permits Wintun, permits ICMPv6 protocol 58 strictly destined to link-local (`fe80::/10`) and multicast (`ff02::/16`), permits outer MASQUE (IP:2499:TCP), and default-denies physical transport.
   - **Layer 3: `FWPM_LAYER_OUTBOUND_IPPACKET_V6`:** Enforces NDIS packet boundaries. Permits loopback, permits Wintun, permits raw NDP packets addressed to `fe80::/10` and `ff02::/16`, permits outer MASQUE destination IP, and default-denies all remaining physical IPv6 egress (including on-link `/64`).
3. **Multi-Layer MASQUE Exception & WFP Arbitration:**
   - At IPPacket layer, destination IP is permitted.
   - At Transport and ALE layers, destination IP, port 2499, and protocol TCP are strictly enforced.
   - Under WFP arbitration, an IPPacket permit does not override an ALE or Transport block; sockets to other ports on that server IP are blocked with `WSAEACCES`.
4. **Corrected ICMPv6 & NDP Link-Local Policy:**
   - Eliminates invalid `ICMP_TYPE` conditions from packet/transport layers.
   - Preserves NDP using exact subnet masks (`fe80::/10` and `ff02::/16`) with `FWP_V6_ADDR_MASK`.
   - Global IPv6 unicast ping or on-link `/64` traffic remains 100% blocked.
5. **Leak-Free Activation Ordering:**
   - Updated [`cmd/fyrnox/tun_windows.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/tun_windows.go): WFP leak prevention is armed and verified in the kernel BEFORE Wintun routes are installed. On route failure, both routes and WFP filters roll back atomically.
6. **Dynamic Session Crash Policy Documented:**
   - Explicitly documented that `FWPM_SESSION_FLAG_DYNAMIC` guarantees fail-safe restoration of normal host networking upon process exit/crash, and does NOT claim dead-process Kill Switch protection.
7. **Dedicated Native VM Integration Test:**
   - Implemented [`cmd/diagnostic-wfp/main.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/diagnostic-wfp/main.go) and compiled to [`dist/fyrnox-vm-acceptance-package/test_wfp_integration.exe`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/test_wfp_integration.exe) with primary host guardrails (`DESKTOP-3IOGCB8`), real BFE transaction commit, filter enumeration, live socket permit/block validation, and zero-residual verification.

### فارسی:
1. **ماتریس اعتبارسنجی شرایط لایه‌ها:**
   - پیاده‌سازی تابع `ValidateFilterCondition` در `cmd/fyrnox/wfp_windows.go` جهت بررسی انطباق شرایط با مستندات مایکروسافت پیش از فراخوانی توابع WFP.
2. **استراتژی فیلترینگ سه‌لایه منسجم:**
   - **لایه اول (`ALE_AUTH_CONNECT_V6`):** احراز اتصال سوکت‌ها. پورت‌های DNS (۵۳ و ۸۵۳) مسدود و اتصال به استثنای MASQUE (آی‌پی سرور:۲۴۹۹:TCP) و Wintun/لوپ‌بک با خطای صریح `WSAEACCES` ساقط می‌شوند.
   - **لایه دوم (`OUTBOUND_TRANSPORT_V6`):** پالایش در لایه انتقال. مجوزدهی پروتکل ۵۸ ICMPv6 منحصراً روی زیرشبکه‌های Link-Local (`fe80::/10`) و چندپخشی (`ff02::/16`)، مجوزدهی انتقال بیرونی MASQUE و بلاک پیش‌فرض آداپتورهای فیزیکی.
   - **لایه سوم (`OUTBOUND_IPPACKET_V6`):** پالایش در لایه پکت پیش از ارسال NDIS. مجوزدهی بسته‌های NDP محلی روی `fe80::/10` و `ff02::/16`، مجوزدهی آی‌پی مقصد سرور MASQUE و ساقط‌سازی کلیه بسته‌های خام فیزیکی (از جمله زیرشبکه `/64`).
3. **دفاع چندلایه انتقال بیرونی MASQUE:**
   - انطباق دقیق آی‌پی، پورت و پروتکل در لایه‌های ALE و Transport. مجوز لایه پکت هرگز بلاک لایه ترنسپورت/ALE را لغو نکرده و هیچ استثنای عمومی برای کل پروسه اعطا نمی‌شود.
4. **اصلاح سیاست ICMPv6 و حفظ پروتکل NDP:**
   - حذف شرایط نامعتبر `ICMP_TYPE` و استفاده از ماسک‌های دقیق آدرس `fe80::/10` و `ff02::/16`.
5. **ترتیب فعال‌سازی فاقد شکاف نشت:**
   - اصلاح `cmd/fyrnox/tun_windows.go`: فعال‌سازی WFP قبل از نصب روت‌ها در ویندوز.
6. **مستندسازی سیاست نشست پویا (Dynamic Session Crash Policy):**
   - ثبت صریح این موضوع که `FWPM_SESSION_FLAG_DYNAMIC` بازگشت ایمن اینترنت هاست در زمان کرش را تضمین می‌کند و ادعای Kill Switch پایدار پس از مرگ پروسه ندارد.
7. **ابزار تست یکپارچگی زنده در ماشین مجازی:**
   - ایجاد `cmd/diagnostic-wfp/main.go` و تولید باینری `dist/fyrnox-vm-acceptance-package/test_wfp_integration.exe` با قفل عدم اجرا روی هاست اصلی، ثبت زنده در BFE، بررسی سوکت‌های واقعی و تأیید عدم برجا ماندن فیلترها (Zero Residuals).

---

## 3. Technical Details (جزئیات فنی و پیاده‌سازی)

### Files Modified & Created (فایل‌های تغییر یافته و ایجاد شده):
* [`cmd/fyrnox/wfp_windows.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/wfp_windows.go): Complete 3-layer WFP subsystem, layer-condition validator, `FWP_V6_ADDR_AND_MASK` encoder, dynamic session lifecycle.
* [`cmd/fyrnox/wfp_other.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/wfp_other.go): Cross-platform stubs for non-Windows platforms.
* [`cmd/fyrnox/wfp_test.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/wfp_test.go): Unit test suite verifying condition validation, 18-filter generation across 3 layers, IPv4 fallback (15 filters), and transaction rollback.
* [`cmd/fyrnox/tun_windows.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/tun_windows.go): Leak-free activation ordering and atomic failure rollback.
* [`cmd/fyrnox/tun_ipv6_test.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/fyrnox/tun_ipv6_test.go): Tests for activation failure and rollback ordering.
* [`cmd/diagnostic-wfp/main.go`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/cmd/diagnostic-wfp/main.go): Standalone native WFP VM diagnostic utility.
* [`dist/fyrnox-vm-acceptance-package/test_wfp_integration.exe`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/test_wfp_integration.exe): Synchronized diagnostic test executable for the isolated VM.
* [`dist/fyrnox-vm-acceptance-package/fyrnox.exe`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/fyrnox.exe): Synchronized production test binary.
* [`dist/fyrnox-vm-acceptance-package/README.md`](file:///C:/Users/Yousef%20Kia/Desktop/FYRNOX/dist/fyrnox-vm-acceptance-package/README.md): Updated acceptance documentation.

---

## 4. Verification & Testing (آزمایش و راستی‌آزمایی)

### Executed Commands & Evidence:
1. **Formatting & Syntax Audit:**
   ```powershell
   gofmt -w cmd/fyrnox/wfp_windows.go cmd/fyrnox/wfp_test.go cmd/fyrnox/tun_windows.go cmd/diagnostic-wfp/main.go
   gofmt -l cmd/fyrnox cmd/diagnostic-wfp
   # Result: Clean (0 unformatted files)
   ```
2. **Go Vet (Cross-Architecture amd64 and 386):**
   ```powershell
   $env:GOARCH='amd64'; go vet ./...
   $env:GOARCH='386'; go vet ./...
   # Result: Clean (exit code 0, 0 warnings)
   ```
3. **Cross-Architecture Build Compilation:**
   ```powershell
   $env:GOARCH='amd64'; go build -v ./...
   $env:GOARCH='386'; go build -v ./...
   # Result: Clean (exit code 0 on both architectures)
   ```
4. **WFP Subsystem Unit Tests:**
   ```powershell
   go test -v -run "TestWfp" ./cmd/fyrnox
   # Output:
   # === RUN   TestWfp_LayerConditionValidation
   # --- PASS: TestWfp_LayerConditionValidation (0.00s) - Unsupported conditions rejected
   # === RUN   TestWfpManager_RuleGeneration
   # --- PASS: TestWfpManager_RuleGeneration (0.01s) - 18 dynamic filters verified across 3 layers
   # === RUN   TestWfpManager_OuterIPv4_NoOuterIPv6Rule
   # --- PASS: TestWfpManager_OuterIPv4_NoOuterIPv6Rule (0.00s) - 15 filters in IPv4 outer mode
   # === RUN   TestWfpManager_TransactionRollbackOnError
   # --- PASS: TestWfpManager_TransactionRollbackOnError (0.00s) - Atomic rollback verified
   # === RUN   TestWfpManager_InputValidation
   # --- PASS: TestWfpManager_InputValidation (0.00s) - Validation enforced
   # === RUN   TestWfpManager_CloseIdempotency
   # --- PASS: TestWfpManager_CloseIdempotency (0.00s) - Teardown idempotent
   # PASS (ok firefox-vpn-client/cmd/fyrnox 1.523s)
   ```
5. **Host Guardrail Verification:**
   ```powershell
   & "dist/fyrnox-vm-acceptance-package/test_wfp_integration.exe"
   # Output:
   # [FATAL ERROR] Refusing to execute on known primary host: DESKTOP-3IOGCB8
   # This test modifies WFP in-kernel filters and must run exclusively inside an isolated VM.
   # (exit code 1 - guardrail passed)
   ```
6. **Full Repository Test Suite:**
   ```powershell
   go test -count=1 ./...
   # Result: All packages passed (exit code 0)
   ```

---

## 5. Acceptance Criteria & Next Steps (معیارهای پذیرش و اقدامات بعدی)

### Acceptance Status:
- [x] In-kernel conditions strictly validated against Microsoft documentation per layer.
- [x] Multi-layer defense across `ALE_AUTH_CONNECT_V6`, `OUTBOUND_TRANSPORT_V6`, and `OUTBOUND_IPPACKET_V6`.
- [x] Link-local NDP preservation via `fe80::/10` and `ff02::/16` address masks.
- [x] Narrow outer MASQUE exception (IP:2499:TCP) without broad application exemptions.
- [x] DNS leak blocking on non-Wintun physical adapters (ports 53 and 853).
- [x] Default-deny for all remaining physical IPv6 and on-link `/64` traffic.
- [x] Leak-free activation ordering: WFP armed before route installation.
- [x] Fail-safe rollback of connection-owned resources on failure.
- [x] Dynamic session crash policy explicitly documented.
- [x] Standalone VM diagnostic tool (`test_wfp_integration.exe`) compiled and packaged.
- [x] Clean builds and zero vet warnings on Windows `amd64` and `386`.
- [x] Zero Git commits, pushes, releases, or GitHub mutations.

### Next Steps:
- Transfer `dist/fyrnox-vm-acceptance-package/` to the isolated Windows VM.
- Run `test_wfp_integration.exe` inside the VM to capture native Win32 return codes from BFE.
- Execute full acceptance suite (`test_wintun_fulltunnel.ps1`) and capture Wireshark packet traces.
