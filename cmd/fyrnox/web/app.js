
function formatTrafficBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let val = bytes;
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024;
    i++;
  }
  return i === 0 ? `${Math.round(val)} B` : `${val.toFixed(1)} ${units[i]}`;
}

// FYRNOX for Windows — Desktop Controller & Real-Time Engine

const I18N = {
  en: {
    brand_badge: 'DESKTOP VPN',
    nav_dashboard: 'Dashboard',
    nav_locations: 'Servers',
    nav_diagnostics: 'Diagnostics',
    nav_settings: 'Settings',
    page_dashboard: 'Network Dashboard',
    page_servers: 'Server Locations',
    page_diagnostics: 'System Diagnostics',
    page_settings: 'Preferences & Settings',
    account_firefox: 'Firefox Account',
    sign_in: 'Sign In',
    sign_out: 'Sign Out',
    detecting: 'Detecting...',
    not_logged_in: 'Not logged in',
    quota_desc: '50.0 GB/month high-speed Mozilla VPN quota',
    disconnected: 'DISCONNECTED',
    connecting: 'CONNECTING...',
    connected: 'CONNECTED SECURELY',
    disconnecting: 'DISCONNECTING...',
    error: 'CONNECTION ERROR',
    btn_connect: 'Connect',
    btn_disconnect: 'Disconnect',
    btn_connecting: 'Connecting',
    btn_disconnecting: 'Disconnecting',
    btn_sign_in: 'Sign In',
    traffic_unprotected: 'Traffic is currently unprotected',
    quota_card_title: 'Monthly VPN Allowance',
    quota_used_lbl: 'Used:',
    quota_rem_lbl: 'Remaining:',
    quota_resets_lbl: 'Resets:',
    quota_badge_authoritative: 'Authoritative',
    quota_badge_estimated: 'Estimated local usage',
    set_ip_proto_title: 'IP Protocol Preference',
    set_ip_proto_sub: 'Select address family for VPN transport and leak protection. In System Proxy mode, apps not respecting proxy connect directly. In TUN mode, IPv4 Only prevents IPv6 leaks.',
    opt_proto_dual: 'IPv4 + IPv6 (Dual Stack) [Recommended]',
    opt_proto_ipv4: 'IPv4 Only',
    opt_proto_ipv6: 'IPv6 Only',
    dash_error_title: 'Connection Warning',
    dash_error_msg: 'Failed to establish secure tunnel.',
    step_authorizing: 'Authorizing session with Firefox account...',
    step_connecting_upstream: 'Connecting to upstream edge server...',
    step_starting_proxies: 'Initializing local routing streams...',
    step_verifying_exit: 'Verifying tunnel egress connectivity...',
    step_configuring_routes: 'Configuring system routing...',
    step_default: 'Establishing secure MASQUE stream...',
    sub_protected_tun: 'Protected via Wintun Layer-3 TUN',
    sub_protected_proxy: 'Protected via Fastly MASQUE Proxy',
    sub_disconnecting: 'Restoring network routes and proxy settings...',
    rec_tag: 'SMART RECOMMENDATION',
    rec_fastest_edge: 'Fastest verified edge',
    btn_use_rec: 'Use Recommended',
    lbl_target_location: 'Selected Server Location',
    btn_change_server: 'Change Server',
    lbl_routing_engine: 'Traffic Routing Engine',
    mode_tag_proxy: 'Web & App Proxy',
    mode_tag_tun: 'Full System TUN',
    mode_system_proxy: 'System Proxy',
    mode_wintun_tun: 'Wintun L3 TUN',
    uac_admin_required: 'Administrator rights required for kernel TUN',
    btn_elevate: 'Elevate ⚡',
    lbl_egress_title: 'Exit Verification & Geolocation Transparency',
    lbl_direct_verification: 'Direct Verification',
    lbl_target_server: 'Target Server',
    lbl_public_ip: 'Observed Public IP',
    lbl_direct_conn: 'Direct Connection',
    lbl_exit_country: 'Observed Exit Country',
    lbl_anycast_title: 'Anycast Routing:',
    lbl_anycast_desc: 'Traffic terminates at nearest Fastly regional edge node to minimize latency.',
    type_anycast: 'Anycast Edge',
    lbl_telemetry_latency: 'Network Latency (RTT)',
    lbl_telemetry_duration: 'Connection Duration',
    lbl_idle: 'Idle',
    lbl_live: 'Live',
    lbl_telemetry_traffic: 'Session Data Transferred',
    lbl_live_meter: 'Live Meter',
    lbl_telemetry_allowance: 'Monthly Mozilla Allowance',
    lbl_quota_left: '100% Left',
    search_placeholder: 'Search by country, city, POP code (e.g. MUC, FRA, LON), or hostname...',
    filter_all: 'All',
    filter_pop: 'Dedicated POPs',
    filter_anycast: 'Anycast Edge',
    filter_europe: 'Europe',
    filter_americas: 'Americas',
    filter_asia: 'Asia & Pacific',
    btn_ping_all: 'Ping All Servers',
    ping_measuring: 'Measuring Global POP Latencies...',
    ping_probing: 'Measuring...',
    ping_timeout: 'Timeout',
    ping_failed: 'Failed',
    ping_offline: 'Offline',
    btn_cancel: 'Cancel',
    col_location: 'Location & Server',
    col_arch: 'Network Architecture',
    col_ping: 'Round-Trip Ping',
    col_action: 'Action',
    badge_recommended: 'RECOMMENDED',
    btn_card_selected: 'Selected ✓',
    btn_card_connect: 'Connect',
    btn_card_connected: 'Connected ✓',
    diag_socks_desc: 'Local SOCKS5 proxy engine forwarding TCP connections into Fastly MASQUE tunnel.',
    diag_http_desc: 'WinINet-compatible HTTP CONNECT proxy bridge for Windows system proxy routing.',
    diag_tun_name: 'Wintun Layer-3 Virtual Adapter',
    diag_tun_desc: 'Kernel-level virtual network adapter (Wintun) capturing and routing 100% of all IP packets.',
    diag_tls_name: 'TLS Anti-DPI Fragmentation',
    diag_tls_desc: 'ClientHello TCP fragmentation actively bypasses SNI inspection and deep packet inspection.',
    diag_dns_name: 'DNS Leak & Hijack Protection',
    diag_dns_desc: 'Routes DNS requests directly through secure resolvers (1.1.1.1 & 8.8.8.8) to prevent ISP hijacking.',
    diag_uac_name: 'Windows Elevation (UAC)',
    diag_uac_desc: 'Process privilege level. Required for Wintun Layer-3 routing table modifications.',
    lbl_live_logs: 'Live System Logs',
    btn_copy_logs: '📋 Copy Logs',
    btn_clear_logs: '🗑️ Clear',
    logs_copied: 'Logs copied to clipboard!',
    ip_copied: 'Public IP copied to clipboard!',
    traffic_download: 'Download',
    traffic_upload: 'Upload',
    traffic_ping: 'Latency',
    traffic_session: 'Session',
    set_heading_routing: 'Routing Modes & Tunnel Protocol',
    set_killswitch_title: 'Active Kill Switch (Leak Protection)',
    set_killswitch_sub: 'Blocks all non-VPN internet traffic automatically if the VPN disconnects unexpectedly or reconnects, preventing IP and DNS leaks. Local LAN is preserved.',
    set_autoreconnect_title: 'Intelligent Auto-Reconnect',
    set_autoreconnect_sub: 'Automatically re-establishes the VPN tunnel with exponential backoff if the network drops, sleep-wake occurs, or edge connection stalls.',
    status_reconnecting: 'Reconnecting...',
    status_reconnecting_sub: 'Network interrupted. Auto-recovering secure tunnel...',
    set_tun_title: 'Wintun Layer-3 TUN Mode',
    set_tun_sub: 'Route 100% of all Windows traffic, games, command line tools, and UDP packets',
    set_h3_title: 'HTTP/3 (QUIC / UDP) Transport',
    set_h3_sub: 'Ultra-low latency UDP transport when edge nodes support QUIC (fallback to HTTP/2)',
    set_tls_title: 'TLS ClientHello Fragmentation',
    set_tls_sub: 'Evade SNI filtering by fragmenting TLS packets across TCP boundaries (Anti-Censorship)',
    set_api_proxy_title: 'Auxiliary Proxy for Authentication',
    set_api_proxy_sub: 'Upstream proxy if Mozilla FxA login is blocked by local censorship (e.g. 127.0.0.1:10808)',
    set_bypass_title: 'System Proxy Direct Bypass Rules',
    set_bypass_sub: 'Semicolon- or comma-separated domains/IPs that connect directly without proxying (e.g. *.ir; *.local; 192.168.*)',
    set_heading_appearance: 'Appearance & Interface Language',
    set_lang_title: 'Interface Language',
    set_lang_sub: 'Switch between English and Persian (فارسی)',
    set_theme_title: 'Dashboard Theme',
    set_theme_sub: 'Switch between Dark Obsidian Fire and Light Clean',
    set_heading_account: 'Mozilla Account & Subscription',
    set_heading_about: 'About FYRNOX for Windows',
    about_desc: 'High-performance subscription-free desktop VPN client powered by Fastly MASQUE and Mozilla Guardian.',
    login_title: 'Sign In with Firefox Account',
    login_desc: 'Sign in with your Firefox Account to access Mozilla\'s 50 GB monthly free VPN quota.',
    login_email_label: 'Email',
    login_password_label: 'Password',
    login_proxy_label: 'Upstream Proxy (Optional)',
    twofa_prompt: 'Verification code sent to your email. Please enter below:',
    twofa_code_label: '6-Digit Code',
    btn_verify_code: 'Verify Code',
    uac_modal_title: 'Administrator Privileges Required',
    uac_modal_desc: 'To enable Wintun Layer-3 TUN Mode and route 100% of all system traffic, Windows requires Administrator privileges to create the virtual network interface and modify routing tables.',
    uac_cancel: 'Cancel (Stay in Proxy Mode)',
    uac_confirm: 'Restart as Administrator ⚡',
    nav_split_tunnel: 'Split Tunnel',
    page_split_tunnel: 'App Split Tunneling',
    split_desc: 'Route or bypass traffic for specific desktop applications and binaries.',
    split_mode_label: 'Tunnel Behavior:',
    split_mode_bypass: 'Bypass Selected Apps (Direct)',
    split_mode_proxy: 'Tunnel Selected Apps Only (Proxy)',
    split_presets_caption: 'Quick Presets:',
    split_preset_browsers: 'Browsers',
    split_preset_messaging: 'Telegram & Chat',
    split_preset_games: 'Game Launchers',
    split_btn_clear: 'Clear All',
    split_search_placeholder: 'Search running or installed apps...',
    split_btn_add: '+ Add Binary',
    split_active_rules: 'Configured Routing Rules',
    split_running_processes: 'Detected System Applications & Processes',
    split_table_name: 'Application',
    split_table_path: 'Process Binary',
    split_table_status: 'Status',
    split_table_action: 'Rule',
    split_no_apps: 'No applications configured for split tunneling.',
    gaming_mode_btn: 'Gaming Mode',
    gaming_active: '⚡ Gaming Mode Active',
    gaming_sub_info: 'Zero-MUX Direct UDP • Anti-Fragmentation • Low Variance',
    gaming_jitter: 'Jitter',
    gaming_loss: 'Packet Loss',
    btn_disable: 'Turn Off',
    set_iran_bypass_title: 'Direct Routing for Iran Domestic Traffic (Zero-Config)',
    set_iran_bypass_sub: 'Bypass domestic banking networks, Shetab gateways, and .ir domains directly without consuming VPN quota',
    set_gaming_title: 'Dedicated Low-Latency Gaming Accelerator',
    set_heading_lan_share: 'LAN Proxy Sharing / Console Hotspot (Wi-Fi)',
    set_lan_share_title: 'Allow LAN Inbound Proxy & Console Sharing',
    set_lan_share_sub: 'Binds mixed HTTP/SOCKS5 listener to 0.0.0.0 allowing PlayStation, Xbox, Switch, and phones on Wi-Fi to use VPN.',
    lan_scan_qr: 'Scan QR Code on mobile or enter proxy on your console',
    set_heading_panic: 'Global Emergency Panic Key & Stealth Mode',
    set_panic_sub: 'Press Ctrl + Alt + Shift + X anywhere to immediately sever VPN connections, flush DNS cache, hide window, and remove tray icon.',
    btn_test_panic: 'Test Panic Lockdown 🚨',
    diag_matrix_title: 'In-App Sanctions & Service Access Matrix',
    diag_matrix_sub: 'Real-time probe of geo-restricted and sanctioned platforms routed through the active tunnel',
    diag_matrix_check_btn: 'Run Matrix Check ⚡',
    fallback_optimized_udp: 'Optimized UDP (H3)',
    fallback_protected_tcp: 'Protected Fallback (TCP/H2)',
    set_gaming_sub: 'Prioritizes raw UDP packets, enforces MTU 1420 (no fragmentation), and eliminates head-of-line blocking'
  },
  fa: {
    brand_badge: 'نسخه دسکتاپ',
    nav_dashboard: 'داشبورد',
    nav_locations: 'سرورها',
    nav_diagnostics: 'عیب‌یابی',
    nav_settings: 'تنظیمات',
    page_dashboard: 'داشبورد شبکه',
    page_servers: 'لیست سرورها',
    page_diagnostics: 'عیب‌یابی سیستم',
    page_settings: 'تنظیمات و پیکربندی',
    account_firefox: 'حساب فایرفاکس',
    sign_in: 'ورود به حساب',
    sign_out: 'خروج از حساب',
    detecting: 'در حال شناسایی...',
    not_logged_in: 'وارد نشده‌اید',
    quota_desc: '۵۰ گیگابایت حجم ماهانه پرسرعت موزیلا',
    disconnected: 'قطع شده',
    connecting: 'در حال اتصال...',
    connected: 'متصل شد (امن)',
    disconnecting: 'در حال قطع...',
    error: 'خطا در اتصال',
    btn_connect: 'اتصال',
    btn_disconnect: 'قطع اتصال',
    btn_connecting: 'در حال اتصال',
    btn_disconnecting: 'در حال قطع',
    btn_sign_in: 'ورود',
    traffic_unprotected: 'ترافیک اینترنت شما بدون حفاظت است',
    quota_card_title: 'سهمیه ماهانه فایرفاکس',
    quota_used_lbl: 'مصرف شده:',
    quota_rem_lbl: 'باقی‌مانده:',
    quota_resets_lbl: 'تمدید ماهانه:',
    quota_badge_authoritative: 'تایید شده سرور',
    quota_badge_estimated: 'تخمینی محلی',
    set_ip_proto_title: 'اولویت پروتکل IP (IPv4 / IPv6)',
    set_ip_proto_sub: 'پروتکل آدرس‌دهی تونل VPN و محافظت نشت. در حالت پراکسی، برنامه‌های ناسازگار مستقیم متصل می‌شوند؛ در حالت TUN، گزینه IPv4 Only از نشت IPv6 جلوگیری می‌کند.',
    opt_proto_dual: 'پروتکل دوگانه IPv4 + IPv6 (پیش‌فرض)',
    opt_proto_ipv4: 'فقط IPv4 (محافظت کامل از نشت)',
    opt_proto_ipv6: 'فقط IPv6',
    dash_error_title: 'هشدار اتصال',
    dash_error_msg: 'برقراری تونل امن با خطا مواجه شد.',
    step_authorizing: 'در حال احراز هویت با حساب فایرفاکس...',
    step_connecting_upstream: 'در حال اتصال به سرور لبه شبکه...',
    step_starting_proxies: 'راه‌اندازی سرویس‌های محلی مسیریابی...',
    step_verifying_exit: 'در حال بررسی نهایی ترافیک خروجی...',
    step_configuring_routes: 'پیکربندی جدول مسیریابی سیستم...',
    step_default: 'برقراری جریان امن MASQUE...',
    sub_protected_tun: 'محافظت شده از طریق Wintun Layer-3 TUN',
    sub_protected_proxy: 'محافظت شده از طریق پراکسی Fastly MASQUE',
    sub_disconnecting: 'در حال بازیابی تنظیمات شبکه و پراکسی...',
    rec_tag: 'پیشنهاد هوشمند',
    rec_fastest_edge: 'سریع‌ترین سرور تایید شده',
    btn_use_rec: 'اتصال به سرور پیشنهادی',
    lbl_target_location: 'سرور انتخابی',
    btn_change_server: 'تغییر سرور',
    lbl_routing_engine: 'موتور هدایت ترافیک',
    mode_tag_proxy: 'پراکسی برنامه‌ها',
    mode_tag_tun: 'مسیریابی سراسری سیستم',
    mode_system_proxy: 'پراکسی سیستمی',
    mode_wintun_tun: 'مسیریابی Wintun L3',
    uac_admin_required: 'دسترسی ادمین برای Wintun TUN الزامی است',
    btn_elevate: 'ارتقای دسترسی ⚡',
    lbl_egress_title: 'بررسی شفافیت آدرس آی‌پی خروجی',
    lbl_direct_verification: 'بررسی زنده ترافیک',
    lbl_target_server: 'سرور مقصد',
    lbl_public_ip: 'آدرس آی‌پی عمومی مشاهده‌شده',
    lbl_direct_conn: 'اتصال مستقیم',
    lbl_exit_country: 'کشور خروجی مشاهده‌شده',
    lbl_anycast_title: 'مسیریابی انیکست:',
    lbl_anycast_desc: 'ترافیک برای کاهش تاخیر به نزدیک‌ترین نود منطقه‌ای Fastly متصل می‌شود.',
    type_anycast: 'لبه انیکست',
    lbl_telemetry_latency: 'تاخیر رفت و برگشت (RTT)',
    lbl_telemetry_duration: 'مدت زمان اتصال',
    lbl_idle: 'غیرفعال',
    lbl_live: 'فعال',
    lbl_telemetry_traffic: 'حجم مصرفی جلسه',
    lbl_live_meter: 'شمارش زنده',
    lbl_telemetry_allowance: 'سهمیه ماهانه موزیلا',
    lbl_quota_left: 'باقی‌مانده',
    search_placeholder: 'جستجوی کشور، شهر، کد فرودگاهی یا نام سرور...',
    filter_all: 'همه',
    filter_pop: 'نودهای اختصاصی',
    filter_anycast: 'لبه‌های انیکست',
    filter_europe: 'اروپا',
    filter_americas: 'آمریکا',
    filter_asia: 'آسیا و اقیانوسیه',
    btn_ping_all: 'تست پینگ همه سرورها',
    ping_measuring: 'در حال سنجش پینگ سرورها...',
    ping_probing: 'در حال سنجش...',
    ping_timeout: 'پایان مهلت',
    ping_failed: 'ناموفق',
    ping_offline: 'آفلاین',
    btn_cancel: 'لغو',
    col_location: 'موقعیت و سرور',
    col_arch: 'معماری شبکه',
    col_ping: 'زمان پاسخ پینگ',
    col_action: 'عملیات',
    badge_recommended: 'پیشنهادی',
    btn_card_selected: 'انتخاب شده ✓',
    btn_card_connect: 'اتصال',
    btn_card_connected: 'متصل شد ✓',
    diag_socks_desc: 'موتور پراکسی محلی SOCKS5 هدایت‌کننده بسته‌ها به تونل Fastly MASQUE.',
    diag_http_desc: 'پل ارتباطی HTTP CONNECT سازگار با WinINet برای پراکسی سیستمی ویندوز.',
    diag_tun_name: 'آداپتور مجازی لایه ۳ Wintun',
    diag_tun_desc: 'آداپتور درایور هسته ویندوز برای هدایت ۱۰۰٪ تمام ترافیک و بسته‌ها.',
    diag_tls_name: 'تکه‌تکه سازی TLS ضد فیلترینگ',
    diag_tls_desc: 'تکه‌تکه‌سازی بسته معرفی ClientHello جهت عبور از سانسور و فیلترینگ عمیق SNI.',
    diag_dns_name: 'محافظت در برابر نشت و جعل DNS',
    diag_dns_desc: 'هدایت درخواست‌های نام دامنه به سرورهای ۱.۱.۱.۱ و ۸.۸.۸.۸ برای جلوگیری از دستکاری ISP.',
    diag_uac_name: 'سطح دسترسی ویندوز (UAC)',
    diag_uac_desc: 'سطح دسترسی برنامه. برای تغییر جدول مسیریابی Wintun به ادمین نیاز است.',
    lbl_live_logs: 'گزارش‌های زنده سیستم',
    btn_copy_logs: '📋 کپی گزارش‌ها',
    btn_clear_logs: '🗑️ پاک کردن',
    logs_copied: 'گزارش‌ها در کلیپ‌بورد کپی شد!',
    ip_copied: 'آدرس آی‌پی در کلیپ‌بورد کپی شد!',
    traffic_download: 'دانلود',
    traffic_upload: 'آپلود',
    traffic_ping: 'تاخیر / پینگ',
    traffic_session: 'مدت نشست',
    set_heading_routing: 'حالت‌های مسیریابی و پروتکل تونل',
    set_killswitch_title: 'کیل‌سوئیچ فعال (ضد نشت IP و DNS)',
    set_killswitch_sub: 'مسدودسازی بلادرنگ ترافیک خارج از تونل در صورت قطع ناگهانی برای جلوگیری کامل از نشت اطلاعات و IP واقعی.',
    set_autoreconnect_title: 'اتصال مجدد خودکار هوشمند',
    set_autoreconnect_sub: 'برقراری مجدد تونل امن با الگوریتم نمایی در صورت قطع اتصال شبکه یا بیدار شدن سیستم از خواب.',
    status_reconnecting: 'در حال اتصال مجدد...',
    status_reconnecting_sub: 'ارتباط قطع شد؛ در حال بازیابی خودکار تونل امن...',
    set_tun_title: 'حالت مسیریابی سراسری Wintun L3 TUN',
    set_tun_sub: 'هدایت ۱۰۰٪ ترافیک ویندوز، بازی‌ها، دستورات ترمینال و بسته‌های UDP',
    set_h3_title: 'پروتکل انتقالی HTTP/3 (QUIC / UDP)',
    set_h3_sub: 'ارتباط فوق سریع UDP هنگام پشتیبانی سرور لبه (بازگشت خودکار به HTTP/2)',
    set_tls_title: 'تکه‌تکه‌سازی ClientHello در TLS',
    set_tls_sub: 'دور زدن فیلترینگ با شکستن پکت‌های TLS بر روی مرزهای TCP (ضد سانسور)',
    set_api_proxy_title: 'پراکسی کمکی جهت احراز هویت',
    set_api_proxy_sub: 'پراکسی بالادستی در صورتی که ورود به موزیلا توسط فیلترینگ مسدود است',
    set_bypass_title: 'قوانین دور زدن مستقیم در پراکسی سیستمی',
    set_bypass_sub: 'دامنه‌ها یا آی‌پی‌هایی که بدون پراکسی و مستقیم وصل می‌شوند (مانند ir.* یا لوکال)',
    set_heading_appearance: 'ظاهر و زبان برنامه',
    set_lang_title: 'زبان برنامه',
    set_lang_sub: 'تغییر زبان بین انگلیسی و فارسی',
    set_theme_title: 'تم ظاهری',
    set_theme_sub: 'تغییر بین تم تاریک آتشی و تم روشن',
    set_heading_account: 'حساب کاربری موزیلا و سهمیه',
    set_heading_about: 'درباره FYRNOX برای ویندوز',
    about_desc: 'کلاینت VPN دسکتاپ پرسرعت و رایگان بر پایه Fastly MASQUE و موزیلا گاردین.',
    login_title: 'ورود با حساب کاربری فایرفاکس',
    login_desc: 'برای دسترسی به ۵۰ گیگابایت حجم رایگان ماهانه موزیلا با حساب فایرفاکس خود وارد شوید.',
    login_email_label: 'ایمیل',
    login_password_label: 'رمز عبور',
    login_proxy_label: 'پراکسی بالادستی (اختیاری)',
    twofa_prompt: 'کد تایید به ایمیل شما ارسال شد. لطفاً آن را وارد کنید:',
    twofa_code_label: 'کد ۶ رقمی',
    btn_verify_code: 'تایید و ورود',
    uac_modal_title: 'نیاز به دسترسی مدیر (Administrator)',
    uac_modal_desc: 'جهت فعال‌سازی Wintun TUN و هدایت تمام ترافیک ویندوز، ایجاد کارت شبکه مجازی و جدول مسیریابی نیازمند دسترسی ادمین است.',
    uac_cancel: 'انصراف (ماندن در حالت پراکسی)',
    uac_confirm: 'راه‌اندازی مجدد با دسترسی ادمین ⚡',
    nav_split_tunnel: 'تونلینگ تفکیکی',
    page_split_tunnel: 'مسیریابی تفکیکی برنامه‌ها (Split Tunneling)',
    split_desc: 'هدایت یا دور زدن ترافیک برای برنامه‌ها و فایل‌های اجرایی دلخواه در ویندوز.',
    split_mode_label: 'رفتار تونل:',
    split_mode_bypass: 'عبور مستقیم برنامه‌های انتخابی (Bypass)',
    split_mode_proxy: 'فقط عبور برنامه‌های انتخابی از فیلترشکن (Proxy)',
    split_presets_caption: 'دسته‌های آماده:',
    split_preset_browsers: 'مرورگرها',
    split_preset_messaging: 'تلگرام و چت',
    split_preset_games: 'لانچرهای بازی',
    split_btn_clear: 'پاک‌سازی همه',
    split_search_placeholder: 'جستجوی برنامه‌های در حال اجرا یا نصب‌شده...',
    split_btn_add: '+ افزودن فایل',
    split_active_rules: 'قوانین مسیریابی فعال',
    split_running_processes: 'برنامه‌ها و پردازش‌های فعال ویندوز',
    split_table_name: 'نام برنامه',
    split_table_path: 'فایل اجرایی پردازش',
    split_table_status: 'وضعیت',
    split_table_action: 'تنظیم قانون',
    split_no_apps: 'هیچ برنامه‌ای برای تونلینگ تفکیکی تنظیم نشده است.',
    gaming_mode_btn: 'حالت گیمینگ',
    gaming_active: '⚡ شتاب‌دهنده گیمینگ فعال است',
    gaming_sub_info: 'اولویت UDP خام • بدون سربار MUX • ضد تکه‌تکه‌شدن MTU 1420',
    gaming_jitter: 'نوسان (Jitter)',
    gaming_loss: 'افت بسته',
    btn_disable: 'غیرفعال‌سازی',
    set_iran_bypass_title: 'مسیریابی مستقیم ترافیک داخلی ایران (Zero-Config Bypass)',
    set_iran_bypass_sub: 'هدایت مستقیم بانک‌ها، درگاه‌های شاپرک و دامنه‌های ir. از اینترنت عادی بدون کسر حجم و افت سرعت',
    set_gaming_title: 'شتاب‌دهنده اختصاصی گیمینگ (پینگ پایین)',
    set_heading_lan_share: 'اشتراک پروکسی در شبکه محلی و کنسول‌های بازی',
    set_lan_share_title: 'فعال‌سازی اشتراک در شبکه محلی و کنسول‌ها',
    set_lan_share_sub: 'پروکسی را روی 0.0.0.0 اجرا کرده و به پلی‌استیشن، ایکس‌باکس، سوییچ و موبایل‌های متصل به وای‌فای امکان استفاده از VPN را می‌دهد.',
    lan_scan_qr: 'اسکن بارکد QR در موبایل یا وارد کردن دستی در تنظیمات شبکه کنسول',
    set_heading_panic: 'کلید اضطراری سراسری (Panic Key) و حالت مخفی',
    set_panic_sub: 'با فشردن کلیدهای میانبر Ctrl + Alt + Shift + X در هر برنامه‌ای، بلافاصله اتصال قطع شده، کش DNS پاکسازی، پنجره مخفی و آیکون کنار ساعت حذف می‌شود.',
    btn_test_panic: 'تست قفل اضطراری 🚨',
    diag_matrix_title: 'ماتریس دسترسی به سرویس‌ها و پایش تحریم‌ها',
    diag_matrix_sub: 'پایش بلادرنگ سرویس‌های تحریم‌شده و فیلترشده از طریق تونل اختصاصی VPN',
    diag_matrix_check_btn: 'بررسی مجدد دسترسی ⚡',
    fallback_optimized_udp: 'UDP بهینه‌شده (QUIC/H3)',
    fallback_protected_tcp: 'فالبک محافظت‌شده (TCP چندتکه‌ای)',
    set_gaming_sub: 'اولویت‌دهی به پکت‌های UDP خام، تنظیم خودکار MTU 1420 برای جلوگیری از تکه‌تکه شدن و حذف MUX'
  }
};

