# [Release v1.0.4] Core Reliability Fixes, Seamless Server Switching, Win32 Icon Binding, Liquid Glass UI & Persian RTL Localization
# [نسخه v1.0.4] ارتقای جامع پایداری، سوییچ بدون‌وقفه سرورها، بایندینگ آیکون ویندوز، طراحی Liquid Glass و بومی‌سازی فارسی RTL

---

## 🇬🇧 English Documentation

### 1. Overview
The FYRNOX Windows desktop client (v1.0.4) has undergone a comprehensive engineering and design overhaul. This update eliminates critical connection lifecycle issues, resolves Windows proxy routing and DNS leakage, introduces seamless automatic reconnection during server switching, ensures high-resolution geometric fox icon display across the Windows taskbar and frameless title bar, overhauls the UI to Apple-grade Liquid Glass aesthetics, and provides 100% complete Persian (fa-IR) localization with full RTL layout mirroring.

---

### 2. Resolved Issues & Engineering Implementations

#### A. System Proxy Routing, Bypass Rules & Remote DNS
* **Issue:** Enabling System Proxy mode previously failed to alter the public IP address across open browsers or broke local network services.
* **Fix (`winproxy_windows.go`, `http_proxy.go`, `socks_proxy.go`):**
  * Registry configuration strictly updates `ProxyEnable = 1` and registers `http=127.0.0.1:20809;https=127.0.0.1:20809;socks=127.0.0.1:20808`.
  * LAN bypass rules (`FormatBypassRules`) inject `<local>;localhost;127.*;10.*;172.16.*;192.168.*` along with user-customized domains into `ProxyOverride`.
  * WinINet settings are applied immediately via `InternetSetOptionW` (`INTERNET_OPTION_SETTINGS_CHANGED` and `INTERNET_OPTION_REFRESH`) followed by a system-wide `WM_SETTINGCHANGE` broadcast.
  * Direct dial fallback (`isBypassTarget`) handles internal domains (`.local`, `.lan`, `.home`, `.internal`) locally, preventing 502 Bad Gateway responses.
  * Standard internet requests tunnel remote domain names directly to Fastly MASQUE edge nodes without local Windows DNS resolution, preventing ISP DNS hijacking and censorship leaks.

#### B. Seamless Hot-Reconnect on Server Switching
* **Issue:** Switching servers while connected required manually clicking "Disconnect", leaving the interface stuck or routes colliding.
* **Fix (`engine.go`, `tun_windows.go`, `app.js`):**
  * When a new server is clicked during an active connection (`StateConnected` or `StateConnecting`), the client triggers an immediate non-blocking switch.
  * The clicked server card reflects `Connecting...` immediately on the frontend.
  * The previous route is cleanly torn down, host routing loops are cleared via `cleanupHostRoutes()`, and the Windows DNS cache is thoroughly purged via Win32 `DnsFlushResolverCache` (`dnsapi.dll`) and `ipconfig /flushdns`.
  * The new outbound tunnel is negotiated and verified via end-to-end `VerifyExit` probe before flipping status to `Connected`.

#### C. Windows Taskbar & Titlebar Icon Binding
* **Issue:** The Windows taskbar and frameless title bar displayed generic host icons or blurred placeholders due to WebView2 grouping.
* **Fix (`main.go`, `window_windows.go`):**
  * Prior to window initialization, `SetCurrentProcessExplicitAppUserModelID` is invoked with `FYRNOX.Client.App.v1`.
  * Multi-resolution `.ico` assets (16x16 up to 256x256) are embedded directly into PE resource ID 1 via `.syso`.
  * Explicit Win32 `WM_SETICON` messages are sent to the window HWND for both `ICON_BIG` (32x32) and `ICON_SMALL` (16x16).

#### D. Apple-Style Liquid Glass Redesign & Global Ping Engine
* **Aesthetics (`style.css`):**
  * Deep obsidian canvas `#0B0D12` paired with layered frosted glass panels (`backdrop-filter: blur(20px)`), ultra-subtle borders (`rgba(255, 255, 255, 0.08)`), and specular inner highlights.
  * Brand neon orange `#FF6A00` and amber gold `#FFB000` accents.
