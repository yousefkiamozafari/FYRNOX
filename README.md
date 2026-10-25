# FYRNOX for Windows (v1.0.4)

[English](#-english-overview) | [فارسی](#-معرفی-و-مستندات-فارسی)

---

## 🇬🇧 English Overview

**FYRNOX** is a high-performance, subscription-free Windows desktop VPN client. It connects securely using your free **Firefox Account**, obtains an authentic proxy pass token from Mozilla's **Guardian service**, and tunnels your entire system traffic through **Fastly's MASQUE edge network** via HTTP/2 or HTTP/3 (QUIC) with advanced anti-censorship TLS ClientHello fragmentation.

[![Platform](https://img.shields.io/badge/Platform-Windows%2011%20%7C%2010%20%7C%208.1%20%7C%207-0078D6?style=for-the-badge&logo=windows&logoColor=white)](https://github.com/yousefkiamozafari/FYRNOX/releases)
[![Go](https://img.shields.io/badge/Backend-Go%201.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![UI](https://img.shields.io/badge/UI-Liquid%20Glass%20%7C%20Apple%20Style-FF6A00?style=for-the-badge)](https://github.com/yousefkiamozafari/FYRNOX)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**Developer & Maintainer:** [Yousef Kia](https://github.com/yousefkiamozafari)  
**Project Issue Tracker:** [GitHub Issues #1](https://github.com/yousefkiamozafari/FYRNOX/issues/1) | [Bilingual Issue Specs](docs/issues/ISSUE_v1.0.4_OVERHAUL.md)

---

### 📦 Distributable Binaries (v1.0.4)

| Edition | File | Architecture | Supported Operating Systems |
| :--- | :--- | :---: | :--- |
| **Setup Installer (Recommended)** | `Fyrnox-Setup-v1.0.4.exe` (~19.7 MB) | 64-bit (`x64`) | Windows 11 & Windows 10 (64-bit) |
| **Legacy Setup Installer** | `Fyrnox-Setup-Windows-7-8-10-x86.exe` (~18.8 MB) | 32-bit (`x86`) | Windows 7, 8.1, 10 (32-bit) |
| **Portable ZIP (x64)** | `Fyrnox-v1.0.4-Windows-x64-Portable.zip` (~7.4 MB) | 64-bit (`x64`) | Standalone portable (No install required) |
| **Portable ZIP (x86)** | `Fyrnox-v1.0.4-Windows-x86-Portable.zip` (~7.3 MB) | 32-bit (`x86`) | Standalone portable 32-bit |

---

### 🚀 Key Features in v1.0.4

* **Seamless Server Switching (Hot-Reconnect):** Switch between any of the 35 global server locations during an active connection without clicking "Disconnect" first. The client tears down old routes, purges the Windows resolver DNS cache (`DnsFlushResolverCache`), establishes the new tunnel, and verifies egress before marking as connected.
* **Liquid Glass (Apple-Grade) Dark Aesthetics:** Deep obsidian backdrop (`#0B0D12`), frosted glass layering with deep hardware-accelerated blur (`blur(20px)`), delicate glowing borders, and signature neon orange (`#FF6A00`) and amber gold accents.
* **Native Frameless Win32 Titlebar & Icon Binding:** Features custom minimize, maximize/restore, close buttons, double-click to maximize, and Aero Snap dragging. Binds high-resolution icons to the Windows taskbar and title bar via `SetCurrentProcessExplicitAppUserModelID` and `WM_SETICON`.
* **Complete System Proxy with Custom Bypass Rules & Remote DNS:** Routes all browser and Windows app traffic through `127.0.0.1:20809` (HTTP) and `127.0.0.1:20808` (SOCKS5). Applies local LAN bypass rules (`<local>;localhost;127.*;10.*;192.168.*`) and remote DNS resolution to prevent censorship leaks.
* **Kernel-Level Wintun Layer-3 TUN Mode:** Optional high-performance driver capturing 100% of all Windows IP/UDP packets, system services, and games.
* **Parallel Ping Worker Pool:** Tests all 35 edge nodes simultaneously with a concurrency limit of 6 and strict 3000ms timeout, updating individual latency badges live.
* **100% Persian (fa-IR) Localization & RTL Support:** Full native right-to-left layout, standard typography, mirrored chevrons/arrows, and comprehensive Persian translations for every setting and connection state.

---

## 🇮🇷 معرفی و مستندات فارسی

**فایرناکس (FYRNOX)** کلاینت دسکتاپ پرسرعت و رایگان ویندوز برای سرویس موزیلا وی‌پی‌ان است. این برنامه با اتصال مستقیم به **حساب فایرفاکس (Firefox Account)**، توکن سهمیه رایگان ۵۰ گیگابایت ماهانه را از سرویس **Mozilla Guardian** دریافت کرده و کل ترافیک را از طریق پروتکل‌های امن MASQUE روی شبکه ابری **Fastly** به صورت رمزنگاری‌شده عبور می‌دهد.

### 🚀 مهم‌ترین قابلیت‌های نسخه v1.0.4

1. ⚡ **سوییچ خودکار و اتصال مجدد بی‌وقفه (Seamless Hot-Reconnect):** در حین اتصال، با کلیک روی هر سرور دیگر، بدون نیاز به دیسکانکت دستی، اتصال قبلی بسته شده، کش دی‌ان‌اس ویندوز تخلیه و اتصال امن به سرور جدید برقرار می‌گردد.
2. 💎 **طراحی شیشه‌ای Liquid Glass اپل:** بوم فوق‌العاده زیبای ذغالی `#0B0D12` با لایه‌های نیمه‌شفاف شیشه‌ای، حاشیه‌های درخشان و رنگ‌های نارنجی نئونی و کهربایی.
3. 🦊 **تثبیت آیکون رسمی در تسک‌بار و تایتل‌بار ویندوز:** اتصال صریح آیکون باکیفیت روباه هندسی با Win32 AppUserModelID و ارسال پیام‌های `WM_SETICON`.
4. 🛡️ **هدایت کامل System Proxy همراه با Bypass Rules و ریموت DNS:** تنظیم دقیق رجیستری ویندوز، عبور مستقیم سایت‌های داخلی و عدم نشت دی‌ان‌اس با ارسال مستقیم دامنه‌ها به خروجی امن ریموت.
5. 🛡️ **حالت Wintun Layer-3 TUN:** درایور مجازی هسته ویندوز جهت هدایت ۱۰۰٪ تمام بسته‌ها، بازی‌ها و ترمینال‌های خط فرمان.
6. ⚡ **موتور تست پینگ موازی (Worker Pool):** سنجش همزمان ۳۵ سرور با به‌روزرسانی زنده برچسب تاخیر هر کارت (سبز، زرد، قرمز).
7. 🇮🇷 **پشتیبانی ۱۰۰٪ از زبان فارسی و چیدمان راست‌به‌چپ (RTL):** ترجمه روان تمام بخش‌ها، فرم‌ها، تنظیمات و تراز اصولی راست‌به‌چپ.

---

## 🛠️ نحوه بیلد و کامپایل از سورس (Build from Source)

پیش‌نیازها: [Go 1.22+](https://golang.org/dl/) بر روی سیستم‌عامل ویندوز.

```powershell
# ۱. کلون ریپازیتوری
git clone https://github.com/yousefkiamozafari/FYRNOX.git
cd FYRNOX

# ۲. اجرای اسکریپت ساخت تمام نسخه‌های نصبی و پرتابل
.\scripts\build_dist.ps1
```

تمام فایل‌های تولیدشده در پوشه `build/dist/` و ریشه پروژه قرار می‌گیرند.

---

## 📄 گزارش‌ها و مستندات توسعه (Documentation & Issues)

* **گزارش تغییرات دوزبانه (Bilingual Issue Report):** [`docs/issues/ISSUE_v1.0.4_OVERHAUL.md`](docs/issues/ISSUE_v1.0.4_OVERHAUL.md)
* **ردیاب گیت‌هاب (GitHub Issue Tracker):** [Issue #1](https://github.com/yousefkiamozafari/FYRNOX/issues/1)

---

## 📄 مجوز انتشار (License)
این نرم‌افزار تحت مجوز متن‌باز **MIT License** منتشر شده است.  
توسعه‌دهنده: **یوسف کیا (Yousef Kia)** — [github.com/yousefkiamozafari](https://github.com/yousefkiamozafari)