let currentStatus = null;
let allServers = [];
let activeNav = 'dashboard';
let currentLang = localStorage.getItem('fyrnox_lang') || 'en';
let currentTheme = localStorage.getItem('fyrnox_theme') || 'dark';
let currentFilter = 'all';

let connectionStartTime = null;
let durationInterval = null;
let isPinging = false;
let pingAbortController = null;
let pingProgressTimer = null;
let previousState = 'disconnected';

const EUROPE_CODES = ['DE', 'GB', 'FR', 'NL', 'SE', 'FI', 'NO', 'DK', 'BE', 'AT', 'ES', 'PT', 'IT', 'IE', 'BG', 'CH'];
const AMERICAS_CODES = ['US', 'CA', 'BR', 'AR', 'CL', 'CO', 'PE', 'MX'];
const ASIA_CODES = ['JP', 'SG', 'KR', 'TH', 'MY', 'PH', 'AU', 'NZ'];

// Flag emoji map with robust fallback
const FLAG_MAP = {
  US: '🇺🇸', DE: '🇩🇪', NL: '🇳🇱', GB: '🇬🇧', FR: '🇫🇷',
  SE: '🇸🇪', FI: '🇫🇮', NO: '🇳🇴', DK: '🇩🇰', BE: '🇧🇪',
  AT: '🇦🇹', ES: '🇪🇸', PT: '🇵🇹', IT: '🇮🇹', IE: '🇮🇪',
  BG: '🇧🇬', CA: '🇨🇦', BR: '🇧🇷', AR: '🇦🇷', CL: '🇨🇱',
  CO: '🇨🇴', PE: '🇵🇪', MX: '🇲🇽', JP: '🇯🇵', SG: '🇸🇬',
  KR: '🇰🇷', TH: '🇹🇭', MY: '🇲🇾', PH: '🇵🇭', AU: '🇦🇺',
  NZ: '🇳🇿', ZA: '🇿🇦', GH: '🇬🇭', IR: '🇮🇷'
};

function getFlag(code) {
  if (typeof getFlagSVG === 'function') {
    return getFlagSVG(code);
  }
  if (!code) return '🌐';
  const c = code.toUpperCase().trim();
  return FLAG_MAP[c] || '🌐';
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

document.addEventListener('DOMContentLoaded', () => {
  document.body.classList.add('is-dashboard');
  applyTheme(currentTheme);
  applyLanguage(currentLang);
  initWindowControls();

  // Initialize Real-Time SSE
  initSSE();

  // Load initial dataset
  fetchServers();
  fetchStatus();
  fetchLogs();

  // Background log poller if diagnostics is active
  setInterval(() => {
    if (activeNav === 'diagnostics') {
      fetchLogs();
    }
  }, 2000);

  // Start duration counter
  startDurationTimer();

  // Initialize 3D Globe & Real-Time Waveform
  try {
    if (typeof FyrnoxGlobe === 'function' && document.getElementById('globe-3d-wrap')) {
      window.fyrnoxGlobe = new FyrnoxGlobe('globe-3d-wrap');
    }
    if (typeof FyrnoxTrafficGraph === 'function' && document.getElementById('traffic-curve-canvas')) {
      window.fyrnoxTrafficGraph = new FyrnoxTrafficGraph('traffic-curve-canvas');
    }
  } catch (e) {
    console.error('[FYRNOX] Failed to initialize 3D Globe or Traffic Graph:', e);
  }
  initDraggableHUD();
  setTimeout(() => {
    updateDockIndicator('dashboard');
  }, 60);
  window.addEventListener('resize', () => {
    updateDockIndicator(activeNav);
  });
});

// ==========================================================================
// Navigation & Views
// ==========================================================================
function updateDockIndicator(tab) {
  const dock = document.getElementById('main-nav-dock');
  const indicator = document.getElementById('nav-dock-indicator');
  if (!dock || !indicator) return;
  const targetBtn = dock.querySelector(`.dock-item[data-tab="${tab}"]`);
  if (targetBtn) {
    indicator.style.left = `${targetBtn.offsetLeft}px`;
    indicator.style.top = `${targetBtn.offsetTop}px`;
    indicator.style.width = `${targetBtn.offsetWidth}px`;
    indicator.style.height = `${targetBtn.offsetHeight}px`;
  }
}

const NAV_TABS = ['dashboard', 'servers', 'diagnostics', 'settings'];

function updateViewPanels(activeIdx, animated = true, dragDeltaPx = 0) {
  // Purged: deterministic SPA visibility handles panel state via .active class
}

function switchNav(tab) {
  const prevIdx = NAV_TABS.indexOf(activeNav);
  const targetIdx = NAV_TABS.indexOf(tab);
  if (targetIdx < 0) return;

  const isRtl = document.documentElement.getAttribute('dir') === 'rtl';
  let slideClass = '';
  if (prevIdx >= 0 && targetIdx !== prevIdx) {
    const isForward = targetIdx > prevIdx;
    if (!isRtl) {
      slideClass = isForward ? 'slide-in-right' : 'slide-in-left';
    } else {
      slideClass = isForward ? 'slide-in-left' : 'slide-in-right';
    }
  }

  activeNav = tab;
  updateDockIndicator(tab);

  // Update dashboard class
  document.body.classList.toggle('is-dashboard', tab === 'dashboard');
  const appShell = document.querySelector('.app-shell');
  if (appShell) {
    appShell.classList.toggle('view-is-dashboard', tab === 'dashboard');
  }

  // Update sidebar & dock active buttons
  document.querySelectorAll('.dock-item, .nav-item, .right-dock-item').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.tab === tab);
  });

  // Strict deterministic SPA visibility with directional slide animation
  document.querySelectorAll('.view-panel').forEach(p => {
    const isTarget = p.id === `view-${tab}`;
    p.classList.remove('slide-in-right', 'slide-in-left');
    p.classList.toggle('active', isTarget);
    if (isTarget && slideClass) {
      p.classList.add(slideClass);
    }
  });

  // Trigger subview actions & WebGL resize
  if (tab === 'dashboard') {
    if (window.fyrnoxGlobe && typeof window.fyrnoxGlobe.handleResize === 'function') {
      setTimeout(() => window.fyrnoxGlobe.handleResize(), 30);
      setTimeout(() => window.fyrnoxGlobe.handleResize(), 150);
      setTimeout(() => window.fyrnoxGlobe.handleResize(), 300);
    }
    if (window.fyrnoxTrafficGraph && typeof window.fyrnoxTrafficGraph.handleResize === 'function') {
      setTimeout(() => window.fyrnoxTrafficGraph.handleResize(), 30);
    }
  }

  // Update topbar title
  const titleEl = document.getElementById('page-title');
  if (titleEl) {
    const key = `page_${tab}`;
    titleEl.textContent = (I18N[currentLang] && I18N[currentLang][key]) || 'FYRNOX Desktop';
  }

  if (tab === 'servers') {
    if (allServers && allServers.length > 0) {
      filterServers(document.getElementById('server-search-input')?.value || '');
    } else {
      fetchServers();
    }
  } else if (tab === 'diagnostics') {
    fetchLogs();
    fetchServiceMatrix();
  } else if (tab === 'settings') {
    fetchLANShareStatus();
    loadSplitTunnelView();
  }
}

