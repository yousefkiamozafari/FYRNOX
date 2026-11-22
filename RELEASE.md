# FYRNOX for Windows (v1.0.6) — Release Documentation
### راهنمای جامع و یادداشت‌های انتشار نسخه v1.0.6

[English Overview](#-english-release-notes) | [مستندات و راهنمای فارسی](#-مستندات-و-راهنمای-فارسی)

---

## ⚖️ Disclaimer
**FYRNOX is an independent open-source third-party client.** It is not affiliated with, endorsed by, or sponsored by Mozilla Corporation or Fastly, Inc. Firefox and Mozilla are registered trademarks of the Mozilla Foundation.

---

## 🇬🇧 English Release Notes

**FYRNOX v1.0.6** delivers complete remediation for Mozilla Guardian HTTP 401 authentication rejections, proactive OAuth access token expiration validation, error decoupling between HTTP 401 and HTTP 403, and full session error preservation.

### 🌟 Key Improvements in v1.0.6
1. **Proactive Access Token Validation:** Checks token validity with clock-skew safety buffer ($\ge 120s$) before dispatching `/api/v1/fpn/token` requests, preventing expired Bearer tokens from reaching Mozilla Guardian.
2. **Strict HTTP 401 vs HTTP 403 Decoupling:** Replaced ambiguous `ErrTokenInvalid` with dedicated sentinels `ErrReauthRequired` and `ErrNotEntitled`. Eliminates erroneous `ActivateGuardian` calls on rejected tokens.
3. **Controlled Single Refresh & Retry:** On HTTP 401, forces an atomic token refresh and retries proxy pass issuance at most once, preventing infinite retry loops.
4. **Transparent Session Expiration Propagation:** Surfaces true underlying FxA session expiration (`ErrReauthRequired`) directly to the user when tokens are revoked or expired.
5. **12 Hermetic Unit Tests:** Comprehensive test suite in `cmd/fyrnox/engine_auth_test.go` covering all authentication lifecycle paths.

### 📦 Distributable Binaries

| Package / Asset | Architecture | Target Operating Systems | Size |
| :--- | :---: | :--- | :---: |
| **`Fyrnox-Setup-v1.0.6.exe`** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB |
| **`Fyrnox-Setup-Windows-11-10-x64.exe`** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB |
| **`Fyrnox-Setup-Windows-7-8-10-x86.exe`** | 32-bit (`x86`) | Windows 7, 8, 8.1, 10 (32-bit) | ~19.8 MB |
| **`Fyrnox-v1.0.6-Windows-x64-Portable.zip`** | 64-bit (`x64`) | Standalone portable (No install required) | ~7.8 MB |
| **`Fyrnox-v1.0.6-Windows-x86-Portable.zip`** | 32-bit (`x86`) | Standalone portable 32-bit | ~7.8 MB |
| **`fyrnox.exe`** | 64-bit (`x64`) | Standalone direct binary | ~15.5 MB |

### 🔐 SHA-256 Checksums
```text
B8FD1D8CB67B0C6E8DC4D60F7EDB4B4EC661EB2D32D1BD76AE7F2E2826AAF5A5  Fyrnox-Setup-v1.0.6.exe
B8FD1D8CB67B0C6E8DC4D60F7EDB4B4EC661EB2D32D1BD76AE7F2E2826AAF5A5  Fyrnox-Setup-Windows-11-10-x64.exe
7562F6FC63D919B0255E2CC198B5F6C131D4881DFE5C758792011F42277DC527  Fyrnox-Setup-Windows-7-8-10-x86.exe
D25A61D1795CD892C0462C0322A13E28EA51C6492F87D253897F359071A1CDA9  Fyrnox-v1.0.6-Windows-x64-Portable.zip
ACDF721078EA0A553340441C415614D13A5722D73129B0369317F6A7CFA6EA49  Fyrnox-v1.0.6-Windows-x86-Portable.zip
C1E32406EA191C8FD86C30834F016E28E50138E63EB622A6955F71FD83A66252  fyrnox.exe
E5DA8447DC2C320EDC0FC52FA01885C103DE8C118481F683643CACC3220DAFCE  wintun.dll
```

---

## 🇮🇷 مستندات و راهنمای فارسی

نسخه **v1.0.6** کلاینت ویندوز FYRNOX، تمرکز کاملی بر رفع خطای ۴۰۱ گاردین موزیلا، تصحیح چرخه حیات و انقضای توکن‌های OAuth، و تفکیک دقیق خطاهای احراز هویت دارد.

### 🌟 مهم‌ترین بهبودهای نسخه v1.0.6:
1. **اعتبارسنجی پیش‌دستانه انقضای توکن:** قبل از ارسال هرگونه درخواست به سرور گاردین موزیلا، اعتبار زمانی توکن بررسی شده و در صورت نیاز با حاشیه امنیتی ۱۲۰ ثانیه‌ای به صورت خودکار تجدید می‌شود.
2. **تفکیک کامل خطای ۴۰۱ و ۴۰۳:** خطای ۴۰۱ (توکن نامعتبر) از خطای ۴۰۳ (عدم فعال‌سازی/عدم دسترسی) جدا شده و از تلاش بیهوده برای فعال‌سازی با توکن رد شده جلوگیری می‌گردد.
3. **تجدید توکن کنترل‌شده و دقیقاً یک‌باره:** پس از دریافت خطای ۴۰۱، نشست به صورت اجباری یک‌بار ریفرش شده و درخواست مجوز پروکسی بازآزمایی می‌شود؛ از لوپ‌های بی‌پایان جلوگیری می‌گردد.
4. **شفاف‌سازی پیام‌های خطا و نشست فایرفاکس:** در صورت ابطال یا انقضای نشست (`invalid_grant`)، خطای صریح ورود مجدد (`ErrReauthRequired`) به کاربر نمایش داده می‌شود.
5. **۱۲ تست یونیت هرمتیک جامع:** تمامی سناریوهای احراز هویت، انقضا، همزمانی و بازگشت از خطا پوشش داده شده‌اند.