* **High-Speed Ping Worker Pool (`app.js`):**
  * Concurrency-bounded Worker Pool (limit 6) measures live latency via TCP + TLS handshakes with a strict 3000ms timeout.
  * Real-time DOM updates color-code latency pills per card: Fast Green (<200ms), Medium Amber (200-500ms), and Slow Red (>500ms).

#### E. 100% Persian (fa-IR) Localization & RTL Support
* **Complete Localization (`index.html`, `app.js`, `style.css`):**
  * All labels, cards, egress parameters, telemetry metrics, diagnostics cards, settings groups, and modal forms are annotated with `data-i18n` and `data-i18n-placeholder`.
  * Expanded dictionary (`I18N.fa`) with 60+ terms covers connection stages (`authorizing`, `connecting_upstream`, `starting_proxies`, `verifying_exit`, `configuring_routes`).
  * Enforced `[dir="rtl"]` layout with Vazirmatn typography, mirrored routing arrows and chevrons, right-to-left sidebar layout, and isolated `direction: ltr` for monospace IP addresses, durations, and numbers.

---

### 3. Verification & Live Test Results

| Test Case | Method | Result | Evidence |
| :--- | :--- | :---: | :--- |
| **System Proxy IP Verification** | `test_live_system_proxy.ps1` | **PASS** | Baseline IP `45.61.148.221` switched to tunnel egress `149.143.187.79` |
| **Registry & WinINet Sync** | Registry inspection | **PASS** | `ProxyEnable=1` with multiprotocol string; restored to `0` upon disconnect |
| **Server Switching Hot-Reconnect** | `TestSelectServerHotReconnect` | **PASS** | Automated switch and reconnect verified in 0.05s |
| **DNS Cache Purging** | `FlushDNSCache()` Win32 call | **PASS** | `DnsFlushResolverCache` + `ipconfig /flushdns` executed on every mode/server toggle |
| **Ping Worker Pool** | `/api/ping-all` live run | **PASS** | 35/35 global edge POPs verified without timeout or UI freezing |
| **Go Package Regression Suite** | `go test -v ./cmd/fyrnox` | **PASS** | 21/21 unit & integration tests passed |

---

## 🇮🇷 مستندات و گزارش فارسی

### ۱. خلاصه تغییرات
کلاینت ویندوز FYRNOX نسخه v1.0.4 تحت یک بازنویسی و ارتقای عمیق مهندسی و بصری قرار گرفت. در این نسخه، مشکلات چرخه حیات اتصال و قفل شدن وضعیت Disconnect برطرف شد، مسیردهی پراکسی سیستمی ویندوز و عدم تغییر IP اصلاح گردید، امکان سوییچ خودکار و بی‌وقفه سرورها (Seamless Hot-Reconnect) اضافه شد، اتصال آیکون باکیفیت روباه هندسی به تسک‌بار و پنجره ویندوز تثبیت شد، رابط کاربری با الهام از زبان طراحی Liquid Glass اپل بازطراحی گردید و زبان فارسی (fa-IR) به صورت ۱۰۰٪ همراه با استایل راست‌به‌چپ (RTL) پیاده‌سازی شد.

---

### ۲. شرح اقدامات و راهکارهای فنی

#### الف) اصلاح مسیردهی System Proxy، قوانین Bypass و ریموت DNS
* **رفع مشکل مسیردهی:** تنظیم قطعی `ProxyEnable = 1` و فرمت چندپروتکلی `http=127.0.0.1:20809;https=127.0.0.1:20809;socks=127.0.0.1:20808` در رجیستری ویندوز.
* **قوانین دور زدن شبکه محلی (Bypass Rules):** ثبت خودکار رنج‌های محلی (`<local>;localhost;127.*;10.*;172.16.*;192.168.*`) به همراه دامنه‌های سفارشی کاربر در `ProxyOverride`.
* **بروزرسانی بدون تاخیر سیستم:** فراخوانی مستقیم APIهای `wininet.dll` (`INTERNET_OPTION_SETTINGS_CHANGED` و `INTERNET_OPTION_REFRESH`) به همراه انتشار پیام سیستمی `WM_SETTINGCHANGE` به پنجره‌ها و مرورگرها بدون نیاز به راه‌اندازی مجدد سیستم.
* **حفاظت در برابر نشت DNS:** ارسال مستقیم نام دامنه‌ها به خروجی امن Fastly MASQUE بدون تفکیک DNS محلی، جهت جلوگیری از فیلترینگ و مسموم‌سازی دی‌ان‌اس توسط ارائه‌دهندگان اینترنت.