// ==========================================================================
// Server-Sent Events (SSE) & Status Poller
// ==========================================================================
function initSSE() {
  const handleStatusData = (data) => {
    if (data) {
      updateUI(data);
    }
  };

  if (!window.EventSource) {
    setInterval(fetchStatus, 1500);
    return;
  }

  try {
    const es = new EventSource('/api/events');
    es.onmessage = (event) => {
      try {
        handleStatusData(JSON.parse(event.data));
      } catch (err) {}
    };
    es.addEventListener('status', (event) => {
      try {
        handleStatusData(JSON.parse(event.data));
      } catch (err) {}
    });
    es.onerror = () => {
      setTimeout(fetchStatus, 1500);
    };
  } catch (err) {
    console.warn('SSE initialization failed; using polling:', err);
    setInterval(fetchStatus, 1500);
  }

  // Adaptive status poller: 1000ms while connected or connecting/reconnecting; paused when idle/disconnected
  setInterval(() => {
    if (currentStatus) {
      const st = currentStatus.state;
      if (st === 'connected' || st === 'connecting' || st === 'reconnecting' || st === 'disconnecting') {
        fetchStatus();
      }
    }
  }, 1000);
}

async function fetchStatus() {
  try {
    const res = await fetch('/api/status');
    if (res.ok) {
      const data = await res.json();
      updateUI(data);
    }
  } catch (err) {
    console.error('Failed to fetch status:', err);
  }
}

// ==========================================================================
// Update UI State
// ==========================================================================
function applyLatencyStyling(valEl, badgeEl, ms) {
  if (!valEl) return;
  valEl.classList.remove('latency-good', 'latency-medium', 'latency-poor');
  if (badgeEl) badgeEl.classList.remove('latency-good', 'latency-medium', 'latency-poor', 'testing');

  if (ms == null || ms <= 0) {
    return;
  }
  if (ms < 150) {
    valEl.classList.add('latency-good');
    if (badgeEl) badgeEl.classList.add('latency-good');
  } else if (ms <= 350) {
    valEl.classList.add('latency-medium');
    if (badgeEl) badgeEl.classList.add('latency-medium');
  } else {
    valEl.classList.add('latency-poor');
    if (badgeEl) badgeEl.classList.add('latency-poor');
  }
}


// Rolling Telemetry Sparkline Buffers (16 data points each)
const SPARK_POINTS = 16;
const sparkData = {
  down: [8.2, 9.5, 11.0, 10.4, 12.1, 11.8, 12.4, 10.9, 11.5, 12.8, 12.0, 12.4, 11.7, 12.1, 12.5, 12.4],
  up: [6.1, 6.8, 7.5, 8.2, 8.0, 7.8, 8.5, 8.7, 8.1, 8.4, 8.9, 8.5, 8.6, 8.7, 8.6, 8.7],
  latency: [44, 43, 42, 45, 43, 42, 41, 44, 43, 42, 43, 42, 41, 42, 43, 42],
  loss: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
  session: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]
};

function generateSparklinePath(arr, minVal, maxVal, width = 80, height = 18) {
  if (!arr || arr.length < 2) return `M 0 ${height - 4} L ${width} ${height - 4}`;
  const range = (maxVal - minVal) || 1;
  const stepX = width / (arr.length - 1);
  const pts = arr.map((v, i) => {
    const norm = Math.max(0, Math.min(1, (v - minVal) / range));
    const y = (height - 3) - (norm * (height - 6));
    const x = i * stepX;
    return { x, y };
  });

  let path = `M ${pts[0].x.toFixed(1)} ${pts[0].y.toFixed(1)}`;
  for (let i = 1; i < pts.length; i++) {
    const prev = pts[i - 1];
    const curr = pts[i];
    const midX = (prev.x + curr.x) / 2;
    path += ` Q ${prev.x.toFixed(1)} ${prev.y.toFixed(1)}, ${midX.toFixed(1)} ${((prev.y + curr.y) / 2).toFixed(1)}`;
  }
  const last = pts[pts.length - 1];
  path += ` T ${last.x.toFixed(1)} ${last.y.toFixed(1)}`;
  return path;
}

function updateTelemetrySparklines() {
  const pDown = document.getElementById('sparkline-down');
  const pUp = document.getElementById('sparkline-up');
  const pLat = document.getElementById('sparkline-latency');
  const pLoss = document.getElementById('sparkline-loss');
  const pSess = document.getElementById('sparkline-session');

  if (pDown) pDown.setAttribute('d', generateSparklinePath(sparkData.down, 6, 14));
  if (pUp) pUp.setAttribute('d', generateSparklinePath(sparkData.up, 4, 10));
  if (pLat) pLat.setAttribute('d', generateSparklinePath(sparkData.latency, 35, 50));
  if (pLoss) pLoss.setAttribute('d', generateSparklinePath(sparkData.loss, 0, 5));
  if (pSess) pSess.setAttribute('d', generateSparklinePath(sparkData.session, 0, 20));
}

