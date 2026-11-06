# Privacy Policy & Data Handling Disclosure

**Last Updated:** November 2026  
**Application:** FYRNOX for Windows (v1.0.4+)  
**Maintainer:** Yousef Kia ([https://github.com/yousefkiamozafari](https://github.com/yousefkiamozafari))

---

## ⚖️ Third-Party Disclaimer
**FYRNOX is an independent open-source desktop client.**  
It is **not** an official product of Mozilla Corporation, Mozilla Foundation, or Fastly, Inc. FYRNOX is not sponsored, authorized, or endorsed by Mozilla. "Firefox" and "Mozilla" are registered trademarks of the Mozilla Foundation.

---

## 🔍 Verified Source-Code Privacy Audit

This document is based on a direct, line-by-line inspection of the FYRNOX Go codebase (`cmd/fyrnox/`, `tokens.go`, `fxa.go`, `guardian.go`, `fastly.go`).

### 1. Data Collection & Telemetry
* **Zero Telemetry:** FYRNOX contains **NO** analytics SDKs, user tracking, metrics collectors, Google Analytics, or diagnostic phone-home mechanisms.
* **No Developer Servers:** FYRNOX does not operate, report to, or communicate with any proprietary backend server owned or managed by the FYRNOX maintainer.

### 2. User Authentication & Credentials
* **Credential Handling (`fxa.go`):** When you sign in with your Firefox Account, your email and password are transmitted directly and exclusively over encrypted HTTPS to Mozilla's official authentication endpoint (`https://api.accounts.firefox.com/v1`).
* **Passwords Never Saved:** Your plaintext password is used solely in memory during the authentication exchange and is **never** written to disk or recorded in log outputs.
* **OAuth Token Storage (`tokens.go`):** OAuth tokens (`access_token`, `refresh_token`, expiration) returned by Mozilla are saved locally in your user profile directory:
  * Windows path: `%USERPROFILE%\.firefox-vpn-tokens.json` (e.g. `C:\Users\<Username>\.firefox-vpn-tokens.json`)
  * File permissions: Restricted via atomic temporary file writing (`0600` permissions equivalent).
  * Logout behavior: Clicking "Log Out" immediately deletes `%USERPROFILE%\.firefox-vpn-tokens.json` from the disk.

### 3. Local Configuration & Storage
* **Settings File (`settings.json`):** Your local preferences (routing mode, selected protocol, kill switch state, bypass rules) are saved locally in `settings.json` within the application execution directory. No sensitive credentials reside in `settings.json`.

### 4. Network Traffic & Tunneling
* **Upstream Edge Network (`fastly.go`):** When connected, your network packets are tunneled through Fastly's edge network using MASQUE protocols over HTTP/2 or HTTP/3 (QUIC).
* **Payload Privacy:** FYRNOX acts strictly as a transport relay between your local machine and the edge node. It does not inspect, parse, alter, or record your web traffic or visited URLs.
* **DNS Privacy:** All domain name resolution is routed either through remote SOCKS5 edge resolution or DNS-over-HTTPS (RFC 8484) to eliminate ISP DNS hijacking and DNS leaks.

### 5. External Network Requests
The application connects strictly to the following verified public endpoints:
| Endpoint | Purpose | Destination |
| :--- | :--- | :--- |
| `https://api.accounts.firefox.com` | User authentication & OAuth token exchange | Mozilla Corporation |
| `https://guardian.services.mozilla.com` | VPN Proxy Pass generation & quota status | Mozilla Corporation |
| `https://firefox.settings.services.mozilla.com` | Fetching the list of 35 global VPN server locations | Mozilla Corporation |
| `https://api.ipify.org?format=json` | Verification of public egress IP upon connection | Public IP Echo API |
| Fastly MASQUE Edge Nodes | Encrypted egress tunnel | Fastly, Inc. |

### 6. Logging
* **Memory & Terminal Only:** Diagnostic log messages are handled by internal log functions (`LogInfo`, `LogWarn`, `LogError`) and printed to stdout/stderr.
* **No Disk Logs:** Logs are not written to persistent disk files or transmitted over the network.
* **Token Scrubbing:** Log outputs do not contain passwords, bearer tokens, or full authorization headers.

---

## 🛡️ Summary of User Guarantees
1. No account data is shared with the FYRNOX developer.
2. No user activity is monitored or logged.
3. Logging out cleanly destroys saved tokens from your PC.
4. The application is completely open-source and auditable on GitHub.
