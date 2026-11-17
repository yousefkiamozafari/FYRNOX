# FYRNOX for Windows (v1.0.5) — Release Documentation
### راهنمای جامع و یادداشت‌های انتشار نسخه v1.0.5

[English Overview](#-english-release-notes) | [مستندات و راهنمای فارسی](#-مستندات-و-راهنمای-فارسی)

---

## ⚖️ Disclaimer
**FYRNOX is an independent open-source third-party client.** It is not affiliated with, endorsed by, or sponsored by Mozilla Corporation or Fastly, Inc. Firefox and Mozilla are registered trademarks of the Mozilla Foundation.

---

## 🇬🇧 English Release Notes

**FYRNOX v1.0.5** represents a major reliability and anti-censorship hardening release. It introduces adaptive TLS ClientHello fragmentation with bounded timeout budgets, guaranteed OAuth refresh token preservation, hermetic test isolation, IPv6 test formatting fixes, and an aligned multi-architecture CI/release pipeline.

### 📦 Distributable Binaries

| Package / Asset | Architecture | Target Operating Systems | Size |
| :--- | :---: | :--- | :---: |
| **`Fyrnox-Setup-v1.0.5.exe`** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB |
| **`Fyrnox-Setup-Windows-11-10-x64.exe`** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB |
| **`Fyrnox-Setup-Windows-7-8-10-x86.exe`** | 32-bit (`x86`) | Windows 7, 8, 8.1, 10 (32-bit) | ~19.8 MB |
| **`Fyrnox-v1.0.5-Windows-x64-Portable.zip`** | 64-bit (`x64`) | Standalone portable (No install required) | ~7.8 MB |
| **`Fyrnox-v1.0.5-Windows-x86-Portable.zip`** | 32-bit (`x86`) | Standalone portable 32-bit | ~7.8 MB |
| **`fyrnox.exe`** | 64-bit (`x64`) | Standalone direct binary | ~15.5 MB |

### 🔐 SHA-256 Checksums
```text
27F1D50589783C24C8F7372C49646F38F8465D1B109CD74D678A07AC68BD3393  Fyrnox-Setup-v1.0.5.exe
27F1D50589783C24C8F7372C49646F38F8465D1B109CD74D678A07AC68BD3393  Fyrnox-Setup-Windows-11-10-x64.exe
C340FA7CAD232770096E27BA1F969E2E3F9581F7159CED7804ADC5930AA6717B  Fyrnox-Setup-Windows-7-8-10-x86.exe
19926E9BBF75006B05A9051C837DE8FCE22878D16011817AD84D15B2271B0798  Fyrnox-v1.0.5-Windows-x64-Portable.zip
392A15D883F9399B18B8A06172C6B0D4920E2552DFF00D151DCD6C7939E747AC  Fyrnox-v1.0.5-Windows-x86-Portable.zip
5F77FBD4158D0886D8865D77E57D8E51E33CF2F6A4EE815215B1F0F93CE8E497  fyrnox.exe
E5DA8447DC2C320EDC0FC52FA01885C103DE8C118481F683643CACC3220DAFCE  wintun.dll
```

---

## 🇮🇷 مستندات و راهنمای فارسی

نسخه **v1.0.5** کلاینت ویندوز FYRNOX، تمرکز ویژه‌ای بر پایداری ارتباط تحت شرایط شدید فیلترینگ و حملات DPI دارد. این نسخه با الگوریتم فرگمنتیشن تطبیقی TLS، از گیر کردن در مراحل اولیه اتصال جلوگیری کرده و چرخه حیات توکن‌ها را در برابر انقضای ناخواسته محافظت می‌نماید.

### 🌟 مهم‌ترین بهبودهای نسخه v1.0.5:
1. **فرگمنتیشن تطبیقی TLS (Adaptive TLS Fragmentation):** الگوریتم خودکار Fallback در دو مرحله ($T_1 = \min(4.5s, R/2, R-3s)$) جهت دور زدن اختلالات DPI در اتصال به لبه‌های Anycast فستلی.
2. **محافظت از نشست و رفرش‌توکن حساب کاربری:** حفظ کامل رفرش‌توکن موزیلا در چرخه تجدید خودکار Access Token و جلوگیری از خروج ناگهانی حساب کاربر.
3. **ایزولاسیون کامل تست‌های واحد و بازگشت از خطا:** تست‌های کاملاً محلی و سریع بدون نیاز به اتصال زنده اینترنت و پوشش تست بازگشت تمیز در زمان خطای احراز خروجی.
4. **رفع اخطار فرمت IPv6 در تحلیل ایستا:** انطباق کامل کدهای شبکه با استاندارد RFC 3986 از طریق `net.JoinHostPort`.
5. **یکپارچگی پایپ‌لاین CI و Release:** تنظیم خودکار نسخه Go به 1.27.4 در فرآیند بیلد و انتشار چند معماریه.