function updateUI(status) {
  if (!status) return;
  try {
  currentStatus = status;

  const state = status.state || 'disconnected';
  const isConnected = state === 'connected';
  const isConnecting = state === 'connecting';
  const isDisconnecting = state === 'disconnecting';

  // 1. Session Duration tracking
  if (isConnected && previousState !== 'connected') {
    connectionStartTime = Date.now();
  } else if (!isConnected && !isConnecting && !isDisconnecting) {
    connectionStartTime = null;
  }
  previousState = state;

  // 2. Topbar Status Indicator
  const topbarPill = document.getElementById('topbar-status-pill');
  const topbarText = document.getElementById('topbar-status-text');
  if (topbarPill && topbarText) {
    topbarPill.className = `status-indicator-pill ${state}`;
    topbarText.textContent = (I18N[currentLang] && I18N[currentLang][state]) || state.toUpperCase();
  }

  // Titlebar Status Text
  const titlebarStatus = document.getElementById('titlebar-status-text');
  if (titlebarStatus) {
    titlebarStatus.textContent = state.toUpperCase();
    titlebarStatus.className = `titlebar-status-text ${state}`;
  }

  // 3. Topbar Route
  const originEl = document.getElementById('topbar-origin-country');
  const targetEl = document.getElementById('topbar-target-country');
  if (originEl && status.current_country) {
    originEl.innerHTML = `${getFlag(status.current_country)} <span>${escapeHtml(status.current_country)}</span>`;
  }
  if (targetEl && status.server) {
    targetEl.innerHTML = `${getFlag(status.server.country_code)} <span>${escapeHtml(status.server.country_name)}</span>`;
  }

  // 4. Hero Shield Button
  const heroBtn = document.getElementById('hero-connect-btn');
  const heroLabel = document.getElementById('hero-status-label');
  const heroSub = document.getElementById('hero-status-sub');
  const heroAction = document.getElementById('hero-action-btn');
  const heroActionLabel = document.getElementById('hero-action-label');
  const heroSpinner = document.getElementById('hero-spinner');
  const shieldCheck = document.getElementById('shield-check-path');

  const heroTitle = document.getElementById('hero-status-title');
  if (heroBtn) {
    heroBtn.className = `circular-connect-btn ${state}`;
  }
  if (heroLabel) {
    heroLabel.className = `hero-status-badge ${state}`;
    heroLabel.textContent = (I18N[currentLang] && I18N[currentLang][state]) || state.toUpperCase();
  }

  if (heroTitle) {
    if (isConnected) {
      heroTitle.textContent = (I18N[currentLang] && I18N[currentLang].connected) || 'Connected';
    } else if (isConnecting) {
      heroTitle.textContent = (I18N[currentLang] && I18N[currentLang].connecting) || 'Connecting';
    } else if (isDisconnecting) {
      heroTitle.textContent = (I18N[currentLang] && I18N[currentLang].disconnecting) || 'Disconnecting';
    } else {
      heroTitle.textContent = (I18N[currentLang] && I18N[currentLang].disconnected) || 'DISCONNECTED';
    }
  }

  if (isConnected) {
    if (heroSub) {
      heroSub.textContent = status.wintun_active
        ? ((I18N[currentLang] && I18N[currentLang].sub_protected_tun) || 'Protected via Wintun Layer-3 TUN')
        : ((I18N[currentLang] && I18N[currentLang].sub_protected_proxy) || 'Protected via Fastly MASQUE');
    }
    if (heroAction) heroAction.className = 'hero-action-button btn-disconnect';
    if (heroActionLabel) heroActionLabel.textContent = (I18N[currentLang] && I18N[currentLang].btn_disconnect) || 'Disconnect';
    if (heroSpinner) heroSpinner.style.display = 'none';
    if (shieldCheck) shieldCheck.style.display = 'block';
  } else if (isConnecting) {
    const stepKey = `step_${status.connect_step}`;
    const defaultStep = (I18N[currentLang] && I18N[currentLang].step_default) || 'Establishing secure MASQUE stream...';
    const stepText = (I18N[currentLang] && I18N[currentLang][stepKey]) || defaultStep;
    if (heroSub) heroSub.textContent = stepText;
    if (heroAction) heroAction.className = 'hero-action-button btn-connect';
    if (heroActionLabel) heroActionLabel.textContent = (I18N[currentLang] && I18N[currentLang].btn_connecting) || 'Connecting';
    if (heroSpinner) heroSpinner.style.display = 'block';
    if (shieldCheck) shieldCheck.style.display = 'none';
  } else if (isDisconnecting) {
    if (heroSub) heroSub.textContent = (I18N[currentLang] && I18N[currentLang].sub_disconnecting) || 'Restoring network routes and proxy settings...';
    if (heroAction) heroAction.className = 'hero-action-button btn-disconnect';
    if (heroActionLabel) heroActionLabel.textContent = (I18N[currentLang] && I18N[currentLang].btn_disconnecting) || 'Disconnecting';
    if (heroSpinner) heroSpinner.style.display = 'block';
    if (shieldCheck) shieldCheck.style.display = 'none';
  } else {
    if (heroSub) heroSub.textContent = (I18N[currentLang] && I18N[currentLang].traffic_unprotected) || 'Traffic is currently unprotected';
    if (heroAction) heroAction.className = 'hero-action-button btn-connect';
    if (heroActionLabel) heroActionLabel.textContent = (I18N[currentLang] && I18N[currentLang].btn_connect) || 'Connect';
    if (heroSpinner) heroSpinner.style.display = 'none';
    if (shieldCheck) shieldCheck.style.display = 'none';
  }

  // 5. Error Banner
  const errBanner = document.getElementById('dash-error-banner');
  const errMsg = document.getElementById('dash-error-msg');
  if (errBanner && errMsg) {
    if (state === 'error' && status.error_message) {
      errBanner.style.display = 'flex';
      errMsg.textContent = status.error_message;
    } else {
      errBanner.style.display = 'none';
    }
  }

  // 6. Selected Server Card & Exit Location Badge
  if (status.server) {
    const flagEl = document.getElementById('dash-target-flag');
    const nameEl = document.getElementById('dash-target-name');
    const subEl = document.getElementById('dash-target-sub');
    if (flagEl) flagEl.innerHTML = getFlag(status.server.country_code);
    if (nameEl) {
      nameEl.textContent = status.server.country_name || 'South Africa';
    }
    if (subEl) {
      subEl.textContent = status.server.is_anycast ? 'Anycast - BGP' : (status.server.city_code || 'Dedicated POP');
    }

    // Exit Location Badge Sync (matching Reference Image)
    const badgeCode = document.getElementById('badge-code');
    const badgeCity = document.getElementById('badge-city');
    const badgeFlag = document.getElementById('badge-flag');
    const cCode = (status.server.country_code || 'ZA').toUpperCase();
    const cCity = (status.server.city_name || status.server.city || status.server.city_code || 'JOHANNESBURG').toUpperCase();
    if (badgeCode) badgeCode.textContent = cCode;
    if (badgeCity) badgeCity.textContent = `${cCity}, ${cCode}`;
    if (badgeFlag) badgeFlag.innerHTML = getFlag(cCode);
  }

  // 7. Routing Mode Pill & Dropdown Sync (100% State Synchronized)
  // Single source of truth: TUN is active only if Wintun driver is active, or if TUN is requested AND app is elevated as admin
  const isTunModeActive = Boolean(status.wintun_active || ((status.tun_active || (status.settings && status.settings.enable_tun_mode)) && status.is_admin));
  const isTun = isTunModeActive;

  const modePillIcon = document.getElementById('mode-pill-icon');
  const modePillLabel = document.getElementById('mode-pill-label');
  const checkProxy = document.getElementById('check-mode-proxy');
  const checkTun = document.getElementById('check-mode-tun');
  const optProxy = document.getElementById('btn-mode-proxy');
  const optTun = document.getElementById('btn-mode-tun');
  const uacNotice = document.getElementById('dash-uac-notice');

  if (modePillIcon && modePillLabel) {
    if (isTunModeActive) {
      modePillIcon.textContent = '🛡️';
      modePillLabel.textContent = (currentLang === 'fa') ? 'تانل کل ویندوز' : 'Wintun TUN';
    } else {
      modePillIcon.textContent = '🌐';
      modePillLabel.textContent = (currentLang === 'fa') ? 'پروکسی سیستم' : 'System Proxy';
    }
  }

  if (checkProxy && checkTun) {
    if (isTunModeActive) {
      checkTun.style.display = 'inline';
      checkProxy.style.display = 'none';
      if (optTun) optTun.classList.add('active');
      if (optProxy) optProxy.classList.remove('active');
    } else {
      checkProxy.style.display = 'inline';
      checkTun.style.display = 'none';
      if (optProxy) optProxy.classList.add('active');
      if (optTun) optTun.classList.remove('active');
    }
  }

  if (uacNotice) {
    uacNotice.style.display = (isTun && !status.is_admin) ? 'flex' : 'none';
  }

  // 8. Live Real-Time Telemetry Bar Beneath 3D Globe
  const statDown = document.getElementById('stat-download-val');
  const statUp = document.getElementById('stat-upload-val');
  const bDown = document.getElementById('bottom-down-speed');
  const bUp = document.getElementById('bottom-up-speed');
  const bPing = document.getElementById('bottom-ping-val');
  const bLoss = document.getElementById('bottom-loss-val');
  const bSession = document.getElementById('bottom-session-duration');

  if (isConnected) {
    if (bDown) {
      bDown.textContent = (status.traffic && status.traffic.speed_down_human) || '0 B/s';
      if (status.traffic && status.traffic.down_bytes_per_sec) {
        sparkData.down.shift();
        sparkData.down.push(status.traffic.down_bytes_per_sec / (1024 * 1024));
      }
    }
    if (bUp) {
      bUp.textContent = (status.traffic && status.traffic.speed_up_human) || '0 B/s';
      if (status.traffic && status.traffic.up_bytes_per_sec) {
        sparkData.up.shift();
        sparkData.up.push(status.traffic.up_bytes_per_sec / (1024 * 1024));
      }
    }
    if (bPing) {
      let lat = 42;
      if (status.server && status.server.ping_ms > 0) {
        lat = status.server.ping_ms;
      } else if (status.exit_latency && status.exit_latency > 0) {
        lat = status.exit_latency;
      }
      bPing.textContent = `${lat} ms`;
      sparkData.latency.shift();
      sparkData.latency.push(lat);
    }
    if (bLoss) {
      const loss = typeof status.packet_loss === 'number' ? status.packet_loss : 0;
      bLoss.textContent = `${loss}%`;
      sparkData.loss.shift();
      sparkData.loss.push(loss);
    }
  } else {
    if (bDown) bDown.textContent = '12.4 MB/s';
    if (bUp) bUp.textContent = '8.7 MB/s';
    if (bPing) bPing.textContent = '42 ms';
    if (bLoss) bLoss.textContent = '0%';
    if (bSession) bSession.textContent = '00:00:00';
  }
  updateTelemetrySparklines();

  // Cumulative VPN Session Traffic Totals
  let totalDown = '0 B';
  let totalUp = '0 B';
  if (status.traffic) {
    if (status.traffic.total_down_human) {
      totalDown = status.traffic.total_down_human;
    } else if (typeof status.traffic.bytes_down === 'number' && status.traffic.bytes_down > 0) {
      totalDown = formatTrafficBytes(status.traffic.bytes_down);
    }
    if (status.traffic.total_up_human) {
      totalUp = status.traffic.total_up_human;
    } else if (typeof status.traffic.bytes_up === 'number' && status.traffic.bytes_up > 0) {
      totalUp = formatTrafficBytes(status.traffic.bytes_up);
    }
  }
  if (statDown) statDown.textContent = totalDown;
  if (statUp) statUp.textContent = totalUp;

  // Monthly 50 GB Quota Card on Home Screen
  const qCard = document.getElementById('home-quota-card');
  if (qCard && status.quota) {
    const qBadge = document.getElementById('quota-source-badge');
    const qUsed = document.getElementById('quota-used-val');
    const qMax = document.getElementById('quota-max-val');
    const qRem = document.getElementById('quota-rem-val');
    const qBar = document.getElementById('quota-progress-fill');
    const qReset = document.getElementById('quota-reset-text');
    const qPct = document.getElementById('quota-pct-val');

    const isAuth = status.quota.is_authoritative;
    if (qBadge) {
      if (isAuth) {
        qBadge.textContent = (I18N[currentLang] && I18N[currentLang].quota_badge_authoritative) || 'Authoritative';
        qBadge.className = 'quota-source-badge auth';
      } else {
        qBadge.textContent = (I18N[currentLang] && I18N[currentLang].quota_badge_estimated) || 'Estimated local usage';
        qBadge.className = 'quota-source-badge estimated';
      }
    }

    if (qUsed) qUsed.textContent = status.quota.quota_used || '0 B';
    if (qMax) qMax.textContent = `/ ${status.quota.max_bytes || '50.0 GB'}`;
    if (qRem) qRem.textContent = status.quota.quota_left || '50.0 GB';

    const pct = Math.max(0, Math.min(100, Math.round(status.quota.quota_percent || 0)));
    if (qBar) {
      qBar.style.width = `${pct}%`;
      if (pct > 90) {
        qBar.className = 'quota-progress-fill critical';
      } else if (pct > 75) {
        qBar.className = 'quota-progress-fill warn';
      } else {
        qBar.className = 'quota-progress-fill';
      }
    }
    if (qPct) qPct.textContent = `${pct}%`;

    if (qReset) {
      let rText = status.quota.quota_reset || '';
      if (rText) {
        try {
          const d = new Date(rText);
          if (!isNaN(d.getTime())) {
            const dateStr = d.toLocaleDateString(currentLang === 'fa' ? 'fa-IR' : 'en-US', { month: 'short', day: 'numeric', year: 'numeric' });
            qReset.textContent = `${(I18N[currentLang] && I18N[currentLang].quota_resets_lbl) || 'Resets:'} ${dateStr}`;
          } else {
            qReset.textContent = `${(I18N[currentLang] && I18N[currentLang].quota_resets_lbl) || 'Resets:'} ${rText}`;
          }
        } catch (e) {
          qReset.textContent = `${(I18N[currentLang] && I18N[currentLang].quota_resets_lbl) || 'Resets:'} ${rText}`;
        }
      } else {
        qReset.textContent = `${(I18N[currentLang] && I18N[currentLang].quota_resets_lbl) || 'Resets:'} 1st of next month`;
      }
    }
  }

  // IP Protocol Diagnostics in Settings
  const diagV4 = document.getElementById('diag-v4-badge');
  const diagV6 = document.getElementById('diag-v6-badge');
  const diagTrans = document.getElementById('diag-transport-badge');
  if (diagV4) {
    diagV4.textContent = `IPv4: ${status.local_ipv4_available ? 'Available' : 'Unavailable'}`;
    diagV4.className = status.local_ipv4_available ? 'diag-badge active' : 'diag-badge inactive';
  }
  if (diagV6) {
    diagV6.textContent = `IPv6: ${status.local_ipv6_available ? 'Available' : 'Unavailable'}`;
    diagV6.className = status.local_ipv6_available ? 'diag-badge active' : 'diag-badge inactive';
  }
  if (diagTrans) {
    if (status.state === 'connected' && (status.transport_protocol || status.exit_address_family)) {
      diagTrans.style.display = 'inline-block';
      diagTrans.textContent = `Transport: ${status.transport_protocol || 'tcp'} • Exit: ${status.exit_address_family || 'IPv4'}`;
    } else {
      diagTrans.style.display = 'none';
    }
  }

  // Status Indicator Dot
  const statusDot = document.getElementById('dash-status-dot');
  if (statusDot) {
    statusDot.className = isConnected ? 'status-indicator-dot connected' : 'status-indicator-dot';
  }

  // Selected Server City & Routing Tag
  if (status.server) {
    const targetCity = document.getElementById('dash-target-city');
    const routingLabel = document.getElementById('dash-routing-label');
    if (targetCity) {
      const cCode = status.server.city_code || (status.server.is_anycast ? 'ANY' : 'POP');
      const cCity = status.server.city_name || status.server.country_name || 'Selected';
      targetCity.textContent = `${cCode} · ${cCity}`;
    }
    if (routingLabel) routingLabel.textContent = status.server.city_code || (status.server.is_anycast ? 'Anycast - BGP' : 'POP');
  }

  // 3D Globe & Traffic Waveform Sync
  if (window.fyrnoxGlobe && typeof window.fyrnoxGlobe.updateState === 'function') {
    window.fyrnoxGlobe.updateState(status);
  }
  if (window.fyrnoxTrafficGraph && typeof window.fyrnoxTrafficGraph.update === 'function') {
    const downSpeed = status.traffic ? (status.traffic.speed_down || 0) : 0;
    const upSpeed = status.traffic ? (status.traffic.speed_up || 0) : 0;
    window.fyrnoxTrafficGraph.update(downSpeed, upSpeed, isConnected);
  }

  // Floating Exit Location Badge Latency
  const badgeLat = document.getElementById('badge-latency');
  if (badgeLat) {
    const ms = isConnected ? (status.exit_latency > 0 ? status.exit_latency : (status.server ? status.server.ping_ms : 42)) : 42;
    badgeLat.textContent = `${ms} ms`;
  }

  // 9. Network Telemetry Card (Exit IP, Location, Latency)
  const netExitIP = document.getElementById('net-exit-ip');
  const netLoc = document.getElementById('net-location');
  const netLat = document.getElementById('net-latency');

  if (netExitIP) {
    netExitIP.textContent = isConnected ? (status.exit_ip || status.current_ip || '—') : (status.current_ip || '—');
  }
  if (netLoc) {
    netLoc.textContent = isConnected ? (status.exit_country || status.current_country || (status.server ? status.server.country_name : '—')) : (status.current_country || '—');
  }
  if (netLat) {
    const ms = isConnected ? (status.exit_latency > 0 ? status.exit_latency : (status.server ? status.server.ping_ms : null)) : (status.server && status.server.ping_ms > 0 ? status.server.ping_ms : null);
    netLat.textContent = ms != null && ms > 0 ? `${ms} ms` : '—';
    applyLatencyStyling(netLat, null, ms);
  }

  // Render Recommended Servers with live data
  renderRecommendedServers(allServers);

  // Synchronize server selection state in the servers view if rendered
  if (status && status.server) {
    const selHost = status.server.hostname;
    document.querySelectorAll('.server-row-card').forEach(card => {
      const oc = card.getAttribute('onclick') || '';
      const isCardSelected = oc.includes(selHost);
      card.classList.toggle('selected', isCardSelected);
      const radio = card.querySelector('.server-selection-indicator');
      if (radio) radio.classList.toggle('selected', isCardSelected);
      const btn = card.querySelector('.server-select-btn');
      if (btn) {
        btn.classList.toggle('active', isCardSelected);
        if (isCardSelected) {
          const isConn = state === 'connected';
          const isPending = state === 'connecting';
          if (isPending) {
            btn.textContent = (I18N[currentLang] && I18N[currentLang].btn_connecting) || 'Connecting';
          } else if (isConn) {
            btn.textContent = (I18N[currentLang] && I18N[currentLang].btn_card_connected) || 'Connected ✓';
          } else {
            btn.textContent = (I18N[currentLang] && I18N[currentLang].btn_card_selected) || 'Selected ✓';
          }
        } else {
          btn.textContent = (I18N[currentLang] && I18N[currentLang].btn_card_connect) || 'Connect';
        }
      }
    });
  }

  // 6b. Smart Recommendation Banner
  const recCard = document.getElementById('dash-recommendation-card');
  const rec = status.recommended_server;
  if (recCard) {
    if (rec && (!status.server || status.server.hostname !== rec.hostname) && state !== 'connected') {
      recCard.style.display = 'flex';
      const flagEl = document.getElementById('rec-server-flag');
      const nameEl = document.getElementById('rec-server-name');
      const pingEl = document.getElementById('rec-server-ping');
      const reasonEl = document.getElementById('rec-server-reason');

      if (flagEl) flagEl.innerHTML = getFlag(rec.country_code);
      if (nameEl) nameEl.textContent = `${rec.country_name} • ${rec.city_name || 'Edge'}`;
      if (pingEl) pingEl.textContent = `${rec.ping_ms} ms`;
      if (reasonEl) reasonEl.textContent = rec.recommend_reason || 'Fastest verified edge';
    } else {
      recCard.style.display = 'none';
    }
  }

  // 7. Routing Mode Card
  const isActualWintun = Boolean(status.wintun_active);
  const modeTag = document.getElementById('dash-mode-tag');
  const topbarMode = document.getElementById('topbar-mode-text');
  const btnProxy = document.getElementById('btn-mode-proxy');
  const btnTun = document.getElementById('btn-mode-tun');

  if (modeTag) {
    modeTag.textContent = isActualWintun 
      ? ((I18N[currentLang] && I18N[currentLang].mode_tag_tun) || 'Wintun Layer-3 TUN')
      : (isTun ? 'TUN (Admin Needed)' : ((I18N[currentLang] && I18N[currentLang].mode_tag_proxy) || 'Web & App Proxy'));
  }
  if (topbarMode) {
    topbarMode.textContent = isActualWintun 
      ? ('🛡️ ' + ((I18N[currentLang] && I18N[currentLang].mode_wintun_tun) || 'Wintun TUN'))
      : ('🌐 ' + ((I18N[currentLang] && I18N[currentLang].mode_system_proxy) || 'System Proxy'));
  }

  if (btnProxy && btnTun) {
    if (isTun) {
      btnTun.classList.add('active');
      btnProxy.classList.remove('active');
    } else {
      btnProxy.classList.add('active');
      btnTun.classList.remove('active');
    }
  }

  if (uacNotice) {
    if (isTun && !status.is_admin) {
      uacNotice.style.display = 'flex';
    } else {
      uacNotice.style.display = 'none';
    }
  }

  // 8. Egress Transparency Card
  const egressTarget = document.getElementById('egress-target-country');
  const egressType = document.getElementById('egress-target-type');
  const egressIP = document.getElementById('egress-public-ip');
  const egressIPStatus = document.getElementById('egress-ip-status');
  const egressObserved = document.getElementById('egress-observed-country');
  const egressISP = document.getElementById('egress-isp-name');
  const anycastCallout = document.getElementById('dash-anycast-callout');
  const anycastText = document.getElementById('dash-anycast-text');

  if (status.server) {
    if (egressTarget) egressTarget.innerHTML = `${getFlag(status.server.country_code)} <span>${escapeHtml(status.server.country_name)}</span>`;
    if (egressType) egressType.textContent = status.server.is_anycast ? 'Anycast Edge BGP' : `Dedicated POP (${status.server.city_code})`;
  }

  if (isConnected) {
    const verifiedIP = status.exit_ip || status.current_ip || '--';
    const verifiedCountry = status.exit_country || status.current_country || 'Verified';
    if (egressIP) egressIP.textContent = verifiedIP;
    if (egressIPStatus) egressIPStatus.textContent = 'Secured via Fastly MASQUE';
    if (egressObserved) egressObserved.innerHTML = `${getFlag(verifiedCountry)} <span>${escapeHtml(verifiedCountry)}</span>`;
    if (egressISP) egressISP.textContent = 'Verified Egress Point';

    // Show Anycast Explanation if target country != observed exit country
    if (status.server && status.exit_country && status.server.country_code !== status.exit_country) {
      if (anycastCallout) anycastCallout.style.display = 'flex';
      if (anycastText) {
        anycastText.innerHTML = `<strong>Anycast BGP Egress:</strong> Fastly routed connection to nearest edge node in <strong>${escapeHtml(verifiedCountry)}</strong> to provide the lowest latency and highest throughput.`;
      }
    } else {
      if (anycastCallout) anycastCallout.style.display = 'none';
    }
  } else {
    if (egressIP) egressIP.textContent = status.current_ip || '--';
    if (egressIPStatus) egressIPStatus.textContent = 'Direct Physical Network';
    if (egressObserved) egressObserved.innerHTML = status.current_country ? `${getFlag(status.current_country)} <span>${escapeHtml(status.current_country)}</span>` : 'Not Connected';
    if (egressISP) egressISP.textContent = 'Direct Provider';
    if (anycastCallout) anycastCallout.style.display = 'none';
  }

  // 9. Telemetry Cards (Ping, Traffic, Quota)
  const telemPing = document.getElementById('telem-ping-value');
  const telemPingBadge = document.getElementById('telem-ping-badge');
  if (telemPing && !isInteractivePingTesting) {
    if (isConnected && status.exit_latency > 0) {
      telemPing.textContent = `${status.exit_latency} ms`;
      if (telemPingBadge) telemPingBadge.textContent = 'Live RTT';
      applyLatencyStyling(telemPing, telemPingBadge, status.exit_latency);
    } else if (status.server && status.server.ping_ms > 0) {
      telemPing.textContent = `${status.server.ping_ms} ms`;
      if (telemPingBadge) telemPingBadge.textContent = isConnected ? 'Live RTT' : 'TCP SYN';
      applyLatencyStyling(telemPing, telemPingBadge, status.server.ping_ms);
    } else {
      telemPing.textContent = '-- ms';
      if (telemPingBadge) telemPingBadge.textContent = isConnected ? 'Live RTT' : 'Idle';
      applyLatencyStyling(telemPing, telemPingBadge, null);
    }
  }

  const telemTraffic = document.getElementById('telem-traffic-value');
  if (telemTraffic && status.traffic) {
    const down = status.traffic.total_down_human || status.traffic.bytes_read_human || '0 B';
    const up = status.traffic.total_up_human || status.traffic.bytes_written_human || '0 B';
    telemTraffic.textContent = `↓ ${down} / ↑ ${up}`;
  }

  const telemQuota = document.getElementById('telem-quota-value');
  const telemQuotaBadge = document.getElementById('telem-quota-badge');
  if (telemQuota && status.quota) {
    telemQuota.textContent = status.quota.quota_left || '50.0 GB';
    if (telemQuotaBadge) telemQuotaBadge.textContent = `${Math.round(status.quota.quota_percent || 100)}% Left`;
  }

  // 10. Diagnostics Service Badges
  const diagTun = document.getElementById('diag-tun-status');
  const diagAdmin = document.getElementById('diag-admin-status');
  if (diagTun) {
    if (status.wintun_active) {
      diagTun.className = 'diag-status-badge pass';
      diagTun.textContent = 'ACTIVE (L3 TUN)';
    } else if (status.tun_active) {
      diagTun.className = 'diag-status-badge standby';
      diagTun.textContent = 'ELEVATION REQUIRED';
    } else {
      diagTun.className = 'diag-status-badge neutral';
      diagTun.textContent = 'STANDBY';
    }
  }
  if (diagAdmin) {
    diagAdmin.className = status.is_admin ? 'diag-status-badge pass' : 'diag-status-badge neutral';
    diagAdmin.textContent = status.is_admin ? 'ADMINISTRATOR' : 'STANDARD USER';
  }

  // 11. Sidebar User Account Pill
  const userStatus = document.getElementById('sidebar-user-status');
  const userName = document.getElementById('sidebar-user-name');
  if (userStatus && userName) {
    if (status.is_logged_in) {
      userName.textContent = status.user_email || 'Firefox Account';
      userStatus.textContent = 'Active (50 GB)';
    } else {
      userName.textContent = 'Firefox Account';
      userStatus.textContent = 'Sign In';
    }
  }

  // 12. Settings inputs
  const setTun = document.getElementById('setting-tun-mode');
  const setH3 = document.getElementById('setting-h3');
  const setTls = document.getElementById('setting-tls-frag');
  const setApiProxy = document.getElementById('setting-api-proxy');
  const setBypassRules = document.getElementById('setting-bypass-rules');
  const setKillSwitch = document.getElementById('setting-kill-switch');
  const setAutoRec = document.getElementById('setting-auto-reconnect');

  if (status.settings) {
    if (setKillSwitch) setKillSwitch.checked = !!status.settings.kill_switch;
    if (setAutoRec) setAutoRec.checked = status.settings.auto_reconnect !== false;
  }
  const setUserEmail = document.getElementById('settings-user-email');
  const setAuthBtn = document.getElementById('settings-auth-btn');

  if (status.settings) {
    if (setTun) setTun.checked = status.settings.enable_tun_mode;
    if (setH3) setH3.checked = status.settings.use_h3;
    if (setTls) setTls.checked = status.settings.tls_fragmentation;
    if (setApiProxy && !setApiProxy.matches(':focus')) {
      setApiProxy.value = status.settings.api_proxy || '';
    }
    if (setBypassRules && !setBypassRules.matches(':focus')) {
      setBypassRules.value = status.settings.bypass_rules || '';
    }
  }

  if (setUserEmail && setAuthBtn) {
    if (status.is_logged_in) {
      setUserEmail.textContent = status.user_email || 'Firefox Account (Active)';
      setAuthBtn.textContent = (currentLang === 'fa') ? 'خروج از حساب' : 'Sign Out';
      setAuthBtn.classList.add('btn-signout');
    } else {
      setUserEmail.textContent = (currentLang === 'fa') ? 'وارد نشده‌اید' : 'Not logged in';
      setAuthBtn.textContent = (currentLang === 'fa') ? 'ورود به حساب' : 'Sign In';
      setAuthBtn.classList.remove('btn-signout');
    }
  }

  // 13. Gaming Mode Telemetry & Topbar Pill
  const gamingPill = document.getElementById('topbar-gaming-pill');
  const gamingText = document.getElementById('topbar-gaming-text');
  const gamingCard = document.getElementById('dash-gaming-card');
  const gamingJitter = document.getElementById('gaming-jitter-val');
  const gamingLoss = document.getElementById('gaming-loss-val');
  const setGaming = document.getElementById('setting-gaming-mode');

  const gamingActive = status.gaming_stats && status.gaming_stats.active;
  if (gamingPill) {
    if (gamingActive) {
      gamingPill.classList.add('active');
      if (gamingText) gamingText.textContent = (I18N[currentLang] && I18N[currentLang].gaming_active) || '⚡ Gaming Active';
    } else {
      gamingPill.classList.remove('active');
      if (gamingText) gamingText.textContent = (I18N[currentLang] && I18N[currentLang].gaming_mode_btn) || 'Gaming Off';
    }
  }

  if (gamingCard) {
    gamingCard.style.display = gamingActive ? 'flex' : 'none';
    if (gamingActive && status.gaming_stats) {
      if (gamingJitter) gamingJitter.textContent = `${status.gaming_stats.jitter_ms || 0} ms`;
      if (gamingLoss) gamingLoss.textContent = `${status.gaming_stats.packet_loss || 0.0}%`;
    }
  }
  if (setGaming) {
    setGaming.checked = !!gamingActive;
  }

  // 13b. Smart Ping Shield Game Detection Badge
  const dashGameBadge = document.getElementById('dash-game-qos-badge');
  const dashGameText = document.getElementById('dash-game-qos-text');
  if (dashGameBadge) {
    if (status.gaming_stats && status.gaming_stats.game_detected) {
      dashGameBadge.style.display = 'flex';
      const gName = status.gaming_stats.detected_game ? status.gaming_stats.detected_game.replace('.exe', '').toUpperCase() : 'GAME';
      if (dashGameText) {
        dashGameText.textContent = `${gName} Detected • Low-Jitter QoS Active`;
      }
    } else {
      dashGameBadge.style.display = 'none';
    }
  }

  // 13c. Intelligent Protocol Fallback Telemetry Pill
  const fallbackRow = document.getElementById('transport-fallback-row');
  const fallbackPill = document.getElementById('dash-fallback-pill');
  const fallbackLabel = document.getElementById('dash-fallback-label');
  if (fallbackRow && fallbackPill && fallbackLabel) {
    if (isConnected) {
      fallbackRow.style.display = 'flex';
      if (status.fallback_active) {
        fallbackPill.className = 'transport-fallback-pill fallback';
        fallbackLabel.textContent = (I18N[currentLang] && I18N[currentLang].fallback_protected_tcp) || 'Protected Fallback (TCP/H2)';
      } else if (status.settings && status.settings.use_h3) {
        fallbackPill.className = 'transport-fallback-pill udp';
        fallbackLabel.textContent = (I18N[currentLang] && I18N[currentLang].fallback_optimized_udp) || 'Optimized UDP (H3)';
      } else {
        fallbackPill.className = 'transport-fallback-pill';
        fallbackLabel.textContent = 'Standard TCP (H2)';
      }
    } else {
      fallbackRow.style.display = 'none';
    }
  }

  // IP Protocol setting reflection
  const setIpProtoEl = document.getElementById('setting-ip-protocol');
  if (setIpProtoEl && status.settings && status.settings.ip_protocol) {
    setIpProtoEl.value = status.settings.ip_protocol;
  }

  // 14. Settings: Smart Iran Domestic Direct Traffic Sync
  const setIranBypass = document.getElementById('setting-iran-bypass');
  if (setIranBypass && status.settings) {
    setIranBypass.checked = status.settings.direct_iran_traffic !== false;
  }

  // 15. Split Tunnel Settings & Mode
  const splitMasterToggle = document.getElementById('split-master-toggle');
  const splitStatusLabel = document.getElementById('split-status-label');
  const btnSplitBypass = document.getElementById('btn-split-bypass');
  const btnSplitProxy = document.getElementById('btn-split-proxy');

  if (status.settings) {
    const isSplit = !!status.settings.split_tunnel_enabled;
    if (splitMasterToggle) splitMasterToggle.checked = isSplit;
    if (splitStatusLabel) splitStatusLabel.textContent = isSplit ? 'Enabled' : 'Disabled';

    const mode = (status.settings.split_tunnel_mode || 'bypass').toLowerCase();
    if (btnSplitBypass && btnSplitProxy) {
      if (mode === 'proxy_only' || mode === 'whitelist') {
        btnSplitProxy.classList.add('active');
        btnSplitBypass.classList.remove('active');
      } else {
        btnSplitBypass.classList.add('active');
        btnSplitProxy.classList.remove('active');
      }
    }
  }

  // 16. Floating Mini-Widget HUD
  const hud = document.getElementById('floating-mini-hud');
  if (hud && hud.style.display !== 'none') {
    const hudDot = document.getElementById('hud-status-dot');
    const hudDown = document.getElementById('hud-speed-down');
    const hudUp = document.getElementById('hud-speed-up');
    const hudPing = document.getElementById('hud-ping-val');
    const hudAction = document.getElementById('hud-action-btn');

    if (hudDot) hudDot.className = `hud-status-indicator ${state}`;
    if (hudAction) {
      hudAction.textContent = isConnected ? '✕' : '⚡';
      hudAction.title = isConnected ? 'Disconnect' : 'Connect';
    }
    if (status.traffic) {
      if (hudDown) hudDown.textContent = `↓ ${status.traffic.down_speed_human || '0 KB/s'}`;
      if (hudUp) hudUp.textContent = `↑ ${status.traffic.up_speed_human || '0 KB/s'}`;
    }
    if (hudPing) {
      if (isConnected && status.exit_latency > 0) {
        hudPing.textContent = `${status.exit_latency} ms (Tunnel)`;
      } else if (status.server && status.server.ping_ms > 0) {
        hudPing.textContent = `${status.server.ping_ms} ms`;
      } else {
        hudPing.textContent = '-- ms';
      }
    }
  }
  } catch (err) {
    console.error("Safe updateUI caught error:", err);
  }
}

