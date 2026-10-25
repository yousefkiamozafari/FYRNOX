# [Design Spec] FYRNOX Brand-Tinted Liquid Glass & Live Structured Log Console
# [سند طراحی] طراحی شیشه‌ای Liquid Glass فایرناکس و کنسول لاگ زنده ساختاریافته

**Date / تاریخ:** 2026-10-25  
**Status / وضعیت:** Proposed / در انتظار تایید  
**Scope / دامنه:** `cmd/fyrnox/web/style.css`, `cmd/fyrnox/web/app.js`, `cmd/fyrnox/web/index.html`

---

## 🇬🇧 English Specification

### 1. Intent & Purpose
Transform the neutral, desaturated grey window glass into the true FYRNOX "Liquid Glass" visual aesthetic—warm, deeply tinted, and harmonized with the geometric flame fox brand palette (`#FF6A00` energetic orange and `#FFB000` warm amber). Simultaneously resolve the in-app live logging bug where log entries were serialized as `[object Object]` due to raw object array joins, converting the diagnostics view into a syntax-colored, real-time terminal monitor.

### 2. Goals, Non-Goals, and Anti-Goals
* **Goals:**
  * Implement deep smoky charcoal base tint (`rgba(11, 13, 18, 0.75)`) layered over native Windows Acrylic blur-behind.
  * Inject fixed subsurface radial gradient meshes bleeding brand orange (`#FF6A00`, opacity 0.08) and warm amber (`#FFB000`, opacity 0.04) through the blurred background.
  * Elevate cards with ambient radial light spots and glowing warm edges (`1px solid rgba(255, 106, 0, 0.25)`).
  * Ensure 100% crisp typography and contrast for all text, metrics, and interactive controls.
  * Fix in-app log rendering to cleanly parse `{timestamp, level, message}` entries with level badge coloring (`INFO`, `WARN`, `ERROR`, `DEBUG`).
  * Wire `clearLogs()` directly to backend `POST /api/logs/clear` and maintain automatic scrolling.
* **Non-Goals:**
  * Modifying underlying Win32 DWM backdrop calls in `window_windows.go` (already configured in Phase 3).
  * Altering networking routing engines (TUN/Proxy modes remain untouched).
* **Anti-Goals:**
  * Avoid neon over-saturation that causes visual fatigue or reduces text legibility.
  * Avoid rendering unescaped HTML inside the log console to prevent DOM injection.

### 3. Approaches Considered
* **Approach 1 (Selected): Layered Dual-Glow Liquid Glass + Structured Colored Log Stream**
  * Subsurface radial gradient mesh on the background canvas + glowing warm edges on cards.
  * Structured log stream parser with color-coded badges and true backend synchronization.
  * *Trade-off:* High visual cohesion across all pages with zero performance overhead.
* **Approach 2 (Rejected): Card-Only Local Glow + Raw Plaintext Logs**
  * Plain background with localized glows on cards only; uncolored text in log terminal.
  * *Trade-off:* Looks disjointed and leaves logs difficult to triage.

### 4. Implementation Details
* **Stylesheet (`cmd/fyrnox/web/style.css`):**
  * Update `--bg-app` to `rgba(11, 13, 18, 0.75)`.
  * Update `--border-subtle` and `--border-medium` to warm translucent tints (`rgba(255, 106, 0, 0.22)` and `rgba(255, 106, 0, 0.38)`).
  * Add multi-point radial mesh background to `.app-content-wrapper` and `body`.
  * Add ambient glow behind `#hero-connect-card` and `.telemetry-grid`.
  * Style log console lines with dedicated `.log-line`, `.log-time`, `.log-badge`, `.log-msg`.
* **Frontend Controller (`cmd/fyrnox/web/app.js`):**
  * Rewrite `renderLogs(logs)`: iterate over `LogEntry` objects, escape message text, build formatted DOM nodes or structured spans, and auto-scroll to bottom.
  * Rewrite `clearLogs()`: dispatch asynchronous `POST /api/logs/clear` and update UI line counter.

---

## 🇮🇷 مستندات و مشخصات فنی به زبان فارسی

