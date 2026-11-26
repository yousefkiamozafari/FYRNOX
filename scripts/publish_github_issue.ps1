<#
.SYNOPSIS
    Automated script to publish professional bilingual GitHub Issues for FYRNOX.
#>
param(
    [Parameter(Mandatory=$true)]
    [string]$Title,

    [Parameter(Mandatory=$true)]
    [string]$BodyMarkdown,

    [string[]]$Labels = @("enhancement", "documentation"),

    [string]$Socks5hProxy = "127.0.0.1:10808"
)

# 1. Retrieve Git Credential Manager Token
$gcmPath = "C:\Users\Yousef Kia\AppData\Local\Microsoft\WinGet\Packages\Git.MinGit_Microsoft.Winget.Source_8wekyb3d8bbwe\ucrt64\bin\git-credential-manager.exe"
if (-not (Test-Path $gcmPath)) {
    $gcmPath = (Get-Command "git-credential-manager" -ErrorAction SilentlyContinue).Source
}

$token = $null
if ($gcmPath) {
    $inputStr = "protocol=https`nhost=github.com`n`n"
    $pinfo = New-Object System.Diagnostics.ProcessStartInfo
    $pinfo.FileName = $gcmPath
    $pinfo.Arguments = "get"
    $pinfo.RedirectStandardInput = $true
    $pinfo.RedirectStandardOutput = $true
    $pinfo.UseShellExecute = $false
    $p = New-Object System.Diagnostics.Process
    $p.StartInfo = $pinfo
    $p.Start() | Out-Null
    $p.StandardInput.Write($inputStr)
    $p.StandardInput.Flush()
    $p.StandardInput.Close()
    $stdout = $p.StandardOutput.ReadToEnd()
    $p.WaitForExit(5000)

    if ($stdout -match "password=(.+)") {
        $token = $matches[1].Trim()
    }
}

if (-not $token) {
    Write-Error "Failed to retrieve GitHub credential token from Git Credential Manager."
    exit 1
}

$repo = "yousefkiamozafari/FYRNOX"
$payloadObj = @{
    title  = $Title
    body   = $BodyMarkdown
    labels = $Labels
}

# Write payload as UTF-8 without BOM so Persian & Markdown bytes are transmitted verbatim.
$tmpPayload = Join-Path ([System.IO.Path]::GetTempPath()) ("fyrnox-issue-" + [System.Guid]::NewGuid().ToString("N") + ".json")
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$jsonText = $payloadObj | ConvertTo-Json -Depth 5
[System.IO.File]::WriteAllText($tmpPayload, $jsonText, $utf8NoBom)

try {
    $useProxy = $false
    if ($Socks5hProxy) {
        $parts = $Socks5hProxy -split ":"
        $proxyHost = $parts[0]
        $proxyPort = if ($parts.Count -ge 2) { [int]$parts[1] } else { 10808 }
        try {
            $tcp = New-Object System.Net.Sockets.TcpClient
            $iar = $tcp.BeginConnect($proxyHost, $proxyPort, $null, $null)
            if ($iar.AsyncWaitHandle.WaitOne(500, $false) -and $tcp.Connected) {
                $useProxy = $true
            }
            $tcp.Close()
        } catch {
            $useProxy = $false
        }
    }

    $curlArgs = @(
        "-sS",
        "--fail-with-body",
        "--max-time", "30",
        "-X", "POST",
        "-H", "Authorization: Bearer $token",
        "-H", "User-Agent: FYRNOX-Automation-Agent",
        "-H", "Accept: application/vnd.github.v3+json",
        "-H", "Content-Type: application/json; charset=utf-8",
        "--data-binary", "@$tmpPayload"
    )
    if ($useProxy) {
        $curlArgs = @("--socks5-hostname", $Socks5hProxy) + $curlArgs
    }
    $curlArgs += "https://api.github.com/repos/$repo/issues"

    $rawOut = & curl.exe @curlArgs 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to publish GitHub Issue via curl.exe (exit $LASTEXITCODE): $rawOut"
        exit 1
    }

    $response = ($rawOut -join "`n") | ConvertFrom-Json
    Write-Host "SUCCESS: Published Issue #$($response.number) - $($response.html_url)"
    return $response
} finally {
    Remove-Item -Path $tmpPayload -Force -ErrorAction SilentlyContinue
}
