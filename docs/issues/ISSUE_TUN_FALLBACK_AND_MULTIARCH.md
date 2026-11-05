# [Issue #17] Wintun Multi-Arch Dynamic Embedding, Proxy Fallback & Non-Admin Auto-Recovery
# [ایشو شماره ۱۷] تفکیک معماری درایور Wintun، مکانیزم بازیابی پروکسی و هدایت خودکار در نبود دسترسی ادمین

---

## 🇬🇧 English Documentation

### 1. Overview
This issue addresses stability, routing deadlocks, driver loading failures, and non-admin privileges when utilizing the Windows Layer-3 Wintun TUN mode in FYRNOX. When encountering system-level restrictions or missing driver components, the engine must never disrupt active user connections or cause deadlocks, but instead recover seamlessly.

---

### 2. Root Cause Analysis & Technical Solutions

#### A. Non-Admin TUN Route Fallback
* **Problem:** In `cmd/fyrnox/tun_windows.go`, when a standard user (without Administrator privileges) started the client with TUN mode enabled, the check `if !IsAdmin()` logged a warning, set `tunModeActive = true`, and returned `nil`. Consequently, `engine.go` treated TUN configuration as successful, skipped the fallback to `SetSystemProxyWithBypass`, and left browser traffic unproxied without active routes.
* **Resolution (`cmd/fyrnox/tun_windows.go`):** Converted the non-admin condition into an explicit error: `return fmt.Errorf("kernel TUN mode requires administrator privileges")`. This triggers the fallback pipeline in `engine.go`, immediately enabling system-wide WinINet proxy with user bypass rules while maintaining the active connection.

#### B. Proxy Deadlock Elimination on TUN Initialization Failure
* **Problem:** When `EnableTunMode` returned an error, the previous logic in `engine.go` immediately closed the local SOCKS5 and HTTP listeners while leaving the Windows system proxy pointing to those dead local ports, resulting in a complete internet outage.
* **Resolution (`cmd/fyrnox/engine.go`):** The engine now keeps active local proxy listeners intact on TUN failure, safely activates `SetSystemProxyWithBypass`, updates the in-memory state to `UseSystemProxy = true`, and logs the fallback. If fallback also fails, `RestoreSystemProxy()` cleanly restores direct connection.

#### C. Multi-Architecture Wintun Driver Embedding (x64 / x86 / ARM64)
* **Problem:** A single 64-bit `wintun.dll` was embedded into the binary. Running the client on 32-bit systems caused crash 193 (`ERROR_BAD_EXE_FORMAT`).
* **Resolution (`cmd/fyrnox/wintun_embed_*.go`):** Segregated drivers via Go build tags:
  * `wintun_embed_amd64.go`: embeds official 64-bit driver (427 KB).
  * `wintun_embed_386.go`: embeds official 32-bit driver (550 KB).
  * `wintun_embed_arm64.go`: embeds official ARM64 driver (222 KB).

#### D. Dynamic Packaging Script (`scripts/build_dist.ps1`)
* **Problem:** Hardcoded user profile paths (`C:\Users\LouiNoor\...`) broke automated builds on other workstations.
* **Resolution:** Replaced with dynamically resolved paths based on `$PSScriptRoot`, ensuring cross-machine portability.

---

### 3. Verification & Automated Tests
* Unit tests in `cmd/fyrnox/wintun_embed_test.go` validate:
  1. DOS MZ magic header (`0x4D`, `0x5A`).
  2. PE signature offset and `PE\0\0`.
  3. Machine architecture matches `runtime.GOARCH` (`0x8664` for x64, `0x014C` for x86).
  4. Runtime disk extraction byte integrity.

---

## 🇮🇷 مستندات فارسی