#### ب) اتصال مجدد هوشمند و خودکار هنگام تغییر سرور (Seamless Hot-Reconnect)
* **سوییچ آنی در حالت اتصال فعال:** با کلیک روی سرور جدید در حین اتصال، دکمه سرور بلافاصله به وضعیت «در حال اتصال...» تغییر یافته و فرآیند اتصال مجدد به صورت کاملاً ناهمگام (Non-blocking) اجرا می‌شود.
* **تخلیه کامل کش DNS و پاکسازی مسیرها:** حذف مسیرهای پیشین از جدول روتینگ سیستم، تخلیه کش ریزالور ویندوز با فراخوانی مستقیم تابع Win32 به نام `DnsFlushResolverCache` و اجرای `ipconfig /flushdns`.
* **اعتبارسنجی ترافیک لایه ۳:** بررسی خروجی داده‌ها از طریق هوک `VerifyExit` قبل از اعلام وضعیت متصل.

#### ج) بایندینگ آیکون تسک‌بار و پنجره برنامه با Win32 API
* **شناسه انحصاری پروسه (`AppUserModelID`):** فراخوانی `SetCurrentProcessExplicitAppUserModelID` با شناسه `FYRNOX.Client.App.v1` در لحظه استارت اپلیکیشن برای جداسازی پنجره از پروسه‌های عمومی میزبان WebView2.
* **ارسال پیام‌های `WM_SETICON`:** ارسال پیام‌های Win32 برای آیکون بزرگ (`ICON_BIG`) به اندازه ۳۲×۳۲ پیکسل و آیکون کوچک (`ICON_SMALL`) به اندازه ۱۶×۱۶ پیکسل با استفاده از ریسورس‌های چندرزولوشنه تعبیه‌شده PE.

#### د) بازطراحی Liquid Glass اپل و تست پینگ موازی
* **ظاهر شیشه‌ای دودی مدرن:** بوم تیره `#0B0D12` با افکت بلور عمیق (`blur(20px)`)، خطوط حاشیه ظریف شیشه‌ای (`rgba(255, 255, 255, 0.08)`) و درخشش نارنجی نئونی (`#FF6A00`) و طلایی کهربایی (`#FFB000`).
* **موتور تست پینگ موازی (Worker Pool با ظرفیت ۶):** سنجش همزمان و کنترل‌شده تاخیر ۳۵ سرور با مهلت پاسخ‌دهی اکید ۳۰۰۰ میلی‌ثانیه و نمایش بلادرنگ وضعیت روی کارت هر سرور (سبز <200ms، زرد 200-500ms، قرمز >500ms).

#### ه) بومی‌سازی کامل زبان فارسی (fa-IR) با پشتیبانی جامع راست‌به‌چپ (RTL)
* **پوشش ۱۰۰٪ المان‌های صفحه:** تجهیز تمام عنوان‌ها، کارت‌ها، متریک‌های تله‌متری، لاگ‌ها، فرم‌های ورود و تایید دو مرحله‌ای به `data-i18n`.
* **دیکشنری جامع:** واژه‌نامه تخصصی فارسی با بیش از ۶۰ کلید واژه برای تمام مراحل پنج‌گانه اتصال و پیام‌های هشدار.
* **چیدمان استاندارد RTL:** تنظیم جهت نمایش با فونت استاندارد، معکوس‌سازی جهت فلش‌ها و چورون‌ها و ایزوله‌سازی نمایش اعداد، زمان‌ها و آی‌پی‌ها با `direction: ltr`.

---

---

### 4. Flagship Desktop Modules & Advanced Networking Features (v1.0.4 Update)

