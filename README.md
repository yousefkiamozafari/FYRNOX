# FYRNOX for Windows (v1.0.4)

[English Overview](#-english-overview) | [معرفی و مستندات فارسی](#-معرفی-و-مستندات-فارسی)

---

[![Platform](https://img.shields.io/badge/Platform-Windows%2011%20%7C%2010%20%7C%208.1%20%7C%207-0078D6?style=for-the-badge&logo=windows&logoColor=white)](https://github.com/yousefkiamozafari/FYRNOX/releases)
[![Architecture](https://img.shields.io/badge/Architecture-x64%20%7C%20x86-0078D6?style=for-the-badge)](https://github.com/yousefkiamozafari/FYRNOX/releases)
[![Go Backend](https://img.shields.io/badge/Backend-Go%201.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![UI](https://img.shields.io/badge/UI-Liquid%20Glass%20%7C%20Apple%20Grade-FF6A00?style=for-the-badge)](https://github.com/yousefkiamozafari/FYRNOX)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

> [!NOTE]
> **Third-Party Open-Source Notice:**  
> FYRNOX is an independent open-source desktop client. It is **not** an official product of Mozilla Corporation, Mozilla Foundation, or Fastly, Inc. FYRNOX is not sponsored, authorized, or endorsed by Mozilla. "Firefox" and "Mozilla" are registered trademarks of the Mozilla Foundation.

---

## 🇬🇧 English Overview

**FYRNOX** is a high-performance, subscription-free Windows desktop VPN client. It connects securely using your free **Firefox Account**, obtains an authentic proxy pass token from Mozilla's **Guardian service**, and tunnels your entire network traffic through **Fastly's MASQUE edge network** via HTTP/2 or HTTP/3 (QUIC) with advanced anti-censorship TLS ClientHello fragmentation.

### 🌟 Key Highlights in v1.0.4
* **Dual Routing Modes with Self-Healing Fallback:** Supports kernel-level Layer-3 Wintun TUN mode for gaming, command-line tools, and full UDP coverage, paired with automatic graceful fallback to System Proxy (WinINet + WinHTTP) when running without Administrator elevation.
* **Seamless Hot-Reconnect:** Switch between any of the 35 global edge locations instantly during an active connection. The client tears down old routes, purges the Windows DNS cache (`DnsFlushResolverCache`), negotiates the new tunnel, and verifies egress before marking as connected.
* **Apple-Grade Liquid Glass Aesthetics:** Deep obsidian canvas (`#0B0D12`), frosted glass blur (`backdrop-filter: blur(20px)`), glowing accents, floating dock, and an interactive 3D Fibonacci dot-matrix globe (26,000 points) with zero GPU overhead when idle or in background.
* **Full Multi-Architecture Coverage:** Dedicated 64-bit (`x64`) and 32-bit (`x86`) native executables and self-contained installers for Windows 11, 10, 8.1, and 7.
* **100% Persian (fa-IR) RTL Localization:** Complete right-to-left mirrored interface, standard Persian typography, and full translations for all connection phases and diagnostics.
* **Zero Telemetry & Private by Design:** No tracking SDKs, no developer servers, and credentials are saved locally in `%USERPROFILE%\.firefox-vpn-tokens.json` with owner-only permissions (`0600`). See [PRIVACY.md](PRIVACY.md).

---

### 📦 Download Packages (v1.0.4)

| Package | Architecture | Supported Operating Systems | Size | Download |
| :--- | :---: | :--- | :---: | :--- |
| **Setup Installer (Recommended)** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB | [Download](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.0.4/Fyrnox-Setup-v1.0.4.exe) |
| **Windows 11 / 10 Installer (Mirror)** | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) | ~20.6 MB | [Download](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.0.4/Fyrnox-Setup-Windows-11-10-x64.exe) |
| **Legacy Windows Installer** | 32-bit (`x86`) | Windows 7, 8, 8.1, 10 (32-bit) | ~19.8 MB | [Download](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.0.4/Fyrnox-Setup-Windows-7-8-10-x86.exe) |
| **Portable ZIP (x64)** | 64-bit (`x64`) | Standalone portable (No install required) | ~7.8 MB | [Download](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.0.4/Fyrnox-v1.0.4-Windows-x64-Portable.zip) |
| **Portable ZIP (x86)** | 32-bit (`x86`) | Standalone portable 32-bit | ~7.8 MB | [Download](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.0.4/Fyrnox-v1.0.4-Windows-x86-Portable.zip) |
| **Direct Executable** | 64-bit (`x64`) | Direct standalone binary | ~15.5 MB | [Download](https://github.com/yousefkiamozafari/FYRNOX/releases/download/v1.0.4/fyrnox.exe) |

---

### 🛠️ Architecture & Data Flow

```
┌────────────────────────────────────────────────────────────────────────┐
│                        FYRNOX Frontend (UI)                            │
│  Liquid Glass Dark Canvas | 3D Fibonacci Globe | Real-time Telemetry   │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ IPC / Local API
┌───────────────────────────────────▼────────────────────────────────────┐
│                       Core Engine (Go 1.22+)                           │
│  Atomic Pipeline | Token Caching | DNS Flush | Auto-Fallback Guard     │
├───────────────────────────────────┬────────────────────────────────────┤
│  Routing Mode A: Wintun L3        │  Routing Mode B: System Proxy      │
│  - Virtual TUN Driver (x86/x64)   │  - Local SOCKS5 (Port 20808)       │
│  - gVisor Userspace Netstack      │  - Local HTTP Proxy (Port 20809)   │
│  - SOCKS5 UDP Associate / DoH     │  - WinINet System Registry Hooks   │
└───────────────────────────────────┴────────────────────────────────────┘
                                    │ TLS ClientHello Fragmentation
┌───────────────────────────────────▼────────────────────────────────────┐
│                    Mozilla Guardian & Fastly MASQUE                    │
│      35 Global Edge Locations (HTTP/2 / HTTP/3 QUIC Egress)            │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 🇮🇷 معرفی و مستندات فارسی

**فایرناکس (FYRNOX)** یک کلاینت دسکتاپ مدرن، پرسرعت، امن و کاملاً رایگان برای سیستم‌عامل ویندوز (ویندوز ۷ تا ۱۱) است. این نرم‌افزار با اتصال مستقیم به **حساب فایرفاکس (Firefox Account)**، توکن سهمیه رسمی ۵۰ گیگابایت ماهانه را از سرویس **Mozilla Guardian** دریافت کرده و تمام ترافیک سیستم را با پروتکل‌های MASQUE روی شبکه ابری **Fastly** به صورت رمزنگاری‌شده عبور می‌دهد.

### 🚀 مهم‌ترین امکانات نسخه v1.0.4:
1. ⚡ **سوییچ خودکار و اتصال مجدد بی‌وقفه (Seamless Hot-Reconnect):** تغییر مکان خروجی میان ۳۵ سرور با یک کلیک بدون نیاز به قطع دستی و بدون نشت ترافیک.
2. 💎 **طراحی لوکس Liquid Glass اپل:** تم ذغالی عمیق، لایه‌های شیشه‌ای نیمه‌شفاف با بلور سخت‌افزاری، داک شناور و کره زمین ۳ بعدی هولوگرافیک بدون اعوجاج قاره‌ها.
3. 🛡️ **دو حالت هدایت ترافیک با سیستم خودترمیمی:** پشتیبانی از تانل کامل لایه ۳ ویندوز (Wintun TUN) با سوئیچ خودکار به پروکسی سیستم در صورت عدم دسترسی ادمین.
4. 🦊 **پشتیبانی کامل از معماری‌های ۳۲ بیتی و ۶۴ بیتی:** پکیج‌های مستقل و بهینه برای ویندوزهای ۷، ۸، ۸.۱، ۱۰ و ۱۱.
5. 🇮🇷 **پشتیبانی ۱۰۰٪ از زبان فارسی:** چینش استاندارد راست‌به‌چپ (RTL)، تایپوگرافی اصولی و ترجمه روان تمام بخش‌ها.

---

## 🔧 عیب‌یابی و نکات مهم (Troubleshooting)

1. **دسترسی Administrator برای حالت TUN:**
   * حالت تانل لایه ۳ (Wintun) برای ساخت کارت شبکه مجازی در هسته ویندوز نیازمند اجرای برنامه به صورت `Run as Administrator` است.
   * اگر برنامه به صورت معمولی اجرا شود، هیچ مشکلی پیش نمی‌آید و سیستم خودترمیمی FYRNOX به طور خودکار به حالت **System Proxy** سوئیچ می‌کند تا وب‌گردی و ارتباط شما پایدار بماند.

2. **هشدار فایروال یا آنتی‌ویروس:**
   * به دلیل استخراج خودکار فایل رسمی درایور کرネル وایروگارد (`wintun.dll`) در پوشه برنامه، ممکن است برخی از آنتی‌ویروس‌ها هشدار دهند. این فایل نسخه رسمی امضاشده مایکروسافت/وایروگارد است و هیچ خطری ندارد.

3. **خروج ناگهانی و تنظیمات پروکسی:**
   * در صورت خروج اضطراری، برنامه به صورت خودکار با فراخوانی `RestoreSystemProxy()` تنظیمات پروکسی سیستم‌عامل را ریست می‌کند تا ارتباط مستقیم اینترنت شما هرگز مسدود نماند.

---

## 🛠️ ساخت و کامپایل از سورس (Build from Source)

پیش‌نیازها: [Go 1.22+](https://golang.org/dl/) بر روی سیستم‌عامل ویندوز.

```powershell
# ۱. کلون ریپازیتوری
git clone https://github.com/yousefkiamozafari/FYRNOX.git
cd FYRNOX

# ۲. اجرای اسکریپت ساخت تمام نسخه‌های نصبی و پرتابل
.\scripts\build_dist.ps1
```

---

## 📚 مستندات پروژه و سیاست‌ها (Documentation & Policies)

* **[LICENSE](LICENSE):** متن کامل پروانه متن‌باز MIT.
* **[SECURITY.md](SECURITY.md):** فرآیند گزارش آسیب‌پذیری و افشای مسئولانه.
* **[PRIVACY.md](PRIVACY.md):** سند شفافیت حریم خصوصی و عدم جمع‌آوری داده.
* **[CONTRIBUTING.md](CONTRIBUTING.md):** راهنمای مشارکت در توسعه و استانداردهای کدنویسی.
* **[CHANGELOG.md](CHANGELOG.md):** تاریخچه کامل تغییرات نسخه‌ها بر اساس استاندارد Keep a Changelog.
* **[RELEASE.md](RELEASE.md):** یادداشت‌های فنی انتشار نسخه v1.0.4.

---

**توسعه‌دهنده و نگه‌دارنده:** [یوسف کیا (Yousef Kia)](https://github.com/yousefkiamozafari)  
تحت مجوز متن‌باز **MIT License**.