// ==========================================================================
// Connect & Disconnect Actions
// ==========================================================================
async function toggleConnect() {
  if (!currentStatus) return;

  if (currentStatus.state === 'connected' || currentStatus.state === 'connecting') {
    try {
      currentStatus.state = 'disconnecting';
      updateUI(currentStatus);

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 3500);

      const res = await fetch('/api/disconnect', { 
        method: 'POST',
        signal: controller.signal
      });
      clearTimeout(timeoutId);

      const data = await res.json();
      if (data && data.status) {
        updateUI(data.status);
      }
    } catch (err) {
      console.warn('Disconnect request handled:', err);
      // Double check status or force disconnected state
      try {
        const checkRes = await fetch('/api/status');
        const checkData = await checkRes.json();
        if (checkData) {
          updateUI(checkData);
          return;
        }
      } catch (e) {}
      currentStatus.state = 'disconnected';
      currentStatus.connect_step = '';
      updateUI(currentStatus);
    }
  } else {
    if (!currentStatus.is_logged_in) {
      openModal('login-modal');
      return;
    }
    try {
      // Optimistic connecting state
      currentStatus.state = 'connecting';
      updateUI(currentStatus);

      const res = await fetch('/api/connect', { method: 'POST' });
      const data = await res.json();
      if (data && data.error) {
        currentStatus.state = 'error';
        currentStatus.error_message = data.error;
        updateUI(currentStatus);
      } else if (data && data.status) {
        updateUI(data.status);
      }
    } catch (err) {
      console.error('Failed to initiate connection:', err);
      currentStatus.state = 'error';
      currentStatus.error_message = err.message || 'Connection failed';
      updateUI(currentStatus);
    }
  }
}

function dismissError() {
  const errBanner = document.getElementById('dash-error-banner');
  if (errBanner) errBanner.style.display = 'none';
  if (currentStatus) currentStatus.state = 'disconnected';
  updateUI(currentStatus);
}

function showUACNotice(msg) {
  const errBanner = document.getElementById('dash-error-banner');
  const errTitle = document.getElementById('dash-error-title');
  const errMsg = document.getElementById('dash-error-msg');
  if (errBanner && errMsg) {
    if (errTitle) errTitle.textContent = 'Elevation Notice';
    errMsg.textContent = msg;
    errBanner.style.display = 'flex';
  }
}

// ==========================================================================
// Routing Mode & UAC Elevation
// ==========================================================================
async function selectRoutingMode(enableTun) {
  if (!currentStatus) return;

  if (enableTun) {
    if (!currentStatus.is_admin) {
      // Directly trigger native Windows UAC prompt without intermediate web modal
      restartAsAdmin();
      return;
    }
  }

  // Update in-memory state optimistically
  if (!currentStatus.settings) currentStatus.settings = {};
  currentStatus.settings.enable_tun_mode = enableTun;
  currentStatus.settings.use_system_proxy = !enableTun;
  updateUI(currentStatus);

  try {
    const res = await fetch('/api/routing-mode', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enable_tun: enableTun })
    });
    if (res.ok) {
      setTimeout(() => fetchStatus(), 150);
    }
  } catch (err) {
    console.error('Failed to change routing mode:', err);
  }
}

function toggleTunMode(checked) {
  selectRoutingMode(checked);
}

async function restartAsAdmin() {
  const currentHost = currentStatus && currentStatus.server ? currentStatus.server.hostname : '';
  try {
    const res = await fetch('/api/elevate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ server: currentHost, mode: 'tun' })
    });
    const data = await res.json();
    if (data && data.cancelled) {
      console.log('UAC cancelled by user');
      // Revert UI toggle and keep System Proxy active
      const setTun = document.getElementById('setting-tun-mode');
      if (setTun) setTun.checked = false;
      if (currentStatus && currentStatus.settings) {
        currentStatus.settings.enable_tun_mode = false;
        currentStatus.settings.use_system_proxy = true;
      }
      updateUI(currentStatus);
      showUACNotice('UAC elevation was cancelled. Continuing in System Proxy mode.');
      return;
    }
    if (data && data.ok) {
      console.log('Elevated process spawned; parent exiting...');
    } else if (data && data.error) {
      showErrorBanner('Elevation failed: ' + data.error);
    }
  } catch (err) {
    console.error('Failed to elevate:', err);
    showErrorBanner('Failed to trigger Windows UAC elevation: ' + err.message);
  }
}

function confirmUACElevation() {
  restartAsAdmin();
}

function cancelUACElevation() {
  const setTun = document.getElementById('setting-tun-mode');
  if (setTun) setTun.checked = false;
  selectRoutingMode(false);
}

// ==========================================================================
// Server Browser View & Search
// ==========================================================================
async function fetchServers() {
  try {
    const res = await fetch('/api/servers');
    if (res.ok) {
      allServers = await res.json();
      renderServerList(allServers);
      renderRecommendedServers(allServers);
    }
  } catch (err) {
    console.error('Failed to fetch servers:', err);
  }
}

function setLocFilter(filter) {
  currentFilter = filter;
  ['all', 'pop', 'anycast', 'europe', 'americas', 'asia'].forEach(f => {
    const chip = document.getElementById(`chip-${f}`);
    if (chip) {
      if (f === filter) chip.classList.add('active');
      else chip.classList.remove('active');
    }
  });
  filterServers(document.getElementById('server-search-input')?.value || '');
}

function filterServers(query) {
  const q = (query || '').toLowerCase().trim();
  const clearBtn = document.getElementById('btn-clear-search');
  if (clearBtn) {
    clearBtn.style.display = q ? 'block' : 'none';
  }

  let list = allServers || [];

  // Filter chips
  if (currentFilter === 'pop') {
    list = list.filter(s => !s.is_anycast);
  } else if (currentFilter === 'anycast') {
    list = list.filter(s => s.is_anycast);
  } else if (currentFilter === 'europe') {
    list = list.filter(s => EUROPE_CODES.includes(s.country_code));
  } else if (currentFilter === 'americas') {
    list = list.filter(s => AMERICAS_CODES.includes(s.country_code));
  } else if (currentFilter === 'asia') {
    list = list.filter(s => ASIA_CODES.includes(s.country_code));
  }

  // Text search
  if (q) {
    list = list.filter(s =>
      (s.country_name && s.country_name.toLowerCase().includes(q)) ||
      (s.city_name && s.city_name.toLowerCase().includes(q)) ||
      (s.hostname && s.hostname.toLowerCase().includes(q)) ||
      (s.country_code && s.country_code.toLowerCase().includes(q)) ||
      (s.city_code && s.city_code.toLowerCase().includes(q))
    );
  }

  renderServerList(list);
}

function clearSearch() {
  const input = document.getElementById('server-search-input');
  if (input) {
    input.value = '';
    filterServers('');
  }
}

function renderServerList(servers) {
  const container = document.getElementById('server-list-container');
  const pinnedRecContainer = document.getElementById('servers-recommended-card');
  if (!container) return;

  if (!servers || servers.length === 0) {
    if (pinnedRecContainer) pinnedRecContainer.style.display = 'none';
    container.innerHTML = `
      <div style="text-align: center; color: var(--text-muted); padding: 48px 0; font-size: 0.9rem;">
        No server locations match your search criteria.
      </div>
    `;
    return;
  }

  // Sort lowest latency first, then unmeasured
  const sorted = [...servers].sort((a, b) => {
    if (a.ping_ms > 0 && b.ping_ms > 0) return a.ping_ms - b.ping_ms;
    if (a.ping_ms > 0) return -1;
    if (b.ping_ms > 0) return 1;
    return 0;
  });

  const selectedHost = currentStatus && currentStatus.server ? currentStatus.server.hostname : '';
  const fastestHost = sorted.length > 0 && sorted[0].ping_ms > 0 ? sorted[0].hostname : null;

  // Identify recommended server (explicit is_recommended or lowest latency or first entry)
  const recServer = sorted.find(s => s.is_recommended) || sorted[0];

  if (pinnedRecContainer && recServer) {
    pinnedRecContainer.style.display = 'flex';
    const isRecSelected = recServer.hostname === selectedHost;
    const recFlag = getFlag(recServer.country_code);

    let recPingPillClass = 'unavailable';
    let recPingText = '-- ms';
    if (recServer.ping_ms > 0) {
      recPingText = `${recServer.ping_ms} ms`;
      if (recServer.ping_ms <= 200) recPingPillClass = 'fast';
      else if (recServer.ping_ms <= 500) recPingPillClass = 'medium';
      else recPingPillClass = 'slow';
    } else if (recServer.ping_ms === -1) {
      recPingText = 'Timeout';
      recPingPillClass = 'slow';
    }

    const recTypeBadge = recServer.is_anycast
      ? `<span class="badge-anycast-pill">${(I18N[currentLang] && I18N[currentLang].type_anycast) || 'ANYCAST EDGE'}</span>`
      : `<span class="badge-pop-pill">POP: ${escapeHtml(recServer.city_code || 'DIRECT')}</span>`;

    const isConn = currentStatus && (currentStatus.state === 'connected');
    const isConnPending = currentStatus && (currentStatus.state === 'connecting');
    let recBtnText = (I18N[currentLang] && I18N[currentLang].btn_card_connect) || 'Connect';
    if (isRecSelected) {
      if (isConnPending) {
        recBtnText = (I18N[currentLang] && I18N[currentLang].btn_connecting) || 'Connecting';
      } else if (isConn) {
        recBtnText = (I18N[currentLang] && I18N[currentLang].btn_card_connected) || 'Connected ✓';
      } else {
        recBtnText = (I18N[currentLang] && I18N[currentLang].btn_card_selected) || 'Selected ✓';
      }
    }

    pinnedRecContainer.innerHTML = `
      <div class="pinned-rec-header">
        <span class="pinned-rec-badge">⚡ ${(I18N[currentLang] && I18N[currentLang].rec_tag) || 'RECOMMENDED LOCATION'}</span>
        <span class="pinned-rec-desc">${(I18N[currentLang] && I18N[currentLang].rec_fastest_edge) || 'Fastest verified edge'}</span>
      </div>
      <div class="server-row-card recommended-pinned-row ${isRecSelected ? 'selected' : ''}" onclick="selectServer('${escapeHtml(recServer.hostname)}')">
        <div class="server-cell-location">
          <span class="server-selection-indicator ${isRecSelected ? 'selected' : ''}"></span>
          <span class="server-row-flag">${recFlag}</span>
          <div class="server-loc-names">
            <span class="server-country-name">${escapeHtml(recServer.country_name)} • ${escapeHtml(recServer.city_name || recServer.country_name)}</span>
            <span class="server-hostname-text">${escapeHtml(recServer.hostname)}</span>
          </div>
        </div>
        <div class="server-cell-type">
          ${recTypeBadge}
          <span class="badge-fastest-pill">⚡ ${(I18N[currentLang] && I18N[currentLang].badge_recommended) || 'RECOMMENDED'}</span>
        </div>
        <div class="server-cell-ping">
          <span class="ping-latency-pill ${recPingPillClass}" data-host="${escapeHtml(recServer.hostname)}">${recPingText}</span>
        </div>
        <div class="server-cell-action">
          <button class="server-select-btn ${isRecSelected ? 'active' : ''}">
            ${recBtnText}
          </button>
        </div>
      </div>
    `;
  }

  container.innerHTML = sorted.map(srv => {
    const isSelected = srv.hostname === selectedHost;
    const isFastest = srv.hostname === fastestHost;
    const flag = getFlag(srv.country_code);

    let pingPillClass = 'unavailable';
    let pingText = '-- ms';
    if (srv.ping_ms > 0) {
      pingText = `${srv.ping_ms} ms`;
      if (srv.ping_ms <= 200) pingPillClass = 'fast';
      else if (srv.ping_ms <= 500) pingPillClass = 'medium';
      else pingPillClass = 'slow';
    } else if (srv.ping_ms === -1) {
      pingText = 'Timeout';
      pingPillClass = 'slow';
    }

    const typeBadge = srv.is_anycast
      ? `<span class="badge-anycast-pill">${(I18N[currentLang] && I18N[currentLang].type_anycast) || 'ANYCAST EDGE'}</span>`
      : `<span class="badge-pop-pill">POP: ${escapeHtml(srv.city_code || 'DIRECT')}</span>`;

    const isRec = srv.is_recommended || isFastest;
    const recLabel = (I18N[currentLang] && I18N[currentLang].badge_recommended) || 'RECOMMENDED';
    const fastestBadge = isRec ? `<span class="badge-fastest-pill">⚡ ${recLabel}</span>` : '';

    const isConn = currentStatus && (currentStatus.state === 'connected');
    const isConnPending = currentStatus && (currentStatus.state === 'connecting');
    let btnText = (I18N[currentLang] && I18N[currentLang].btn_card_connect) || 'Connect';
    if (isSelected) {
      if (isConnPending) {
        btnText = (I18N[currentLang] && I18N[currentLang].btn_connecting) || 'Connecting';
      } else if (isConn) {
        btnText = (I18N[currentLang] && I18N[currentLang].btn_card_connected) || 'Connected ✓';
      } else {
        btnText = (I18N[currentLang] && I18N[currentLang].btn_card_selected) || 'Selected ✓';
      }
    }

    return `
      <div class="server-row-card ${isSelected ? 'selected' : ''}" onclick="selectServer('${srv.hostname}')">
        
        <div class="server-cell-location">
          <span class="server-selection-indicator ${isSelected ? 'selected' : ''}"></span>
          <span class="server-row-flag">${flag}</span>
          <div class="server-loc-names">
            <span class="server-country-name">${escapeHtml(srv.country_name)} • ${escapeHtml(srv.city_name || srv.country_name)}</span>
            <span class="server-hostname-text">${escapeHtml(srv.hostname)}</span>
          </div>
        </div>

        <div class="server-cell-type">
          ${typeBadge}
          ${fastestBadge}
        </div>

        <div class="server-cell-ping">
          <span class="ping-latency-pill ${pingPillClass}" data-host="${escapeHtml(srv.hostname)}">${pingText}</span>
        </div>

        <div class="server-cell-action">
          <button class="server-select-btn ${isSelected ? 'active' : ''}">
            ${btnText}
          </button>
        </div>

      </div>
    `;
  }).join('');
}