#### Module 1: Windows Per-App Split Tunneling (Dual Mode)
* **Architecture (`split_tunnel_windows.go`):**
  * Operates across **BOTH** Wintun Layer-3 TUN mode and WinINet System Proxy mode without requiring third-party kernel drivers.
  * Native socket inspection via Windows IP Helper API: calls `GetExtendedTcpTable` and `GetExtendedUdpTable` from `iphlpapi.dll` to query owner process IDs (PIDs) mapped to local socket ports or destination target endpoints.
  * In Proxy mode, client port maps directly to PID. In Wintun TUN mode, `globalPortCache` resolves outgoing TCP/UDP connections against system socket tables to extract process binary names via `QueryFullProcessImageNameW`.
  * **Dual Routing Modes:**
    * *Bypass Mode (Blacklist):* Selected applications connect directly to the internet; all other apps and system traffic route securely through the VPN tunnel.
    * *Proxy Mode (Whitelist):* Only selected applications route through the VPN tunnel; remaining traffic connects directly.
  * **Frontend UI (`view-split-tunnel`):**
    * Live searchable list of active running processes and installed applications.
    * Instant presets: *Browsers* (Chrome, Edge, Firefox, Brave, Opera), *Telegram & Chat* (Telegram, Discord, WhatsApp), *Game Launchers* (Steam, Epic Games, Riot, Battle.net).
    * Custom `.exe` binary path input.

#### Module 2: Dedicated Low-Latency Gaming Mode (Game Accelerator)
* **Architecture (`gaming_mode.go`):**
  * Enforces pure UDP packet prioritization with zero multiplexing (MUX) overhead to eliminate Head-of-Line blocking.
  * Dynamically optimizes MTU to **1420** on the Wintun adapter (`netsh interface ipv4 set subinterface "FyrnoxTun" mtu=1420 store=persistent`) to completely prevent packet fragmentation across congested networks.
  * Continuous background UDP telemetry probe calculating live round-trip latency (RTT), jitter (moving variance in ms), and packet loss percentage (%).
  * Real-time dashboard telemetry badge with animated pulse effects.

#### Module 3: Smart Iran / Domestic Traffic Direct Routing (Zero-Config Bypass)
* **Architecture (`iran_bypass.go`):**
  * Embedded database of Iranian top-level domains (`.ir`, `.ایران`), major domestic banking portals (Shaparak, Mellat, Tejarat, Melli, Pasargad, Saman, ZarinPal, etc.), and national services (Snapp, Digikala, Divar, Aparat, Rubika, Bale).
  * In-memory CIDR lookup against 44 major Iranian Autonomous System (ASN) IP blocks covering millions of domestic IP addresses.
  * Enabled by default: all domestic traffic bypasses the VPN tunnel directly via the physical network interface without consuming VPN quota or incurring international routing latency.

#### Module 4: Floating Mini-Widget HUD & Compact Win32 System Tray Monitor
* **Win32 System Tray (`tray_windows.go`):**
  * Integrated notification area icon via `Shell_NotifyIconW` with multi-state tooltip (`Connected`, `Connecting`, `Disconnected`).
  * Native Win32 popup context menu (`TrackPopupMenuEx`) allowing users to Connect/Disconnect, switch between System Proxy and Wintun TUN, toggle Gaming Accelerator, restore the main window, or cleanly exit.
* **Floating Desktop Mini-Widget HUD (`index.html`, `style.css`, `app.js`):**
  * Translucent, draggable compact HUD overlay on the desktop.
  * Live upload and download speed meters (↓ / ↑ KB/s), tunnel latency indicator, connection status glow, and one-click quick toggle.

#### Module 5: Ultra-Fluid Liquid Glass UI (iOS-Grade UX)
* **Styling Upgrades (`style.css`):**
  * Upgraded backdrop blur from 20px to **28px** for deeper optical frosted-glass dispersion.
  * Apple-style continuous squircles (16px - 24px corner radius) with ultra-thin translucent borders (`rgba(255, 255, 255, 0.08)`) and warm amber glow.
  * Complete bidirectional LTR and RTL mirroring for seamless Persian user experience.

---

### ۴. ماژول‌های پرچمدار و قابلیت‌های پیشرفته شبکه (بخش دوم آپدیت v1.0.4)

