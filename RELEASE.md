# FYRNOX for Windows (v1.0.4) — Release Documentation
### راهنمای جامع و یادداشت‌های انتشار نسخه v1.0.4

[English Overview](#-english-release-notes) | [مستندات و راهنمای فارسی](#-مستندات-و-راهنمای-فارسی)

---

## ⚖️ Disclaimer
**FYRNOX is an independent open-source third-party client.** It is not affiliated with, endorsed by, or sponsored by Mozilla Corporation or Fastly, Inc. Firefox and Mozilla are registered trademarks of the Mozilla Foundation.

---

## 🇬🇧 English Release Notes

**FYRNOX v1.0.4** represents a major engineering and aesthetic overhaul of the Windows desktop client. This release introduces multi-architecture Wintun kernel driver embedding, self-healing routing fallback to System Proxy, an Apple-grade Liquid Glass dark interface with an interactive 3D Fibonacci globe, seamless server hot-reconnect, and complete Persian (fa-IR) RTL localization.

### 📦 Distributable Binaries

| Package / Asset | Architecture | Target Operating Systems | Size |
| :--- | :---: | :--- | :---: |
| **`Fyrnox-Setup-v1.0.4.exe`** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB |
| **`Fyrnox-Setup-Windows-11-10-x64.exe`** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB |
| **`Fyrnox-Setup-Windows-7-8-10-x86.exe`** | 32-bit (`x86`) | Windows 7, 8, 8.1, 10 (32-bit) | ~19.8 MB |
| **`Fyrnox-v1.0.4-Windows-x64-Portable.zip`** | 64-bit (`x64`) | Standalone portable (No install required) | ~7.8 MB |
| **`Fyrnox-v1.0.4-Windows-x86-Portable.zip`** | 32-bit (`x86`) | Standalone portable 32-bit | ~7.8 MB |
| **`fyrnox.exe`** | 64-bit (`x64`) | Standalone direct binary | ~15.5 MB |

### 🔐 SHA-256 Checksums
```text
600158CF0BF57105CBD9E9B86EE09A0DA83A543CB862DD457B842E9D20ABEC4F  Fyrnox-Setup-v1.0.4.exe
600158CF0BF57105CBD9E9B86EE09A0DA83A543CB862DD457B842E9D20ABEC4F  Fyrnox-Setup-Windows-11-10-x64.exe
B44BE7664AE52926BF499DCFC3D3D638A74631037AD40454DAF5311133666B3E  Fyrnox-Setup-Windows-7-8-10-x86.exe
9C42C936802A7A265D6C249B83527E7AD4B47E2736092E1F4B77EDB236E5AB8E  Fyrnox-v1.0.4-Windows-x64-Portable.zip
BE02A2FC62EB883FA77543B34F3407D1D7B6D0987B77FC677CDE7E095CBAD403  Fyrnox-v1.0.4-Windows-x86-Portable.zip
594DE6790C21FBE1FBC30018A2CCE33880B03716666A8B3310DCD30B27469DFB  fyrnox.exe
```

---

## 🇮🇷 مستندات و راهنمای فارسی

نسخه **v1.0.4** کلاینت ویندوز FYRNOX، پایدارترین و کامل‌ترین نسخه منتشرشده تا این لحظه است. این نسخه با هدف ارائه یک ارتباط پرسرعت، امن و ضد سانسور، بدون نیاز به خرید اشتراک و با تکیه بر ۵۰ گیگابایت سهمیه ماهانه رسمی موزیلا گاردین و شبکه ابری Fastly MASQUE توسعه داده شده است.

### 🌟 مهم‌ترین بهبودهای نسخه v1.0.4:
1. **جاسازی درایورهای ۳۲ بیتی و ۶۴ بیتی Wintun:** تفکیک درایورهای کرネル بر اساس معماری سیستم (`x64`, `x86`, `arm64`) جهت جلوگیری قطعی از خطای کرش ۱۹۳.
2. **مکانیزم بازیابی و فال‌بک خودکار (Self-Healing Fallback):** در صورت عدم اجرای برنامه به عنوان Administrator، برنامه کاربر را قطع نمی‌کند، بلکه به صورت خودکار به حالت System Proxy تغییر وضعیت می‌دهد.
3. **طراحی Liquid Glass و کره ۳ بعدی فیبوناچی:** رابط کاربری مدرن با تم ذغالی، داک شیشه‌ای معلق، انیمیشن‌های اسلایدی روان و مصرف صفر GPU در حالت پس‌زمینه.
4. **سوییچ آنی میان ۳۵ سرور (Hot-Reconnect):** جابجایی سریع میان موقعیت‌ها بدون قطعی دستی، همراه با تخلیه کش DNS سیستم‌عامل.
5. **پشتیبانی ۱۰۰٪ دو زبانه و چیدمان استاندارد RTL:** ترجمه روان تمام بخش‌ها، فرم‌ها و تنظیمات به زبان فارسی.
