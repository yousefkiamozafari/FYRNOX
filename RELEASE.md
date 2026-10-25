# Fyrnox for Windows v1.0.0 Release Notes

**توضیحات و یادداشت انتشار نسخه ۱.۰.۰ ویندوز Fyrnox**

---

## 🦊 درباره برنامه (About)
نسخه ویندوزی Fyrnox یک کلاینت غیررسمی، پرسرعت و پایدار برای استفاده از سرویس VPN رایگان فایرفاکس (Mozilla Guardian / Fastly MASQUE) در تمامی نسخه‌های سیستم‌عامل ویندوز است. این پروژه پورت‌شده و توسعه‌یافته از نسخه اندروید Fyrnox می‌باشد و برای استفاده همگانی بدون نیاز به خرید اشتراک یا پرداخت هزینه طراحی شده است.

- **Developer**: Powered by Yousef Kia ([GitHub](https://github.com/yousefkiamozafari))
- **Version**: `v1.0.0`
- **Supported OS**: Windows 11, Windows 10, Windows 8.1, Windows 7 (x64 / x86)

---

## 📦 دانلود فایل‌های نصبی و پرتابل (Downloads)

| فایل (Asset) | نوع (Type) | معماری (Arch) | حجم (Size) | سازگاری ویندوز |
| :--- | :--- | :--- | :--- | :--- |
| **[Fyrnox-Setup-v1.0.0.exe](https://github.com/yousefkiamozafari/Fyrnox-Windows/releases/download/v1.0.0/Fyrnox-Setup-v1.0.0.exe)** | **نصاب خودکار (Installer)** | 64-bit (x64) | ~14.7 MB | Windows 10 / 11 / 8.1 / 7 (64-bit) |
| **[Fyrnox-Setup-x86-v1.0.0.exe](https://github.com/yousefkiamozafari/Fyrnox-Windows/releases/download/v1.0.0/Fyrnox-Setup-x86-v1.0.0.exe)** | **نصاب خودکار (Installer)** | 32-bit (x86) | ~14.0 MB | تمامی نسخه‌های ۳۲ بیتی و قدیمی ویندوز |
| **[Fyrnox-v1.0.0-Windows-Portable.zip](https://github.com/yousefkiamozafari/Fyrnox-Windows/releases/download/v1.0.0/Fyrnox-v1.0.0-Windows-Portable.zip)** | **نسخه پرتابل (بدون نیاز به نصب)** | 64-bit (x64) | ~4.19 MB | تمامی سیستم‌ها (فقط Extract و اجرا) |
| **[Fyrnox-v1.0.0-Windows-x86-Portable.zip](https://github.com/yousefkiamozafari/Fyrnox-Windows/releases/download/v1.0.0/Fyrnox-v1.0.0-Windows-x86-Portable.zip)** | **نسخه پرتابل (بدون نیاز به نصب)** | 32-bit (x86) | ~4.15 MB | سیستم‌های ۳۲ بیتی |

---

## ✨ امکانات و قابلیت‌های کلیدی (Key Features)

1. **نصاب کاملاً مستقل و هوشمند ویندوز**:
   - ایجاد آیکون در دسکتاپ و منوی استارت (Start Menu).
   - ثبت رسمی در بخش Programs and Features ویندوز (Add/Remove Programs در Control Panel و Settings).
   - حذف کامل و هوشمند با Uninstaller اختصاصی به همراه قطع خودکار پروکسی و پاکسازی رجیستری.
   - عدم نیاز به مجوز Administrator به هنگام نصب (نصب اختصاصی و بدون ریسک در پوشه کاربر).

2. **پشتیبانی کامل از تمامی نسخه‌های ویندوز**:
   - کامپایل مجزا برای معماری‌های 64-bit و 32-bit.
   - سازگار با ویندوز 7، 8.1، 10 و 11.
   - اجرای کاملاً سایلنت و بدون پنجره مزاحم مشکی کنسول CMD.

3. **عبور از فیلترینگ با تکنولوژی تکه‌تکه کردن پکت (TLS Fragmentation)**:
   - پیاده‌سازی متدهای پیشرفته `sni` و `first_byte` دقیقاً مشابه نسخه ۱.۰.۷ اندروید برای دور زدن سیستم فیلترینگ شدید DPI (Deep Packet Inspection).

4. **تست پینگ فوق‌سریع و دقیق (Ultra-low Latency Ping)**:
   - محاسبه پینگ لحظه‌ای سرورها با هندشیک مستقیم TCP SYN/ACK و کش هوشمند DNS، با زمان پاسخ بسیار پایین (زیر ۱۰۰ میلی‌ثانیه).

5. **کارت نمایش زنده آی‌پی کاربر (Real-time IP Badge)**:
   - نمایش خودکار آی‌پی واقعی، پرچم کشور و وضعیت اتصال در صفحه اصلی برنامه.
   - تغییر لحظه‌ای وضعیت به "Protected / محافظت‌شده" همراه با پرچم کشور خروجی VPN پس از اتصال موفق.

6. **پشتیبانی از چند زبان و تم تاریک/روشن**:
   - سوئیچ سریع زبان بین **فارسی (Persian)** و **انگلیسی (English)** با چینش راست‌به‌چپ (RTL).
   - انتخاب تم **تاریک (Dark)** و **روشن (Light)** متناسب با سلیقه کاربر.

7. **تونل کل سیستم‌عامل (System-wide WinINet Tunnel)**:
   - هدایت خودکار تمام ترافیک وب و نرم‌افزارها از طریق پورت‌های محلی SOCKS5 (`127.0.0.1:20808`) و HTTP Proxy (`127.0.0.1:20809`).

8. **قابلیت TUN Mode برای کل سیستم (Full System TUN Mode)**:
   - هدایت کامل و ۱۰۰ درصدی کلیه بسته‌های شبکه سیستم‌عامل (شامل بازی‌ها، خط فرمان CMD/PowerShell، ابزارهای برنامه‌نویسی و ترافیک UDP) با یک کلیک بدون نشتی ترافیک.

9. **انیمیشن حرفه‌ای و مدرن دکمه اتصال (Cyber Sonar Radar & Glow)**:
   - امواج ضربانی چندلایه رادار (Sonar Ripple Waves)، هاله نورانی نئونی با تنفس ملایم (Aura Glow)، مدار چرخشی در هنگام اتصال (Orbit Spin) و بازخورد لمسی فنری.

10. **تست پینگ مجدد مستقیم از صفحه اصلی (Instant Re-ping on Click)**:
    - با کلیک روی نشانگر پینگ در کارت CURRENT LOCATION صفحه اصلی، پینگ سرور بلافاصله به صورت زنده اندازه‌گیری و با رنگ متناسب بروزرسانی می‌شود.

11. **بهینه‌سازی حداکثری سرعت و پهنای باند تمامی سرورها (High Throughput Pipeline)**:
    - بهره‌گیری از استخرهای بافر حافظه ۶۴ کیلوبایتی جهت حذف سربار GC، بافر انتقال ۲ مگابایتی و فعال‌سازی مستقیم فلگ‌های `TCP_NODELAY` جهت دستیابی به حداکثر سرعت دانلود و استریم.

---

## 🔐 کدهای اعتبارسنجی فایل‌ها (SHA-256 Checksums)

```text
76188CBA095AEBC8B282D1BC812CDE67DA1E8D8E5AA27DD74FD6F4E814DE9332  Fyrnox-Setup-v1.0.0.exe
76188CBA095AEBC8B282D1BC812CDE67DA1E8D8E5AA27DD74FD6F4E814DE9332  Fyrnox-Setup-Windows-11-10-x64.exe
B8BFE0D1F66C86B1D98E2B974BE5816A6B379F569DA9D9D7B08745002163C661  Fyrnox-Setup-Windows-7-8-10-x86.exe
76188CBA095AEBC8B282D1BC812CDE67DA1E8D8E5AA27DD74FD6F4E814DE9332  Fyrnox-Setup-x64-v1.0.0.exe
B8BFE0D1F66C86B1D98E2B974BE5816A6B379F569DA9D9D7B08745002163C661  Fyrnox-Setup-x86-v1.0.0.exe
B319C125CD2432DF53C1FFF54C56C465945C7525F63A32698E7EB7FC68E7DA06  Fyrnox-v1.0.0-Windows-Portable.zip
B319C125CD2432DF53C1FFF54C56C465945C7525F63A32698E7EB7FC68E7DA06  Fyrnox-v1.0.0-Windows-x64-Portable.zip
1E2437A915AA5A97101E688D13A1C4E219B769924E6BD5EDA962FD52DA616131  Fyrnox-v1.0.0-Windows-x86-Portable.zip
```

---

## 🚀 راهنمای گیت برای انتشار در گیت‌هاب (Publishing to GitHub)

اگر قصد دارید پروژه را در مخزن گیت‌هاب خود منتشر کنید، مراحل زیر را در ترمینال یا PowerShell اجرا فرمایید:

```powershell
# ۱. رفتن به پوشه پروژه
cd "Fyrnox-Windows"

# ۲. ساخت ریپازیتوری محلی گیت (در صورت نیاز)
git init
git add .
git commit -m "feat: release Fyrnox for Windows v1.0.0 with installer and multi-arch support"

# ۳. اتصال به مخزن گیت‌هاب شما
git remote add origin https://github.com/yousefkiamozafari/Fyrnox-Windows.git
git branch -M main
git push -u origin main

# ۴. ساخت تگ انتشار جهت فعال‌سازی اکشن اتوماتیک گیت‌هاب
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```
با ارسال تگ `v1.0.0`، سیستم **GitHub Actions** به صورت خودکار تمامی فایل‌های نصبی و پرتابل را کامپایل کرده و به عنوان یک Release رسمی در صفحه گیت‌هاب قرار می‌دهد. همچنین می‌توانید فایل‌های پوشه `build\dist` را مستقیماً به صفحه Release گیت‌هاب خود Drag & Drop نمایید.