#### ماژول ۱: تونلینگ تفکیکی برنامه‌ها در ویندوز (Windows Per-App Split Tunneling)
* **معماری فنی (`split_tunnel_windows.go`):**
  * کارکرد همزمان در **هر دو حالت** Wintun L3 TUN و پراکسی سیستمی ویندوز (System Proxy) بدون نیاز به درایورهای متفرقه.
  * بازرسی سوکت‌ها با فراخوانی APIهای بومی IP Helper ویندوز (`GetExtendedTcpTable` و `GetExtendedUdpTable` از `iphlpapi.dll`) و تطبیق پورت سوکت با شناسه پروسه (PID) و نام فایل اجرایی از طریق `QueryFullProcessImageNameW`.
  * **دو حالت مجزا:**
    * *حالت دور زدن (Bypass / لیست سیاه):* برنامه‌های انتخابی به صورت مستقیم وصل شده و بقیه ترافیک سیستم از فیلترشکن عبور می‌کند.
    * *حالت فقط پراکسی (Proxy Only / لیست سفید):* تنها برنامه‌های انتخابی از فیلترشکن رد شده و باقی ترافیک سیستم دست‌نخورده می‌ماند.
  * **رابط کاربری (`view-split-tunnel`):**
    * فهرست زنده و قابل جستجوی برنامه‌های در حال اجرا و برنامه‌های نصب‌شده سیستم.
    * دسته‌های آماده: مرورگرها، پیام‌رسان‌ها (تلگرام و دیسکورد) و لانچرهای بازی (استیم، اپیک‌گیمز).
    * امکان افزودن دستی هر فایل اجرایی `.exe`.

#### ماژول ۲: شتاب‌دهنده اختصاصی گیمینگ (Low-Latency Gaming Mode)
* **معماری فنی (`gaming_mode.go`):**
  * اولویت‌دهی به پکت‌های UDP خام بدون مالتی‌پلکسینگ (MUX) جهت حذف سربار تاخیر صف (Head-of-Line blocking).
  * تنظیم خودکار و آنی MTU به مقدار **1420** بر روی آداپتور Wintun برای جلوگیری قطعی از تکه‌تکه‌شدن بسته‌ها (Zero Packet Fragmentation).
  * حلقه پایش و تله‌متری مداوم پس‌زمینه برای سنجش دقیق تاخیر RTT، نوسان پینگ (Jitter بر حسب میلی‌ثانیه) و درصد افت بسته (Packet Loss %).

#### ماژول ۳: مسیریابی مستقیم و هوشمند ترافیک ایران (Zero-Config Iran Bypass)
* **معماری فنی (`iran_bypass.go`):**
  * پایگاه داده داخلی شامل تمام دامنه‌های ir. و ایران، درگاه‌های پرداخت شاپرک، اینترنت‌بانک‌های ملی، ملت، تجارت، صادرات، پاسارگاد، سامان، زرین‌پال و سرویس‌های پرکاربرد (اسنپ، تپسی، دیجی‌کالا، دیوار، آپارات، روبیکا و بله).
  * بررسی در حافظه رنج‌های IP بیش از ۴۴ بلوک CIDR مربوط به اپراتورها و ASهای داخل کشور.
  * فعال به صورت پیش‌فرض: اتصال مستقیم از اینترنت محلی بدون افت سرعت و بدون کسر ترافیک فیلترشکن.

#### ماژول ۴: ویجت شناور رومیزی (Floating Mini-Widget HUD) و مانیتور تسک‌بار Win32
* **آیکون ناحیه اعلانات ویندوز (Tray Icon در `tray_windows.go`):**
  * ادغام در سیستم تری ویندوز با `Shell_NotifyIconW` و نمایش تولتیپ بر اساس وضعیت اتصال.
  * منوی راست‌کلیک بومی Win32 (`TrackPopupMenuEx`) برای اتصال/قطع، سوییچ بین حالت پراکسی و TUN، فعال‌سازی شتاب‌دهنده گیمینگ و بازکردن داشبورد.
* **ویجت مینی مانیتور شناور رومیزی (`floating-mini-hud`):**
  * ویجت شیشه‌ای فشرده، قابل جابجایی (Draggable) با نمایش زنده سرعت دانلود و آپلود (↓ / ↑ KB/s)، پینگ لحظه‌ای و دکمه اتصال سریع.