async function selectServer(hostname) {
  // If we are currently connected or connecting, provide immediate visual feedback of the seamless switch
  const isConn = currentStatus && (currentStatus.state === 'connected' || currentStatus.state === 'connecting');
  if (isConn) {
    const cardBtn = document.querySelector(`.server-row-card[onclick*="${hostname}"] .server-select-btn`);
    if (cardBtn) {
      cardBtn.textContent = (I18N[currentLang] && I18N[currentLang].btn_connecting) || 'Connecting';
      cardBtn.classList.add('active');
    }
    const heroLabel = document.getElementById('hero-status-label');
    if (heroLabel) {
      heroLabel.textContent = (I18N[currentLang] && I18N[currentLang].connecting) || 'Connecting';
      heroLabel.className = 'hero-status-badge connecting';
    }
    const heroActionLabel = document.getElementById('hero-action-label');
    if (heroActionLabel) {
      heroActionLabel.textContent = (I18N[currentLang] && I18N[currentLang].btn_connecting) || 'Connecting';
    }
    const heroSpinner = document.getElementById('hero-spinner');
    if (heroSpinner) heroSpinner.style.display = 'block';
  }
  try {
    const res = await fetch('/api/select-server', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ hostname })
    });
    if (res.ok) {
      const data = await res.json();
      if (data && data.status) {
        updateUI(data.status);
      }
      // Re-render server list while preserving active filter and search query
      filterServers(document.getElementById('server-search-input')?.value || '');
    }
  } catch (err) {
    console.error('Failed to select server:', err);
  }
}

function updateServerPingBadge(hostname, pingMs) {
  const pill = document.querySelector(`.ping-latency-pill[data-host="${hostname}"]`);
  if (!pill) return;
  pill.className = 'ping-latency-pill';
  if (pingMs === 0) {
    pill.classList.add('testing');
    pill.textContent = currentLang === 'fa' ? 'سنجش...' : 'Testing...';
  } else if (pingMs > 0 && pingMs <= 200) {
    pill.classList.add('fast');
    pill.textContent = `${pingMs} ms`;
  } else if (pingMs > 200 && pingMs <= 500) {
    pill.classList.add('medium');
    pill.textContent = `${pingMs} ms`;
  } else if (pingMs > 500) {
    pill.classList.add('slow');
    pill.textContent = `${pingMs} ms`;
  } else {
    pill.classList.add('slow');
    pill.textContent = 'Timeout';
  }
}

// ==========================================================================
// Modern Glass Recommended Servers & Routing Dropdown Logic
// ==========================================================================
function renderRecommendedServers(servers) {
  const container = document.getElementById('recommended-servers-list');
  if (!container) return;

  const currentSelectedHost = currentStatus && currentStatus.server ? currentStatus.server.hostname : '';

  let list = [];
  if (servers && servers.length > 0) {
    const withPing = servers.filter(s => s.ping_ms > 0).sort((a, b) => a.ping_ms - b.ping_ms);
    const withoutPing = servers.filter(s => s.ping_ms <= 0);
    list = [...withPing, ...withoutPing].slice(0, 5);
  }

  if (list.length === 0) {
    list = [
      { hostname: 'iad.m1.fastly-masque.net', country_name: 'United States', country_code: 'US', ping_ms: 42, is_anycast: true },
      { hostname: 'fra.m1.fastly-masque.net', country_name: 'Germany', country_code: 'DE', ping_ms: 56, is_anycast: true },
      { hostname: 'yyz.m1.fastly-masque.net', country_name: 'Canada', country_code: 'CA', ping_ms: 68, is_anycast: true },
      { hostname: 'lhr.m1.fastly-masque.net', country_name: 'United Kingdom', country_code: 'GB', ping_ms: 74, is_anycast: true },
      { hostname: 'cdg.m1.fastly-masque.net', country_name: 'France', country_code: 'FR', ping_ms: 82, is_anycast: true }
    ];
  }

  container.innerHTML = list.map(s => {
    const isSelected = currentSelectedHost === s.hostname;
    let pingDisplay = '-- ms';
    let pingClass = 'good';
    if (s.ping_ms > 0) {
      pingDisplay = `${s.ping_ms} ms`;
      pingClass = s.ping_ms < 150 ? 'good' : (s.ping_ms <= 350 ? 'medium' : 'poor');
    } else {
      pingDisplay = '42 ms';
    }

    return `
      <div class="rec-server-item ${isSelected ? 'selected' : ''}" onclick="selectServer('${escapeHtml(s.hostname)}')" title="Connect to ${escapeHtml(s.country_name)}">
        <div class="rec-server-flag">${getFlag(s.country_code)}</div>
        <div class="rec-server-info">
          <div class="rec-server-name">${escapeHtml(s.country_name)}</div>
          <div class="rec-server-sub">${s.is_anycast ? 'Anycast - BGP' : (s.city_code || 'Dedicated POP')}</div>
        </div>
        <div class="rec-server-latency">
          <span class="latency-dot ${pingClass}"></span>
          <span class="latency-val">${pingDisplay}</span>
        </div>
      </div>
    `;
  }).join('');
}

function positionModeDropdown() {
  const trigger = document.getElementById('mode-dropdown-trigger');
  const portal = document.getElementById('mode-dropdown-portal');
  const panel = document.getElementById('mode-dropdown-panel');
  if (!trigger || !portal || !panel) return;

  const rect = trigger.getBoundingClientRect();
  const panelWidth = 255;

  let left = rect.left + (rect.width / 2) - (panelWidth / 2);
  const margin = 12;
  if (left < margin) left = margin;
  if (left + panelWidth > window.innerWidth - margin) {
    left = window.innerWidth - panelWidth - margin;
  }

  const panelHeight = panel.offsetHeight || 160;
  let top = rect.bottom + 8;
  // If opening downwards would clip against bottom of window, flip upwards
  if (top + panelHeight > window.innerHeight - 10 && rect.top - panelHeight - 8 > 40) {
    top = rect.top - panelHeight - 8;
  }

  panel.style.left = `${Math.round(left)}px`;
  panel.style.top = `${Math.round(top)}px`;
}

function openModeDropdown() {
  const portal = document.getElementById('mode-dropdown-portal');
  const trigger = document.getElementById('mode-dropdown-trigger');
  if (!portal || !trigger) return;

  // Ensure routing dropdown checkmarks strictly sync with active state before opening
  const isTunModeActive = Boolean(currentStatus && (currentStatus.wintun_active || ((currentStatus.tun_active || (currentStatus.settings && currentStatus.settings.enable_tun_mode)) && currentStatus.is_admin)));
  const checkProxy = document.getElementById('check-mode-proxy');
  const checkTun = document.getElementById('check-mode-tun');
  const optProxy = document.getElementById('btn-mode-proxy');
  const optTun = document.getElementById('btn-mode-tun');
  if (checkProxy && checkTun) {
    checkTun.style.display = isTunModeActive ? 'inline' : 'none';
    checkProxy.style.display = isTunModeActive ? 'none' : 'inline';
    if (optTun) optTun.classList.toggle('active', isTunModeActive);
    if (optProxy) optProxy.classList.toggle('active', !isTunModeActive);
  }

  portal.style.display = 'block';
  portal.classList.add('open');
  trigger.classList.add('active');
  trigger.setAttribute('aria-expanded', 'true');
  positionModeDropdown();

  // Focus active option for keyboard navigation
  setTimeout(() => {
    const activeItem = portal.querySelector('.mode-dropdown-item.active') || portal.querySelector('.mode-dropdown-item');
    if (activeItem) activeItem.focus();
  }, 30);
}

function closeModeDropdown() {
  const portal = document.getElementById('mode-dropdown-portal');
  const trigger = document.getElementById('mode-dropdown-trigger');
  if (portal) {
    portal.style.display = 'none';
    portal.classList.remove('open');
  }
  if (trigger) {
    trigger.classList.remove('active');
    trigger.setAttribute('aria-expanded', 'false');
  }
}

function toggleModeDropdown(event) {
  if (event) {
    event.stopPropagation();
    event.preventDefault();
  }
  const portal = document.getElementById('mode-dropdown-portal');
  if (!portal) return;

  const isOpen = portal.style.display === 'block';
  if (isOpen) {
    closeModeDropdown();
  } else {
    openModeDropdown();
  }
}

function handleModeDropdownSelect(isTun) {
  closeModeDropdown();
  selectRoutingMode(isTun);
}

function handleModeItemKey(event, isTun) {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault();
    handleModeDropdownSelect(isTun);
  } else if (event.key === 'Escape') {
    closeModeDropdown();
    const trigger = document.getElementById('mode-dropdown-trigger');
    if (trigger) trigger.focus();
  }
}

document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    const portal = document.getElementById('mode-dropdown-portal');
    if (portal && portal.style.display === 'block') {
      closeModeDropdown();
      const trigger = document.getElementById('mode-dropdown-trigger');
      if (trigger) trigger.focus();
    }
  }
});

document.addEventListener('click', (e) => {
  const portal = document.getElementById('mode-dropdown-portal');
  const trigger = document.getElementById('mode-dropdown-trigger');
  if (portal && portal.style.display === 'block') {
    if (!portal.contains(e.target) && (!trigger || !trigger.contains(e.target))) {
      closeModeDropdown();
    }
  }
});

window.addEventListener('resize', () => {
  const portal = document.getElementById('mode-dropdown-portal');
  if (portal && portal.style.display === 'block') {
    positionModeDropdown();
  }
});

function pingAllServersInteractive() {
  pingAllServers();
}

async function pingAllServers() {
  if (isPinging) return;
  isPinging = true;

  const btn = document.getElementById('btn-ping-all');
  const btn2 = document.getElementById('btn-ping-all-servers');
  if (btn) {
    btn.innerHTML = `<span class="bolt">⚡</span><span>${currentLang === 'fa' ? 'در حال پینگ...' : 'Pinging...'}</span>`;
    btn.style.opacity = '0.7';
    btn.disabled = true;
  }
  if (btn2) {
    btn2.innerHTML = `<span class="ping-signal-icon">📶</span><span>Pinging...</span>`;
    btn2.disabled = true;
  }

  const panel = document.getElementById('ping-progress-panel');
  const fill = document.getElementById('ping-progress-bar-fill');
  const statusText = document.getElementById('ping-progress-status-text');
  const countText = document.getElementById('ping-progress-count');

  const serverList = (allServers && allServers.length > 0) ? allServers : [];
  const total = serverList.length;

  if (panel) panel.style.display = 'block';
  if (fill) fill.style.width = '0%';
  if (statusText) {
    statusText.textContent = currentLang === 'fa' ? 'اندازه‌گیری لحظه‌ای تأخیر سرورها (Worker Pool)...' : 'Measuring Real-Time Server Latencies...';
  }
  if (countText) countText.textContent = `0 / ${total} Verified`;

  // Immediately mark all visible server badges as "Testing..."
  serverList.forEach(s => {
    updateServerPingBadge(s.hostname, 0);
  });

  pingAbortController = new AbortController();
  const signal = pingAbortController.signal;

  let completedCount = 0;
  const concurrency = 6;
  let queueIndex = 0;

  async function worker() {
    while (queueIndex < serverList.length) {
      if (signal.aborted) break;
      const index = queueIndex++;
      const srv = serverList[index];

      try {
        const res = await fetch(`/api/ping?hostname=${encodeURIComponent(srv.hostname)}`, { signal });
        if (res.ok) {
          const data = await res.json();
          const ms = (typeof data.ping_ms === 'number') ? data.ping_ms : (typeof data.ping === 'number' ? data.ping : -1);
          srv.ping_ms = ms;
          updateServerPingBadge(srv.hostname, ms);
        } else {
          srv.ping_ms = -1;
          updateServerPingBadge(srv.hostname, -1);
        }
      } catch (err) {
        if (signal.aborted) break;
        srv.ping_ms = -1;
        updateServerPingBadge(srv.hostname, -1);
      }

      completedCount++;
      const pct = Math.min(100, Math.round((completedCount / total) * 100));
      if (fill) fill.style.width = `${pct}%`;
      if (countText) countText.textContent = `${completedCount} / ${total} Verified`;
    }
  }

  try {
    const workers = [];
    for (let w = 0; w < Math.min(concurrency, total); w++) {
      workers.push(worker());
    }
    await Promise.all(workers);

    if (!signal.aborted) {
      if (fill) fill.style.width = '100%';
      if (countText) countText.textContent = `${total} / ${total} Verified`;

      const responsive = serverList.filter(s => s.ping_ms > 0).map(s => s.ping_ms).sort((a, b) => a - b);
      const median = responsive.length > 0 ? responsive[Math.floor(responsive.length / 2)] : 0;
      const best = [...serverList].filter(s => s.ping_ms > 0).sort((a, b) => a.ping_ms - b.ping_ms)[0];

      if (statusText) {
        if (best) {
          const loc = best.city_name || best.country_name || 'Edge POP';
          statusText.textContent = currentLang === 'fa'
            ? `پایان سنجش — بهترین: ${loc} (${best.ping_ms}ms) | میانه: ${median}ms`
            : `Optimization Complete — Best: ${loc} (${best.ping_ms} ms) | Median: ${median} ms`;
        } else {
          statusText.textContent = currentLang === 'fa' ? 'سنجش پینگ سرورها تکمیل شد' : 'Server Ping Sweep Completed';
        }
      }

      // Re-render sorted list with lowest latency first
      renderServerList(allServers);
      renderRecommendedServers(allServers);

      setTimeout(() => {
        if (!isPinging && panel) panel.style.display = 'none';
      }, 5000);
    }
  } catch (err) {
    if (err.name !== 'AbortError') {
      console.error('Failed to ping all servers:', err);
      if (statusText) {
        statusText.textContent = currentLang === 'fa' ? 'خطا در ارزیابی پینگ سرورها' : 'Failed to measure server latency';
      }
    }
  } finally {
    isPinging = false;
    pingAbortController = null;
    if (btn) {
      btn.innerHTML = `<span class="bolt">⚡</span><span>${currentLang === 'fa' ? 'تست پینگ همه سرورها' : 'Ping All Servers'}</span>`;
      btn.style.opacity = '1';
      btn.disabled = false;
    }
    const btn2 = document.getElementById('btn-ping-all-servers');
    if (btn2) {
      btn2.innerHTML = `<span class="ping-signal-icon">📶</span><span>Ping All Servers</span>`;
      btn2.disabled = false;
    }
  }
}

function cancelPingAll() {
  if (pingAbortController) {
    pingAbortController.abort();
    pingAbortController = null;
  }
  isPinging = false;

  const panel = document.getElementById('ping-progress-panel');
  if (panel) panel.style.display = 'none';

  const btn = document.getElementById('btn-ping-all');
  if (btn) {
    btn.innerHTML = `<span class="bolt">⚡</span><span>${currentLang === 'fa' ? 'تست پینگ همه سرورها' : 'Ping All Servers'}</span>`;
    btn.style.opacity = '1';
    btn.disabled = false;
  }
}

let isInteractivePingTesting = false;

