## 🚀 Implementation Complete: 5 Flagship Desktop Modules & Advanced Networking (v1.0.4)
## 🚀 پیاده‌سازی کامل ۵ ماژول پرچمدار و قابلیت‌های پیشرفته شبکه (نسخه v1.0.4)

---

### 🇬🇧 English Release Update

We have completed the architecture, backend implementation, UI integration, unit tests, and production distribution builds for all 5 requested flagship modules:

1. **Windows Per-App Split Tunneling (Dual Mode Routing):**
   * Works across **both** Wintun Layer-3 TUN mode and WinINet System Proxy mode using Windows IP Helper API (`GetExtendedTcpTable`, `GetExtendedUdpTable`, `QueryFullProcessImageNameW`).
   * Supports both **Bypass Mode** (selected apps connect direct; all others tunneled) and **Proxy Mode** (only selected apps tunneled; remaining traffic direct).
   * Dedicated UI (`view-split-tunnel`) with real-time process scanner, quick presets (*Browsers*, *Telegram & Chat*, *Game Launchers*), and custom `.exe` binary path input.

2. **Dedicated Low-Latency Gaming Mode (Game Accelerator):**
   * Prioritizes raw UDP packets, enforces zero multiplexing (MUX), eliminates Head-of-Line blocking, and sets MTU to **1420** on Wintun to eliminate packet fragmentation.
   * Background telemetry loop continuously probes UDP latency, computing live Jitter (ms variance) and Packet Loss (%).

3. **Smart Iran / Domestic Traffic Direct Routing (Zero-Config Bypass):**
   * Embedded GeoSite mapping for `.ir`, `.ایران`, banking portals (Shaparak, Mellat, Tejarat, Melli, Pasargad, Saman, ZarinPal), and national platforms (Snapp, Digikala, Divar, Aparat).
   * In-memory CIDR lookup for 44 major Iranian Autonomous System (ASN) blocks.
   * Enabled by default to route domestic traffic directly via physical adapter without consuming VPN quota.

4. **Floating Mini-Widget HUD & Compact Win32 System Tray Monitor:**
   * Win32 System Tray (`Shell_NotifyIconW`) with rich right-click context menu (Connect/Disconnect, Proxy vs TUN mode, Gaming toggle, Restore window, Exit).
   * Compact translucent floating desktop HUD pill with live download/upload speeds (↓ / ↑ KB/s), latency, and one-click VPN toggle.

5. **Ultra-Fluid Liquid Glass UI (iOS-Grade Aesthetics & Full Persian RTL Symmetry):**
   * Deep 28px backdrop blur, continuous squircle geometry (16-24px radius), amber-gold accents, and 100% bidirectional Persian RTL mirroring.

---

### 🇮🇷 مستندات و گزارش تغییرات فارسی

پیاده‌سازی، یکپارچه‌سازی، تست‌های واحد و ساخت پکیج‌های نصاب برای تمامی ۵ ماژول پرچمدار با موفقیت به پایان رسید:

۱. **تونلینگ تفکیکی برنامه‌ها در ویندوز (Windows Per-App Split Tunneling):**
   * پشتیبانی کامل در هر دو حالت Wintun TUN و پراکسی سیستمی با استفاده از Windows IP Helper API (`GetExtendedTcpTable` و `GetExtendedUdpTable`).
   * پشتیبانی از دو حالت: **حالت دور زدن (Bypass)** برای عبور مستقیم برنامه‌های انتخابی و **حالت فقط پراکسی (Proxy Only)** برای عبور انحصاری برنامه‌های انتخابی از فیلترشکن.
   * صفحه اختصاصی در برنامه همراه با لیست زنده برنامه‌های فعال، دسته‌های آماده (مرورگرها، پیام‌رسان‌ها، لانچرهای بازی) و امکان افزودن دستی فایل exe.

۲. **شتاب‌دهنده اختصاصی گیمینگ (Low-Latency Gaming Mode):**
   * اولویت‌دهی به پکت‌های UDP خام، حذف سربار مالتی‌پلکسینگ (Zero MUX) و تنظیم خودکار MTU به **1420** بر روی آداپتور Wintun برای جلوگیری قطعی از تکه‌تکه‌شدن بسته‌ها.
   * تله‌متری مداوم پس‌زمینه برای سنجش دقیق تاخیر RTT، نوسان پینگ (Jitter) و درصد افت بسته (Packet Loss %).

۳. **مسیریابی مستقیم ترافیک داخلی ایران (Zero-Config Iran Bypass):**
   * شناسایی هوشمند دامنه‌های ir. و ایران، تمام درگاه‌های شاپرک، اینترنت‌بانک‌ها و پلتفرم‌های پرکاربرد به همراه ۴۴ رنج CIDR مربوط به اپراتورهای کشور.
   * اتصال مستقیم از اینترنت بدون افت سرعت و بدون کسر حجم سهمیه فیلترشکن.

۴. **ویجت شناور رومیزی (Floating Mini-Widget HUD) و آیکون تسک‌بار Win32:**
   * آیکون سیستم تری ویندوز (`Shell_NotifyIconW`) با منوی راست‌کلیک بومی Win32 برای تغییر وضعیت، تغییر حالت شبکه و خروج امن.
   * ویجت شیشه‌ای معلق و قابل جابجایی روی دسکتاپ با نمایش زنده سرعت دانلود و آپلود و پینگ لحظه‌ای.

۵. **ارتقای استایل Liquid Glass به استاندارد اپل و تقارن RTL فارسی:**
   * ارتقای شکست نور بلور به **28px**، انحنای پیوسته مربعی (Squircle)، درخشش کهربایی و چیدمان ۱۰۰٪ راست‌به‌چپ (RTL) همراه با هماهنگی کامل فونت‌های فارسی.

---

### 📦 Build Artifacts / فایل‌های خروجی کامپایل‌شده:
* `build/dist/Fyrnox-Setup-v1.0.4.exe` (19.8 MB) — Official Inno Setup Windows Installer
* `build/dist/Fyrnox-Setup-Windows-11-10-x64.exe` (19.8 MB) — 64-bit Installer
* `build/dist/Fyrnox-Setup-Windows-7-8-10-x86.exe` (18.8 MB) — 32-bit Installer
* `build/dist/Fyrnox-v1.0.4-Windows-x64-Portable.zip` (7.4 MB) — 64-bit Portable Bundle
* `build/dist/Fyrnox-v1.0.4-Windows-x86-Portable.zip` (7.4 MB) — 32-bit Portable Bundle