### ۱. معرفی و هدف
این ایشو به منظور رفع خطاهای بارگذاری درایور، جلوگیری از بن‌بست پروکسی (Proxy Deadlock) و رفع مشکل اجرای برنامه بدون دسترسی ادمین در حالت Wintun TUN در کلاینت FYRNOX ایجاد و پیاده‌سازی شده است. در صورت بروز هرگونه محدودیت سیستمی، کلاینت نباید اینترنت کاربر را قطع کند و باید به صورت خودکار به پایدارترین حالت سوئیچ نماید.

---

### ۲. تحلیل ریشه‌ای و راه‌حل‌های مهندسی

#### الف. سوئیچ خودکار در غیاب دسترسی ادمین (Non-Admin Auto-Recovery)
* **ریشه مشکل:** در فایل `cmd/fyrnox/tun_windows.go`، اگر کاربر بدون دسترسی Administrator اقدام به اتصال در حالت TUN می‌کرد، شرط `if !IsAdmin()` متغیر `tunModeActive = true` را ست کرده و بدون خطا (`nil`) برمی‌گشت. در نتیجه، موتور برنامه (`engine.go`) تصور می‌کرد تانل فعال شده و مرحله فال‌بک به پروکسی سیستم را اجرا نمی‌کرد که منجر به عدم هدایت ترافیک مرورگرها می‌شد.
* **راه‌حل (`cmd/fyrnox/tun_windows.go`):** تبدیل این وضعیت به یک خطای صریح:
  `return fmt.Errorf("kernel TUN mode requires administrator privileges")`  
  با این کار، تابع `engine.go` بلافاصله خطا را دریافت کرده و بدون وقفه تنظیمات پروکسی سیستم (`SetSystemProxyWithBypass`) را فعال و کاربر را متصل نگه می‌دارد.

#### ب. رفع بن‌بست پروکسی و قطعی کامل اینترنت (Proxy Deadlock)
* **ریشه مشکل:** در صورت بروز هرگونه خطا در استارتاپ Wintun، کد قبلی سرورهای محلی SOCKS5 و HTTP را خاموش می‌کرد در حالی که رجیستری ویندوز همچنان به این پورت‌ها اشاره داشت که باعث قطع شدن وب‌گردی می‌شد.
* **راه‌حل (`cmd/fyrnox/engine.go`):** سرورهای محلی دیگر خاموش نمی‌شوند و برنامه بلافاصله با فال‌بک نرم، پروکسی ویندوز را فعال کرده و وضعیت اتصال را حفظ می‌کند.

#### ج. تفکیک معماری درایورهای Wintun (نسخه‌های ۳۲ و ۶۴ بیتی)
* **ریشه مشکل:** تعبیه تکی فایل ۶۴ بیتی موجب بروز خطای ۱۹۳ (`ERROR_BAD_EXE_FORMAT`) در سیستم‌های ۳۲ بیتی می‌شد.
* **راه‌حل (`cmd/fyrnox/wintun_embed_*.go`):** تفکیک فایل‌ها با استفاده از تگ‌های ساخت Go (`//go:build`):
  * درایور ۶۴ بیتی (`amd64` با حجم ۴۲۷ کیلوبایت)
  * درایور ۳۲ بیتی (`386` با حجم ۵۵۰ کیلوبایت)
  * درایور پردازنده‌های `arm64` (با حجم ۲۲۲ کیلوبایت)

#### د. داینامیک‌سازی اسکریپت پکیجینگ (`scripts/build_dist.ps1`)
* **ریشه مشکل:** هاردکد بودن مسیر کاربر قبلی که مانع از اجرای بیلد در سیستم‌های دیگر می‌شد.
* **راه‌حل:** داینامیک‌سازی کامل مسیرها بر پایه مکان فایل اسکریپت و فولدرهای کش داخلی پروژه.

---

### ۳. اعتبارسنجی و تست‌ها
* تست‌های خودکار در `cmd/fyrnox/wintun_embed_test.go` صحت هدرهای باینری PE، شناسه معماری ماشین و استخراج صحیح فایل‌های DLL را با موفقیت تضمین می‌کنند.
