$ErrorActionPreference = "Continue"
$repoRoot = "C:\Users\LouiNoor\.gemini\antigravity\scratch\FoxyVPN-Windows"
Set-Location $repoRoot

Write-Host "=========================================================="
Write-Host " FYRNOX LIVE REGRESSION & SYSTEM PROXY VERIFICATION TEST  "
Write-Host "=========================================================="

# 1. Baseline Direct IP
Write-Host "`n[STEP 1] Measuring Baseline Direct Public IP..."
$directIp = ""
try {
    $res = Invoke-RestMethod -Uri "https://api.ipify.org?format=json" -TimeoutSec 8
    $directIp = $res.ip
    Write-Host "  Direct Baseline IP: $directIp"
} catch {
    Write-Host "  Error checking direct IP: $_"
}

# 2. Check initial proxy state
Write-Host "`n[STEP 2] Inspecting Initial Windows Registry Proxy State..."
$regKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings"
$initEnable = (Get-ItemProperty -Path $regKey -Name ProxyEnable -ErrorAction SilentlyContinue).ProxyEnable
$initServer = (Get-ItemProperty -Path $regKey -Name ProxyServer -ErrorAction SilentlyContinue).ProxyServer
Write-Host "  Initial ProxyEnable: $initEnable"
Write-Host "  Initial ProxyServer: $initServer"

# 3. Launch Fyrnox in Headless Mode on Port 19180
Write-Host "`n[STEP 3] Launching Fyrnox.exe -headless -port 19180..."
$proc = Start-Process -FilePath "$repoRoot\Fyrnox.exe" -ArgumentList "-headless", "-port", "19180" -PassThru
Start-Sleep -Seconds 4

try {
    # 4. Check API status (Default Mode Verification)
    Write-Host "`n[STEP 4] Verifying Initial API Status & Default Mode..."
    $status = Invoke-RestMethod -Uri "http://127.0.0.1:19180/api/status" -TimeoutSec 5
    Write-Host "  Engine State: $($status.state)"
    Write-Host "  EnableTunMode: $($status.enable_tun_mode) (Expected: False)"
    Write-Host "  UseSystemProxy: $($status.use_system_proxy) (Expected: True)"
    Write-Host "  User Email: $($status.user_email)"

    # 5. Connect via System Proxy
    Write-Host "`n[STEP 5] Sending POST /api/connect..."
    $connRes = Invoke-RestMethod -Uri "http://127.0.0.1:19180/api/connect" -Method Post -TimeoutSec 10
    Write-Host "  Connect triggered. Current state: $($connRes.status.state)"

    # Wait for connection to establish (up to 30s)
    $connected = $false
    for ($i = 0; $i -lt 30; $i++) {
        Start-Sleep -Seconds 1
        $currStatus = Invoke-RestMethod -Uri "http://127.0.0.1:19180/api/status" -TimeoutSec 3
        Write-Host "  [$i s] State: $($currStatus.state) - Step: $($currStatus.connect_step)"
        if ($currStatus.state -eq "connected") {
            $connected = $true
            break
        }
        if ($currStatus.state -eq "error") {
            Write-Host "  ERROR reported by engine: $($currStatus.error_message)"
            break
        }
    }

    if (-not $connected) {
        Write-Host "  FAILED: VPN did not reach 'connected' state!"
    } else {
        Write-Host "  SUCCESS: Reached connected state!"

        # 6. Verify Windows Proxy Registry & WinHTTP
        Write-Host "`n[STEP 6] Verifying Windows System Proxy Configuration..."
        $activeEnable = (Get-ItemProperty -Path $regKey -Name ProxyEnable -ErrorAction SilentlyContinue).ProxyEnable
        $activeServer = (Get-ItemProperty -Path $regKey -Name ProxyServer -ErrorAction SilentlyContinue).ProxyServer
        Write-Host "  Active ProxyEnable: $activeEnable (Expected: 1)"
        Write-Host "  Active ProxyServer: $activeServer"

        $winhttp = netsh winhttp show proxy
        Write-Host "  netsh winhttp output:"
        $winhttp | ForEach-Object { Write-Host "    $_" }

        # 7. Test Traffic Through HTTP Proxy (127.0.0.1:20809)
        Write-Host "`n[STEP 7] Performing External IP Request Through Proxy (127.0.0.1:20809)..."
        $proxyIpOutput = & curl.exe -s --max-time 15 -x http://127.0.0.1:20809 https://api.ipify.org?format=json
        Write-Host "  Proxy IP response: $proxyIpOutput"

        # Also test HTTP CONNECT to httpbin
        $httpbinOutput = & curl.exe -s --max-time 15 -x http://127.0.0.1:20809 http://httpbin.org/ip
        Write-Host "  HTTPBin response: $httpbinOutput"

        # Compare observed IP
        Write-Host "`n[STEP 8] IP Comparison Analysis:"
        Write-Host "  Direct Baseline IP: $directIp"
        Write-Host "  Observed Tunnel IP: $proxyIpOutput"
        if ($proxyIpOutput -match '"ip":\s*"([^"]+)"') {
            $observedIp = $matches[1]
            if ($observedIp -ne $directIp) {
                Write-Host "  PASS: Public IP successfully changed from $directIp to $observedIp!"
            } else {
                Write-Host "  WARNING: Observed IP is identical to baseline."
            }
        }

        # 8. Test Ping All Servers API
        Write-Host "`n[STEP 9] Testing Server Ping Measurement (/api/ping-all)..."
        $pingRes = Invoke-RestMethod -Uri "http://127.0.0.1:19180/api/ping-all" -Method Post -TimeoutSec 15
        Write-Host "  Measured servers count: $($pingRes.Count)"
        $responsive = $pingRes | Where-Object { $_.ping_ms -gt 0 }
        Write-Host "  Responsive servers: $($responsive.Count) / $($pingRes.Count)"
        if ($responsive.Count -gt 0) {
            $top3 = $responsive | Select-Object -First 3
            foreach ($s in $top3) {
                Write-Host "    $($s.country_name) ($($s.city_name)) - $($s.hostname): $($s.ping_ms) ms"
            }
        }
    }

    # 9. Disconnect
    Write-Host "`n[STEP 10] Sending POST /api/disconnect..."
    $discRes = Invoke-RestMethod -Uri "http://127.0.0.1:19180/api/disconnect" -Method Post -TimeoutSec 5
    Write-Host "  Disconnect triggered. State: $($discRes.status.state)"
    Start-Sleep -Seconds 1

    # 10. Verify Clean Disconnect and Proxy Restoration
    Write-Host "`n[STEP 11] Verifying System Proxy Clean Reset..."
    $postEnable = (Get-ItemProperty -Path $regKey -Name ProxyEnable -ErrorAction SilentlyContinue).ProxyEnable
    $postServer = (Get-ItemProperty -Path $regKey -Name ProxyServer -ErrorAction SilentlyContinue).ProxyServer
    Write-Host "  Post-Disconnect ProxyEnable: $postEnable (Expected: 0)"
    Write-Host "  Post-Disconnect ProxyServer: $postServer"

    $winhttpPost = netsh winhttp show proxy
    Write-Host "  netsh winhttp post-disconnect:"
    $winhttpPost | ForEach-Object { Write-Host "    $_" }

} finally {
    Write-Host "`n[STEP 12] Stopping Headless Test Instance..."
    if ($proc -and -not $proc.HasExited) {
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "`n=========================================================="
Write-Host " VERIFICATION TEST COMPLETE                               "
Write-Host "=========================================================="