### ۱. هدف و فلسفه طراحی
تبدیل افکت شیشه‌ای خنثی و خاکستری فعلی به استایل اصیل و اختصاصی **Liquid Glass برند فایرناکس**؛ همراه با تنالیته زغالی دودی عمیق و انتشار رنگ‌های گرم نارنجی نئونی (`#FF6A00`) و طلایی کهربایی (`#FFB000`) در عمق بلور شیشه. همزمان، رفع باگ نمایش لاگ‌ها در بخش عیب‌یابی (که به دلیل متصل کردن مستقیم آرایه اشیا به شکل `[object Object]` چاپ می‌شدند) و ارتقای آن به یک ترمینال زنده با تفکیک رنگی سطوح مختلف گزارش.

### ۲. اهداف، مرزها و ضد‌اهداف
* **اهداف پروژه:**
  * اعمال لایه پایه شیشه‌ای با رنگ زغالی دودی عمیق (`rgba(11, 13, 18, 0.75)`) بر روی افکت آکریلیک بومی ویندوز.
  * تزریق گرادیان‌های زیرسطحی ملایم با رنگ‌های نارنجی آتشین (`#FF6A00` با شفافیت 0.08) و کهربایی (`#FFB000` با شفافیت 0.04) در پس‌زمینه.
  * ایجاد کادرهای درخشان و ظریف شیشه‌ای با تینت گرم (`1px solid rgba(255, 106, 0, 0.25)`).
  * تعبیه هاله‌های نور رادیال در پشت دکمه مرکزی اتصال و کارت‌های داشبورد.
  * تضمین خوانایی فوق‌العاده متون، کنتراست شفاف و عدم تار شدگی نوشته‌ها.
  * فرمت‌بندی استاندارد لاگ‌ها به شکل `[15:04:05] [INFO] Message` با رنگ‌بندی تفکیک‌شده (سبز، زرد، قرمز).
  * هماهنگی دکمه پاکسازی لاگ با اندپوینت بک‌اند `/api/logs/clear` و اسکرول هوشمند به انتهای خطوط.
* **موارد خارج از دامنه (Non-Goals):**
  * تغییر در لایه ران‌تایم DWM در فایل `window_windows.go` (پیش‌تر در فاز ۳ پیاده‌سازی شده است).
  * دستکاری در موتورهای روتینگ و هسته تونل شبکه.
* **ضد‌اهداف (Anti-Goals):**
  * جلوگیری از اشباع رنگی بیش از حد (Oversaturation) که سبب خستگی چشم کاربر شود.
  * جلوگیری از آسیب‌پذیری تزریق HTML در پنجره لاگ‌ها با استفاده از Escape استاندارد.

### ۳. رویکردهای بررسی‌شده
* **رویکرد ۱ (تایید شده): سیستم نوری دو لایه (Layered Dual-Glow) + مانیتور لاگ ساختاریافته رنگی**
  * ایجاد اتمسفر یکپارچه در کل پنجره به همراه هایلایت‌های موضعی کارت‌ها.
  * فرمت تمیز خطوط لاگ با برچسب‌های رنگی استاندارد و پاکسازی واقعی در حافظه.
* **رویکرد ۲ (رد شده): نورپردازی محدود به کارت‌ها + نمایش لاگ به صورت متن ساده خام**
  * عدم ایجاد جلوه منسجم برند در کل فریم برنامه و سختی خوانایی لاگ‌ها در حالت تک‌رنگ.

---

### ۴. اقدامات اجرایی و فایل‌های هدف
1. [`cmd/fyrnox/web/style.css`](file:///C:/Users/LouiNoor/.gemini/antigravity/scratch/FoxyVPN-Windows/cmd/fyrnox/web/style.css): اعمال متغیرهای رنگی جدید، گرادیان مش پس‌زمینه، کادرهای درخشان نارنجی و استایل ترمینال لاگ.
2. [`cmd/fyrnox/web/app.js`](file:///C:/Users/LouiNoor/.gemini/antigravity/scratch/FoxyVPN-Windows/cmd/fyrnox/web/app.js): بازنویسی تابع `renderLogs` با پارسر آبجکت، رنگ‌بندی برچسب‌ها و اصلاح تابع `clearLogs` با متد `POST`.
3. [`cmd/fyrnox/web/index.html`](file:///C:/Users/LouiNoor/.gemini/antigravity/scratch/FoxyVPN-Windows/cmd/fyrnox/web/index.html): تنظیم المان‌های کمکی هاله‌های نوری در صورت نیاز.