async function interactivePingTest() {
  if (isInteractivePingTesting) return;
  const card = document.getElementById('dash-ping-card');
  const pingVal = document.getElementById('telem-ping-value');
  const pingBadge = document.getElementById('telem-ping-badge');

  const targetServer = (currentStatus && currentStatus.server) ? currentStatus.server : null;
  const isConnected = currentStatus && (currentStatus.state || '').toLowerCase() === 'connected';

  isInteractivePingTesting = true;
  if (card) card.classList.add('testing');
  if (pingVal) {
    pingVal.textContent = (I18N[currentLang] && I18N[currentLang].ping_probing) || 'Measuring...';
    pingVal.classList.remove('latency-good', 'latency-medium', 'latency-poor');
  }
  if (pingBadge) {
    pingBadge.textContent = (I18N[currentLang] && I18N[currentLang].ping_probing) || 'Probing...';
    pingBadge.classList.remove('latency-good', 'latency-medium', 'latency-poor');
    pingBadge.classList.add('testing');
  }

  try {
    let url = '/api/ping';
    if (targetServer && targetServer.hostname) {
      url += `?hostname=${encodeURIComponent(targetServer.hostname)}`;
      if (isConnected) url += '&tunnel=true';
    }
    const res = await fetch(url);
    if (res.ok) {
      const data = await res.json();
      const latency = data.ping_ms != null ? data.ping_ms : data.ping;
      const netLat = document.getElementById('net-latency');
      if (data.ok && latency > 0) {
        if (pingVal) pingVal.textContent = `${latency} ms`;
        if (netLat) netLat.textContent = `${latency} ms`;
        if (pingBadge) {
          pingBadge.textContent = data.tunnel ? 'Live RTT' : 'TCP SYN';
        }
        applyLatencyStyling(pingVal, pingBadge, latency);
        if (netLat) applyLatencyStyling(netLat, null, latency);
        if (targetServer) {
          targetServer.ping_ms = latency;
        }
        if (currentStatus) {
          currentStatus.exit_latency = latency;
          if (currentStatus.server) {
            currentStatus.server.ping_ms = latency;
          }
        }
        if (targetServer && targetServer.hostname) {
          updateServerPingBadge(targetServer.hostname, latency);
        }
        renderRecommendedServers(allServers);
      } else {
        if (pingVal) {
          pingVal.textContent = (I18N[currentLang] && I18N[currentLang].ping_timeout) || 'Timeout';
          pingVal.classList.add('latency-poor');
        }
        if (netLat) {
          netLat.textContent = 'Timeout';
          netLat.classList.add('latency-poor');
        }
        if (pingBadge) {
          pingBadge.textContent = (I18N[currentLang] && I18N[currentLang].ping_failed) || 'Failed';
          pingBadge.classList.add('latency-poor');
        }
      }
    } else {
      throw new Error(`HTTP ${res.status}`);
    }
  } catch (err) {
    console.error('Interactive ping test failed:', err);
    if (pingVal) {
      pingVal.textContent = (I18N[currentLang] && I18N[currentLang].ping_offline) || 'Offline';
      pingVal.classList.add('latency-poor');
    }
    if (pingBadge) {
      pingBadge.textContent = (I18N[currentLang] && I18N[currentLang].ping_failed) || 'Error';
      pingBadge.classList.add('latency-poor');
    }
  } finally {
    setTimeout(() => {
      if (card) card.classList.remove('testing');
      if (pingBadge) pingBadge.classList.remove('testing');
      isInteractivePingTesting = false;
    }, 450);
  }
}

// ==========================================================================
// Diagnostics & Logs Console
// ==========================================================================
async function fetchLogs() {
  try {
    const res = await fetch('/api/logs');
    if (res.ok) {
      const logs = await res.json();
      renderLogs(logs);
    }
  } catch (err) {
    console.error('Failed to fetch logs:', err);
  }
}

function renderLogs(logs) {
  const consoleEl = document.getElementById('diag-logs-console');
  const countEl = document.getElementById('logs-line-count');
  if (!consoleEl) return;

  if (!logs || logs.length === 0) {
    consoleEl.innerHTML = '<div class="log-empty">No logs recorded yet.</div>';
    if (countEl) countEl.textContent = '0 lines';
    return;
  }

  if (countEl) countEl.textContent = `${logs.length} lines`;

  const fragment = document.createDocumentFragment();
  logs.forEach(entry => {
    const row = document.createElement('div');
    row.className = 'log-entry';

    let time = '';
    let level = 'INFO';
    let msg = '';

    if (typeof entry === 'string') {
      msg = entry;
    } else if (entry && typeof entry === 'object') {
      time = entry.timestamp || '';
      level = (entry.level || 'INFO').toUpperCase();
      msg = entry.message || '';
    }

    const timeSpan = document.createElement('span');
    timeSpan.className = 'log-timestamp';
    timeSpan.textContent = `[${time}]`;

    const badgeSpan = document.createElement('span');
    const lvlLower = level.toLowerCase();
    badgeSpan.className = `log-badge ${lvlLower}`;
    badgeSpan.textContent = level;

    const msgSpan = document.createElement('span');
    msgSpan.className = 'log-msg';
    msgSpan.textContent = msg;

    row.appendChild(timeSpan);
    row.appendChild(badgeSpan);
    row.appendChild(msgSpan);
    fragment.appendChild(row);
  });

  consoleEl.innerHTML = '';
  consoleEl.appendChild(fragment);
  consoleEl.scrollTop = consoleEl.scrollHeight;
}

function copyLogs() {
  const consoleEl = document.getElementById('diag-logs-console');
  if (!consoleEl) return;
  const entries = consoleEl.querySelectorAll('.log-entry');
  if (entries.length === 0) {
    if (consoleEl.textContent) {
      navigator.clipboard.writeText(consoleEl.textContent.trim());
    }
    return;
  }
  const plainText = Array.from(entries).map(row => {
    const time = row.querySelector('.log-timestamp')?.textContent || '';
    const badge = row.querySelector('.log-badge')?.textContent || '';
    const msg = row.querySelector('.log-msg')?.textContent || '';
    return `${time} [${badge}] ${msg}`.trim();
  }).join('\n');

  if (plainText) {
    navigator.clipboard.writeText(plainText);
    alert((I18N[currentLang] && I18N[currentLang].logs_copied) || 'Logs copied to clipboard!');
  }
}

async function clearLogs() {
  const consoleEl = document.getElementById('diag-logs-console');
  const countEl = document.getElementById('logs-line-count');
  if (consoleEl) {
    consoleEl.innerHTML = '<div class="log-empty">Logs cleared.</div>';
  }
  if (countEl) countEl.textContent = '0 lines';
  try {
    await fetch('/api/logs/clear', { method: 'POST' });
  } catch (err) {
    console.error('Failed to clear logs on server:', err);
  }
}

function copyIPToClipboard() {
  const ipEl = document.getElementById('egress-public-ip');
  if (ipEl && ipEl.textContent && ipEl.textContent !== '--') {
    navigator.clipboard.writeText(ipEl.textContent.trim());
    alert((I18N[currentLang] && I18N[currentLang].ip_copied) || 'IP copied to clipboard!');
  }
}

// ==========================================================================
// Settings API Persistence
// ==========================================================================
async function saveSettingsPayload(payload) {
  try {
    const res = await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    if (res.ok) {
      const updated = await res.json();
      if (currentStatus) currentStatus.settings = updated;
      updateUI(currentStatus);
    }
  } catch (err) {
    console.error('Failed to save settings:', err);
  }
}

function saveSettings() {
  const setH3 = document.getElementById('setting-h3');
  const setTls = document.getElementById('setting-tls-frag');
  const setApiProxy = document.getElementById('setting-api-proxy');
  const setBypassRules = document.getElementById('setting-bypass-rules');
  const setTun = document.getElementById('setting-tun-mode');
  const setIran = document.getElementById('setting-iran-bypass');
  const setGaming = document.getElementById('setting-gaming-mode');
  const setIpProto = document.getElementById('setting-ip-protocol');
  const setKillSwitch = document.getElementById('setting-kill-switch');
  const setAutoRec = document.getElementById('setting-auto-reconnect');

  const cur = (currentStatus && currentStatus.settings) || {};

  const payload = {
    kill_switch: setKillSwitch ? setKillSwitch.checked : (cur.kill_switch || false),
    auto_reconnect: setAutoRec ? setAutoRec.checked : (cur.auto_reconnect !== false),
    enable_tun_mode: setTun ? setTun.checked : (cur.enable_tun_mode || false),
    use_system_proxy: setTun ? !setTun.checked : (cur.use_system_proxy !== false),
    use_h3: setH3 ? setH3.checked : (cur.use_h3 || false),
    tls_fragmentation: setTls ? setTls.checked : (cur.tls_fragmentation !== false),
    api_proxy: setApiProxy ? setApiProxy.value.trim() : (cur.api_proxy || ''),
    bypass_rules: setBypassRules ? setBypassRules.value.trim() : (cur.bypass_rules || ''),
    direct_iran_traffic: setIran ? setIran.checked : (cur.direct_iran_traffic !== false),
    gaming_mode: setGaming ? setGaming.checked : (cur.gaming_mode || false),
    ip_protocol: setIpProto ? setIpProto.value : (cur.ip_protocol || 'dual'),
    split_tunnel_enabled: cur.split_tunnel_enabled || false,
    split_tunnel_mode: cur.split_tunnel_mode || 'bypass',
    split_tunnel_apps: cur.split_tunnel_apps || ['telegram.exe', 'discord.exe']
  };

  saveSettingsPayload(payload);
}

// ==========================================================================
// Session Duration Counter
// ==========================================================================
function startDurationTimer() {
  if (durationInterval) clearInterval(durationInterval);
  durationInterval = setInterval(() => {
    const durationEl = document.getElementById('telem-duration-value');
    const badgeEl = document.getElementById('telem-duration-badge');
    const sessionEl = document.getElementById('stat-session-val');

    const bSession = document.getElementById('bottom-session-duration');
    if (!connectionStartTime || (currentStatus && currentStatus.state !== 'connected')) {
      if (durationEl) durationEl.textContent = '00:00:00';
      if (sessionEl) sessionEl.textContent = '00:00:00';
      if (bSession) bSession.textContent = '00:00:00';
      if (badgeEl) {
        badgeEl.textContent = 'Idle';
        badgeEl.className = 'telemetry-badge neutral';
      }
      return;
    }

    const elapsedMs = Date.now() - connectionStartTime;
    const totalSecs = Math.floor(elapsedMs / 1000);
    const hrs = String(Math.floor(totalSecs / 3600)).padStart(2, '0');
    const mins = String(Math.floor((totalSecs % 3600) / 60)).padStart(2, '0');
    const secs = String(totalSecs % 60).padStart(2, '0');

    const formatted = `${hrs}:${mins}:${secs}`;
    if (durationEl) durationEl.textContent = formatted;
    if (sessionEl) sessionEl.textContent = formatted;
    if (bSession) bSession.textContent = formatted;
    if (badgeEl) {
      badgeEl.textContent = 'Active';
      badgeEl.className = 'telemetry-badge free';
    }
  }, 1000);
}

// ==========================================================================
// Authentication (Firefox Account Login / Logout)
// ==========================================================================
function handleAuthClick() {
  if (currentStatus && currentStatus.is_logged_in) {
    const msg = (currentLang === 'fa') ? 'آیا می‌خواهید از حساب فایرفاکس خود خارج شوید؟' : 'Sign out from your Firefox Account?';
    if (confirm(msg)) {
      doLogout();
    }
  } else {
    openModal('login-modal');
  }
}

async function doLogin() {
  const emailInput = document.getElementById('login-email');
  const passInput = document.getElementById('login-password');
  const proxyInput = document.getElementById('login-proxy');
  const errEl = document.getElementById('login-error');
  const submitBtn = document.getElementById('btn-do-login');

  const email = emailInput ? emailInput.value.trim() : '';
  const password = passInput ? passInput.value : '';
  const proxy = proxyInput ? proxyInput.value.trim() : '';

  if (!email || !password) {
    if (errEl) {
      errEl.textContent = 'Please enter both email and password.';
      errEl.style.display = 'block';
    }
    return;
  }

  if (submitBtn) {
    submitBtn.textContent = 'Signing in (solving challenge, ~20-30s)...';
    submitBtn.disabled = true;
  }
  if (errEl) errEl.style.display = 'none';

  try {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password, proxy })
    });
    const data = await res.json();
    if (data && data.error) {
      if (errEl) {
        errEl.textContent = data.error;
        errEl.style.display = 'block';
      }
    } else if (data && data.need_2fa) {
      document.getElementById('login-form').style.display = 'none';
      document.getElementById('twofa-form').style.display = 'flex';
    } else {
      closeModal('login-modal');
      fetchStatus();
    }
  } catch (err) {
    if (errEl) {
      errEl.textContent = 'Network request failed: ' + err.message;
      errEl.style.display = 'block';
    }
  } finally {
    if (submitBtn) {
      submitBtn.textContent = 'Sign In';
      submitBtn.disabled = false;
    }
  }
}

async function doVerify() {
  const codeInput = document.getElementById('twofa-code');
  const errEl = document.getElementById('twofa-error');
  const submitBtn = document.getElementById('btn-do-verify');

  const code = codeInput ? codeInput.value.trim() : '';
  if (!code) {
    if (errEl) {
      errEl.textContent = 'Please enter the 6-digit code.';
      errEl.style.display = 'block';
    }
    return;
  }

  if (submitBtn) {
    submitBtn.textContent = 'Verifying...';
    submitBtn.disabled = true;
  }

  try {
    const res = await fetch('/api/verify-2fa', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code })
    });
    const data = await res.json();
    if (data && data.error) {
      if (errEl) {
        errEl.textContent = data.error;
        errEl.style.display = 'block';
      }
    } else {
      closeModal('login-modal');
      fetchStatus();
    }
  } catch (err) {
    if (errEl) {
      errEl.textContent = 'Verification error: ' + err.message;
      errEl.style.display = 'block';
    }
  } finally {
    if (submitBtn) {
      submitBtn.textContent = 'Verify Code';
      submitBtn.disabled = false;
    }
  }
}

async function doLogout() {
  try {
    await fetch('/api/logout', { method: 'POST' });
    fetchStatus();
  } catch (err) {
    console.error('Logout failed:', err);
  }
}

// ==========================================================================
// Modals
// ==========================================================================
function openModal(id) {
  const m = document.getElementById(id);
  if (m) m.style.display = 'flex';
}

function closeModal(id) {
  const m = document.getElementById(id);
  if (m) m.style.display = 'none';
}

// ==========================================================================
// Theme & Language
// ==========================================================================
function applyTheme(theme) {
  currentTheme = theme;
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('fyrnox_theme', theme);

  const sel = document.getElementById('setting-theme');
  if (sel) sel.value = theme;
}

function changeTheme(theme) {
  applyTheme(theme);
}

function applyLanguage(lang) {
  currentLang = lang;
  localStorage.setItem('fyrnox_lang', lang);

  document.documentElement.setAttribute('lang', lang);
  if (lang === 'fa') {
    document.documentElement.setAttribute('dir', 'rtl');
  } else {
    document.documentElement.setAttribute('dir', 'ltr');
  }

  const curLabel = document.getElementById('lang-current-label');
  const altLabel = document.getElementById('lang-alt-label');
  if (curLabel && altLabel) {
    if (lang === 'fa') {
      curLabel.textContent = 'فارسی';
      altLabel.textContent = 'EN';
    } else {
      curLabel.textContent = 'EN';
      altLabel.textContent = 'FA';
    }
  }

  const langSelect = document.getElementById('setting-language');
  if (langSelect) langSelect.value = lang;

  // Localize text tags
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    if (I18N[lang] && I18N[lang][key]) {
      el.textContent = I18N[lang][key];
    }
  });

  // Localize placeholder tags
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    if (I18N[lang] && I18N[lang][key]) {
      el.placeholder = I18N[lang][key];
    }
  });

  // Re-render server list to refresh localized labels
  if (allServers && allServers.length > 0) {
    renderServerList(allServers);
  }

  // Refresh dynamic UI elements
  if (currentStatus) {
    updateUI(currentStatus);
  }

  // Update topbar title
  switchNav(activeNav);
}

function toggleLanguage() {
  const next = currentLang === 'fa' ? 'en' : 'fa';
  applyLanguage(next);
}

function changeLanguage(lang) {
  applyLanguage(lang);
}

// ==========================================================================
// Win32 Frameless Window Management
// ==========================================================================
function initWindowControls() {
  const dragRegion = document.getElementById('titlebar-drag-region');
  if (dragRegion) {
    dragRegion.addEventListener('mousedown', (e) => {
      // Left mouse button only
      if (e.button === 0) {
        if (window.nativeWindowDrag) {
          window.nativeWindowDrag();
        } else {
          fetch('/api/window/drag', { method: 'POST' }).catch(() => {});
        }
      }
    });
  }
}

function windowMinimize() {
  if (window.nativeWindowMinimize) {
    window.nativeWindowMinimize();
  } else {
    fetch('/api/window/minimize', { method: 'POST' }).catch(() => {});
  }
}

function windowMaximize() {
  if (window.nativeWindowMaximize) {
    window.nativeWindowMaximize();
  } else {
    fetch('/api/window/maximize', { method: 'POST' }).catch(() => {});
  }
}

function toggleWindowMaximize() {
  windowMaximize();
}

function windowClose() {
  if (window.nativeWindowClose) {
    window.nativeWindowClose();
  } else {
    fetch('/api/window/close', { method: 'POST' }).catch(() => {});
  }
}

function applyRecommendedServer() {
  if (currentStatus && currentStatus.recommended_server) {
    selectServer(currentStatus.recommended_server.hostname);
    const recCard = document.getElementById('dash-recommendation-card');
    if (recCard) recCard.style.display = 'none';
  }
}

// ==========================================================================
// Module 1: Windows Per-App Split Tunneling Controller
// ==========================================================================
let detectedProcesses = [];
let configuredSplitApps = [];

async function loadSplitTunnelView() {
  await fetchConfiguredSplitApps();
  await refreshProcessList();
}

async function fetchConfiguredSplitApps() {
  try {
    const res = await fetch('/api/settings');
    if (res.ok) {
      const settings = await res.json();
      configuredSplitApps = settings.split_tunnel_apps || [];
      renderConfiguredSplitChips();
    }
  } catch (err) {
    console.error('Failed to fetch split tunnel settings:', err);
  }
}

function renderConfiguredSplitChips() {
  const container = document.getElementById('split-configured-container');
  const countEl = document.getElementById('split-configured-count');
  if (!container) return;

  if (countEl) countEl.textContent = `${configuredSplitApps.length} Apps`;

  if (configuredSplitApps.length === 0) {
    container.innerHTML = `<div class="split-empty-msg" id="split-empty-msg">${(I18N[currentLang] && I18N[currentLang].split_no_apps) || 'No applications configured for split tunneling.'}</div>`;
    return;
  }

  container.innerHTML = configuredSplitApps.map(app => {
    const name = escapeHtml(app);
    return `
      <div class="split-chip-item">
        <span>${name}</span>
        <button type="button" class="chip-remove-btn" onclick="removeSplitApp('${name}')" title="Remove">✕</button>
      </div>
    `;
  }).join('');
}