#### ماژول ۵: ارتقای استایل Liquid Glass به استاندارد اپل و تقارن RTL
* **جلوه‌های بصری (`style.css`):**
  * ارتقای میزان شکست نور بلور از 20px به **28px** برای عمق شیشه‌ای فوق‌العاده.
  * انحنای پیوسته مربعی (Squircle با شعاع ۱۶ تا ۲۴ پیکسل)، حاشیه‌های ترنسلوسنت شیشه‌ای و درخشش کهربایی.
  * تقارن و چیدمان ۱۰۰٪ راست‌به‌چپ (RTL) همراه با هماهنگی کامل فونت‌ها.

---

---

### 6. Full Desktop Acrylic Glassmorphism & Interactive Latency Probe (v1.0.4 Phase 3)

#### A. Native Windows Desktop Acrylic & Transparent Canvas Overhaul
* **Win32 DWM Backdrop Integration (`window_windows.go`):**
  * Invokes `DwmExtendFrameIntoClientArea` with negative margins (`-1, -1, -1, -1`) to extend the desktop window manager blur surface across the entire client canvas.
  * Sets `DWMWA_SYSTEMBACKDROP_TYPE = 3` (`DWMSBT_TRANSIENTWINDOW` / Desktop Acrylic) via `DwmSetWindowAttribute` for true Gaussian desktop blur-behind on Windows 11 and Windows 10 (22H2+). Includes automatic fallback to `SetWindowCompositionAttribute` (`ACCENT_ENABLE_ACRYLICBLURBEHIND`).
  * Queries `ICoreWebView2Controller2` via COM interface pointer inspection and calls `PutDefaultBackgroundColor(COREWEBVIEW2_COLOR{A: 0, R: 0, G: 0, B: 0})`. This turns the Chromium WebView2 canvas completely transparent, allowing the native Windows Acrylic frosted background to visibly shine through.
* **Liquid Glass CSS Design System (`style.css`):**
  * Configured `html, body { background: transparent !important; }` and translucent container tints (`rgba(11, 13, 18, 0.40)` to `rgba(18, 22, 32, 0.45)`).
  * Upgraded backdrop blur to `blur(30px) saturate(160%)` for deep, fluid optical distortion.
  * Translucent borders (`1px solid rgba(255, 255, 255, 0.12)`) and ambient soft shadows (`0 8px 32px 0 rgba(0, 0, 0, 0.37)`).

#### B. Interactive Ping Badge & Live Latency Re-Test
* **Interactive Trigger & Real-Time Animation (`index.html`, `style.css`, `app.js`):**
  * Dashboard latency card (`#dash-ping-card`) and refresh icon (`#telem-ping-refresh`) are directly clickable to initiate immediate latency measurements.
  * Displays active testing animations: card pulse glow (`pulseCardGlow`), refresh button 360° spin (`spinRefresh`), and icon scaling (`pulseIconScale`).
  * In active tunnel mode, queries `/api/ping?tunnel=true` to measure live round-trip layer-3 latency through Fastly MASQUE. In disconnected mode, measures direct TCP/TLS handshake latency to the selected server.
* **Dynamic Visual Latency Classification:**
  * `< 150ms`: Fast / Optimal (Green `#22C55E`, `.latency-good`)
  * `150ms - 350ms`: Moderate (Amber `#EAB308`, `.latency-medium`)
  * `> 350ms` or Timeout / Offline: Poor (Red `#EF4444`, `.latency-poor`)
  * Fully localized state labels across English and Persian (`Measuring...` / `در حال سنجش...`, `Timeout` / `پایان مهلت`, `Failed` / `ناموفق`, `Offline` / `آفلاین`).

---

### ۶. افکت شیشه‌ای آکریلیک بومی ویندوز و تست تعاملی تاخیر پینگ (فاز ۳ آپدیت v1.0.4)