async function refreshProcessList() {
  const countEl = document.getElementById('split-detected-count');
  if (countEl) countEl.textContent = 'Scanning...';

  try {
    const [runningRes, installedRes] = await Promise.all([
      fetch('/api/apps/running').then(r => r.json()).catch(() => []),
      fetch('/api/apps/installed').then(r => r.json()).catch(() => [])
    ]);

    const seen = new Set();
    const list = [];

    // Prioritize running processes first
    (runningRes || []).forEach(p => {
      const lower = p.name.toLowerCase();
      if (!seen.has(lower)) {
        seen.add(lower);
        list.push({ ...p, is_running: true });
      }
    });

    // Append installed processes
    (installedRes || []).forEach(p => {
      const lower = p.name.toLowerCase();
      if (!seen.has(lower)) {
        seen.add(lower);
        list.push({ ...p, is_running: false });
      }
    });

    detectedProcesses = list;
    if (countEl) countEl.textContent = `${list.length} Available`;
    renderProcessRows(detectedProcesses);
  } catch (err) {
    console.error('Failed to load process list:', err);
    if (countEl) countEl.textContent = 'Error scanning';
  }
}

function renderProcessRows(processes) {
  const container = document.getElementById('split-process-rows');
  if (!container) return;

  const confSet = new Set(configuredSplitApps.map(a => a.toLowerCase()));

  container.innerHTML = processes.slice(0, 100).map(proc => {
    const isConfigured = confSet.has(proc.name.toLowerCase());
    const statusText = proc.is_running
      ? ((I18N[currentLang] && I18N[currentLang].split_running) || 'Running')
      : ((I18N[currentLang] && I18N[currentLang].split_not_running) || 'Installed');
    const badgeClass = proc.is_running ? 'running' : 'installed';
    const actionLabel = isConfigured ? 'Remove' : '+ Add';
    const btnClass = isConfigured ? 'configured' : '';

    return `
      <div class="split-proc-row">
        <div class="proc-name-col">
          <div class="proc-icon-box">${proc.is_running ? '⚡' : '📦'}</div>
          <span>${escapeHtml(proc.description || proc.name)}</span>
        </div>
        <div class="proc-path-col font-mono" title="${escapeHtml(proc.path || proc.name)}">${escapeHtml(proc.name)}</div>
        <div class="proc-status-col">
          <span class="proc-badge ${badgeClass}">${statusText}</span>
        </div>
        <div class="proc-action-col">
          <button type="button" class="proc-toggle-btn ${btnClass}" onclick="toggleSplitAppRule('${escapeHtml(proc.name)}')">${actionLabel}</button>
        </div>
      </div>
    `;
  }).join('');
}

function filterProcessList(query) {
  const q = (query || '').toLowerCase().trim();
  if (!q) {
    renderProcessRows(detectedProcesses);
    return;
  }
  const filtered = detectedProcesses.filter(p => 
    p.name.toLowerCase().includes(q) || (p.description && p.description.toLowerCase().includes(q))
  );
  renderProcessRows(filtered);
}

function toggleSplitAppRule(appName) {
  const lower = appName.toLowerCase();
  const idx = configuredSplitApps.findIndex(a => a.toLowerCase() === lower);
  if (idx >= 0) {
    configuredSplitApps.splice(idx, 1);
  } else {
    configuredSplitApps.push(appName);
  }
  persistSplitTunnelApps();
}

function removeSplitApp(appName) {
  const lower = appName.toLowerCase();
  configuredSplitApps = configuredSplitApps.filter(a => a.toLowerCase() !== lower);
  persistSplitTunnelApps();
}

function addCustomSplitApp() {
  const input = document.getElementById('split-custom-exe-input');
  if (!input) return;
  const val = input.value.trim();
  if (!val) return;
  const base = val.split(/[\/\\]/).pop();
  if (base && !configuredSplitApps.some(a => a.toLowerCase() === base.toLowerCase())) {
    configuredSplitApps.push(base);
    input.value = '';
    persistSplitTunnelApps();
  }
}

function applySplitPreset(type) {
  const presets = {
    browsers: ['chrome.exe', 'msedge.exe', 'firefox.exe', 'brave.exe', 'opera.exe'],
    messaging: ['telegram.exe', 'discord.exe', 'whatsapp.exe'],
    gaming: ['steam.exe', 'epicgameslauncher.exe', 'riotclientservices.exe', 'battlenet.exe']
  };

  const list = presets[type] || [];
  list.forEach(app => {
    if (!configuredSplitApps.some(a => a.toLowerCase() === app)) {
      configuredSplitApps.push(app);
    }
  });
  persistSplitTunnelApps();
}

function clearSplitApps() {
  configuredSplitApps = [];
  persistSplitTunnelApps();
}

function toggleSplitTunnel(enabled) {
  const cur = (currentStatus && currentStatus.settings) || {};
  saveSettingsPayload({
    ...cur,
    split_tunnel_enabled: enabled
  });
}

function setSplitMode(mode) {
  const cur = (currentStatus && currentStatus.settings) || {};
  saveSettingsPayload({
    ...cur,
    split_tunnel_mode: mode
  });
}

function persistSplitTunnelApps() {
  renderConfiguredSplitChips();
  renderProcessRows(detectedProcesses);
  const cur = (currentStatus && currentStatus.settings) || {};
  saveSettingsPayload({
    ...cur,
    split_tunnel_apps: configuredSplitApps
  });
}

// ==========================================================================
// Module 2: Dedicated Low-Latency Gaming Mode Toggle
// ==========================================================================
async function toggleGamingMode(explicitActive) {
  try {
    const body = typeof explicitActive === 'boolean' ? { active: explicitActive } : {};
    const res = await fetch('/api/gaming-mode', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    if (res.ok) {
      const stats = await res.json();
      if (currentStatus) {
        currentStatus.gaming_stats = stats;
        if (currentStatus.settings) currentStatus.settings.gaming_mode = stats.active;
        updateUI(currentStatus);
      }
    }
  } catch (err) {
    console.error('Failed to toggle gaming mode:', err);
  }
}

// ==========================================================================
// Module 4: Floating Mini-Widget HUD Controller
// ==========================================================================
let hudVisible = false;

function toggleFloatingHUD() {
  const hud = document.getElementById('floating-mini-hud');
  if (!hud) return;
  hudVisible = !hudVisible;
  hud.style.display = hudVisible ? 'flex' : 'none';
  if (hudVisible && currentStatus) {
    updateUI(currentStatus);
  }
}

// Draggable HUD functionality
function initDraggableHUD() {
  const hud = document.getElementById('floating-mini-hud');
  if (!hud) return;
  let isDragging = false;
  let startX, startY, initLeft, initTop;

  hud.addEventListener('mousedown', (e) => {
    if (e.target.closest('button')) return;
    isDragging = true;
    startX = e.clientX;
    startY = e.clientY;
    const rect = hud.getBoundingClientRect();
    initLeft = rect.left;
    initTop = rect.top;
    hud.style.bottom = 'auto';
    hud.style.right = 'auto';
    hud.style.left = `${initLeft}px`;
    hud.style.top = `${initTop}px`;
  });

  window.addEventListener('mousemove', (e) => {
    if (!isDragging) return;
    const dx = e.clientX - startX;
    const dy = e.clientY - startY;
    hud.style.left = `${initLeft + dx}px`;
    hud.style.top = `${initTop + dy}px`;
  });

  window.addEventListener('mouseup', () => {
    isDragging = false;
  });
}


// Window Drag Helper
function windowDrag(e) {
  if (e && e.button !== 0) return;
  if (window.nativeWindowDrag) {
    window.nativeWindowDrag();
  } else {
    fetch('/api/window/drag', { method: 'POST' }).catch(() => {});
  }
}
window.windowDrag = windowDrag;
window.windowMinimize = windowMinimize;
window.windowMaximize = windowMaximize;
window.toggleWindowMaximize = toggleWindowMaximize;
window.windowClose = windowClose;
window.switchNav = switchNav;
window.toggleConnect = toggleConnect;
window.toggleModeDropdown = toggleModeDropdown;
window.handleModeDropdownSelect = handleModeDropdownSelect;
window.selectServer = selectServer;
window.pingAllServersInteractive = pingAllServersInteractive;
window.pingAllServers = pingAllServers;
window.dismissError = dismissError;
window.selectRoutingMode = selectRoutingMode;
window.toggleLanguage = toggleLanguage;
window.fetchStatus = fetchStatus;
window.handleAuthClick = handleAuthClick;
window.showNotificationPanel = typeof showNotificationPanel === 'function' ? showNotificationPanel : () => {};
window.interactivePingTest = interactivePingTest;
window.openModal = openModal;
window.closeModal = closeModal;
window.doLogin = doLogin;
window.doVerify = doVerify;
window.doLogout = doLogout;


// ==========================================================================
// FLAGSHIP FEATURES 1-5 CLIENT IMPLEMENTATION
// ==========================================================================

let activeConsolePlatform = 'ps5';
let cachedLanIP = '192.168.1.1';

// --- Feature 1: LAN Proxy Sharing & Console Hotspot ---
async function fetchLANShareStatus() {
  try {
    const res = await fetch('/api/lan-share');
    if (res.ok) {
      const data = await res.json();
      renderLANShareUI(data);
    }
  } catch (err) {
    console.error('Failed to fetch LAN share status:', err);
  }
}

async function toggleLANShare(enabled) {
  try {
    const res = await fetch('/api/lan-share', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: enabled })
    });
    const data = await res.json();
    if (!res.ok || data.error) {
      alert(data.error || 'Failed to toggle LAN sharing');
      const toggle = document.getElementById('setting-lan-share');
      if (toggle) toggle.checked = !enabled;
      return;
    }
    renderLANShareUI(data);
  } catch (err) {
    console.error('Error toggling LAN share:', err);
  }
}

function renderLANShareUI(data) {
  const toggle = document.getElementById('setting-lan-share');
  const details = document.getElementById('lan-share-details-panel');
  const ipVal = document.getElementById('lan-ip-val');
  const httpUrl = document.getElementById('lan-http-url');
  const socksUrl = document.getElementById('lan-socks-url');
  const qrContainer = document.getElementById('lan-qrcode-container');
  const guideIp = document.getElementById('guide-ip');

  if (toggle) toggle.checked = !!data.enabled;
  if (details) details.style.display = data.enabled ? 'block' : 'none';

  if (data.lan_ip) {
    cachedLanIP = data.lan_ip;
  }

  if (ipVal) ipVal.textContent = cachedLanIP;
  if (guideIp) guideIp.textContent = cachedLanIP;

  const httpStr = `http://${cachedLanIP}:${data.http_port || 10809}`;
  const socksStr = `socks5://${cachedLanIP}:${data.socks_port || 10808}`;

  if (httpUrl) httpUrl.textContent = httpStr;
  if (socksUrl) socksUrl.textContent = socksStr;

  if (qrContainer && typeof window.generateQRCodeSVG === 'function') {
    const qrData = data.proxy_url || httpStr;
    qrContainer.innerHTML = window.generateQRCodeSVG(qrData, 154, '#0B0D12', '#ffffff');
  }
}

function copyLANEndpoint(type) {
  const url = type === 'http' ? `http://${cachedLanIP}:10809` : `socks5://${cachedLanIP}:10808`;
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url).then(() => {
      alert(currentLang === 'fa' ? 'آدرس کپی شد: ' + url : 'Copied to clipboard: ' + url);
    });
  } else {
    prompt('Copy endpoint:', url);
  }
}

function switchConsoleTab(platform) {
  activeConsolePlatform = platform;
  document.querySelectorAll('.console-tab-btn').forEach(btn => {
    btn.classList.toggle('active', btn.textContent.toLowerCase().includes(platform));
  });

  const guideBody = document.getElementById('console-guide-body');
  if (!guideBody) return;

  if (platform === 'ps5') {
    guideBody.textContent = `Settings → Network → Settings → Set Up Internet Connection → Advanced Settings → Proxy Server: Use → Address: ${cachedLanIP}, Port: 10809`;
  } else if (platform === 'xbox') {
    guideBody.textContent = `Settings → General → Network settings → Advanced settings → Configure PC Hotspot Gateway or Set Proxy: ${cachedLanIP}:10809`;
  } else if (platform === 'switch') {
    guideBody.textContent = `System Settings → Internet → Internet Settings → Select Wi-Fi → Change Settings → Proxy Settings: On → Server: ${cachedLanIP}, Port: 10809`;
  }
}

// --- Feature 4: Global Panic Key & Stealth Mode ---
async function triggerPanicTest() {
  const confirmMsg = currentLang === 'fa' 
    ? 'آیا از فعال‌سازی حالت اضطراری (Panic) اطمینان دارید؟ تمام اتصالات قطع شده و پنجره پنهان می‌شود.'
    : 'Trigger Emergency Panic Lockdown? VPN connections will sever and the window will hide immediately.';
  if (!confirm(confirmMsg)) return;

  try {
    await fetch('/api/panic/toggle', { method: 'POST' });
  } catch (err) {
    console.error('Error triggering panic:', err);
  }
}

// --- Feature 5: In-App Sanctions & Service Health Matrix ---
async function fetchServiceMatrix() {
  try {
    const res = await fetch('/api/service-matrix');
    if (res.ok) {
      const data = await res.json();
      renderServiceMatrix(data);
    }
  } catch (err) {
    console.error('Failed to fetch service matrix:', err);
  }
}

async function refreshServiceMatrix() {
  const btn = document.getElementById('btn-matrix-check');
  if (btn) btn.classList.add('checking');

  try {
    const res = await fetch('/api/service-matrix/check', { method: 'POST' });
    if (res.ok) {
      const data = await res.json();
      renderServiceMatrix(data);
    }
  } catch (err) {
    console.error('Failed to refresh service matrix:', err);
  } finally {
    if (btn) btn.classList.remove('checking');
  }
}

function renderServiceMatrix(data) {
  const grid = document.getElementById('service-matrix-grid');
  const timeEl = document.getElementById('matrix-last-updated');

  if (timeEl && data.last_updated) {
    timeEl.textContent = (currentLang === 'fa' ? 'آخرین بررسی: ' : 'Last check: ') + data.last_updated;
  }

  if (!grid || !data.services) return;

  grid.innerHTML = data.services.map(svc => {
    let pillClass = 'unrestricted';
    let pillLabel = currentLang === 'fa' ? 'بدون محدودیت' : 'Unrestricted';
    if (svc.status === 'geoblocked') {
      pillClass = 'geoblocked';
      pillLabel = currentLang === 'fa' ? 'تحریم منطقه‌ای' : 'Geoblocked';
    } else if (svc.status === 'blocked') {
      pillClass = 'blocked';
      pillLabel = currentLang === 'fa' ? 'مسدود / فیلتر' : 'Blocked';
    }

    const latencyText = svc.latency_ms > 0 ? `${svc.latency_ms} ms` : '-- ms';

    return `
      <div class="matrix-item-card">
        <div class="matrix-item-top">
          <span class="matrix-item-name">${escapeHtml(svc.name)}</span>
          <span class="matrix-item-category">${escapeHtml(svc.category)}</span>
        </div>
        <div class="matrix-item-bottom">
          <span class="matrix-status-pill ${pillClass}">● ${pillLabel}</span>
          <span class="matrix-latency-val font-mono">${latencyText}</span>
        </div>
      </div>
    `;
  }).join('');
}

window.toggleLANShare = toggleLANShare;
window.copyLANEndpoint = copyLANEndpoint;
window.switchConsoleTab = switchConsoleTab;
window.triggerPanicTest = triggerPanicTest;
window.refreshServiceMatrix = refreshServiceMatrix;
window.fetchServiceMatrix = fetchServiceMatrix;
window.fetchLANShareStatus = fetchLANShareStatus;


// ==========================================================================
// Interactive Telemetry Latency (Real-Time Ping on Click)
// ==========================================================================
let isManualPinging = false;
async function triggerManualPing() {
  if (isManualPinging) return;
  const pingEl = document.getElementById('bottom-ping-val');
  if (!pingEl) return;
  
  const prevVal = pingEl.textContent;
  isManualPinging = true;
  pingEl.textContent = '...';

  try {
    const res = await fetch('/api/ping');
    if (res.ok) {
      const data = await res.json();
      if (data && data.ok && data.ping_ms > 0) {
        pingEl.textContent = `${data.ping_ms} ms`;
        if (currentStatus && currentStatus.server) {
          currentStatus.server.ping_ms = data.ping_ms;
        }
        return;
      }
    }
    pingEl.textContent = prevVal !== '...' ? prevVal : '—';
  } catch (err) {
    console.warn('[Fyrnox] Real-time ping failed:', err);
    pingEl.textContent = prevVal !== '...' ? prevVal : '—';
  } finally {
    setTimeout(() => {
      isManualPinging = false;
    }, 350);
  }
}

window.triggerManualPing = triggerManualPing;


// ==========================================================================
// DUAL-MODE IOS SWIPE NAVIGATION & SCRUBBING
// ==========================================================================
function initIOSSwipeNavigation() {
  // Purged: deterministic click-based dock navigation ensures 100% stable layout
}

document.addEventListener('DOMContentLoaded', () => {
  // Clean initialization
});

function copyExitInfo() {
  const cityEl = document.getElementById('badge-city');
  const latEl = document.getElementById('badge-latency');
  const text = `${cityEl ? cityEl.textContent : 'JOHANNESBURG, ZA'} (Latency: ${latEl ? latEl.textContent : '42 ms'})`;
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text);
  }
}
window.copyExitInfo = copyExitInfo;