#### الف) پیاده‌سازی افکت Desktop Acrylic و ترنسپرنت‌سازی بوم WebView2
* **اتصال به DWM در سطح Win32 (`window_windows.go`):**
  * گسترش فریم پنجره به تمام فضای کلاینت از طریق تابع `DwmExtendFrameIntoClientArea` با مارجین منفی (`-1, -1, -1, -1`).
  * فعال‌سازی افکت اکریلیک بومی ویندوز ۱۱ و ۱۰ با فراخوانی `DwmSetWindowAttribute` و فلگ `DWMWA_SYSTEMBACKDROP_TYPE = 3` (حالت Acrylic بلور پشت پنجره) به همراه فال‌بک خودکار به `SetWindowCompositionAttribute`.
  * ارتباط مستقیم با اینترفیس COM مربوط به `ICoreWebView2Controller2` و تنظیم آلفای پس‌زمینه وب‌ویو به عدد ۰ (`PutDefaultBackgroundColor` با رنگ کاملاً شفاف)، که باعث می‌شود بلور دسکتاپ بومی ویندوز به صورت بلادرنگ از پشت المنت‌های رابط کاربری نمایان شود.
* **طراحی Glassmorphism شیشه‌ای (`style.css`):**
  * شفاف‌سازی کامل زمینه (`background: transparent !important`) با رنگ مایع شیشه‌ای ملایم (`rgba(11, 13, 18, 0.40)`).
  * اعمال شکست نور عمیق ۳۰ پیکسلی (`backdrop-filter: blur(30px) saturate(160%)`)، حاشیه‌های درخشان کریستالی (`1px solid rgba(255, 255, 255, 0.12)`) و سایه‌های نرم و عمیق.

#### ب) نشانگر تعاملی پینگ و تست مجدد آنی روی داشبورد اصلی
* **تریگر تعاملی و انیمیشن‌های پویا (`index.html`, `style.css`, `app.js`):**
  * کارت پینگ داشبورد (`#dash-ping-card`) و دکمه بازخوانی (`#telem-ping-refresh`) با یک کلیک تست زنده تاخیر را اجرا می‌کنند.
  * اعمال انیمیشن‌های چرخش ۳۶۰ درجه آیکون رفرش (`spinRefresh`)، تپش نوری کارت (`pulseCardGlow`) و بزرگ‌نمایی ملایم آیکون رعد (`pulseIconScale`).
  * ارسال درخواست به اندپوینت `/api/ping`؛ در حالت اتصال فعال، تاخیر واقعی رفت‌وبرگشت تونل (Tunnel RTT) و در حالت قطع، تاخیر هندشیک TCP/TLS به سرور انتخاب‌شده اندازه‌گیری می‌شود.
* **رنگ‌بندی هوشمند کیفیت تاخیر:**
  * کمتر از ۱۵۰ میلی‌ثانیه: عالی و سریع (رنگ سبز `#22C55E`)
  * ۱۵۰ تا ۳۵۰ میلی‌ثانیه: متوسط و قابل قبول (رنگ کهربایی `#EAB308`)
  * بیش از ۳۵۰ میلی‌ثانیه یا تایم‌اوت: ضعیف یا خطا (رنگ قرمز `#EF4444`)
  * ترجمه کامل و دوزبانه وضعیت‌ها به زبان‌های فارسی و انگلیسی.

---

### ۷. فایل‌های خروجی نهایی و پکیج‌های کامپایل‌شده (Final Verified Distribution Artifacts)

| نام فایل | نوع بسته | پلتفرم | حجم | وضعیت تست |
| :--- | :--- | :--- | :--- | :---: |
| **Fyrnox.exe** | باینری ریشه برنامه | Windows x64 | 14.5 MB | **تایید شده (PASS)** |
| **Fyrnox-Setup-v1.0.4.exe** | نصاب رسمی ویندوز | Windows x64 | 19.8 MB | **تایید شده (PASS)** |
| **Fyrnox-Setup-Windows-11-10-x64.exe** | نصاب آماده توزیع | Windows x64 | 19.8 MB | **تایید شده (PASS)** |
| **Fyrnox-Setup-Windows-7-8-10-x86.exe** | نصاب آماده توزیع | Windows x86 | 18.8 MB | **تایید شده (PASS)** |
| **Fyrnox-v1.0.4-Windows-x64-Portable.zip** | نسخه پرتابل فشرده | Windows x64 | 7.4 MB | **تایید شده (PASS)** |
| **Fyrnox-v1.0.4-Windows-x86-Portable.zip** | نسخه پرتابل فشرده | Windows x86 | 7.4 MB | **تایید شده (PASS)** |

