<#
.SYNOPSIS
    Automated script to publish professional bilingual GitHub Issues for FYRNOX.
#>
param(
    [Parameter(Mandatory=$true)]
    [string]$Title,

    [Parameter(Mandatory=$true)]
    [string]$BodyMarkdown,

    [string[]]$Labels = @("enhancement", "documentation")
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
$payload = @{
    title = $Title
    body = $BodyMarkdown
    labels = $Labels
}

$jsonBytes = [System.Text.Encoding]::UTF8.GetBytes(($payload | ConvertTo-Json -Depth 5))

$headers = @{
    "Authorization" = "Bearer $token"
    "User-Agent" = "FYRNOX-Automation-Agent"
    "Accept" = "application/vnd.github.v3+json"
}

try {
    $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/issues" -Method Post -Headers $headers -Body $jsonBytes -ContentType "application/json; charset=utf-8"
    Write-Host "SUCCESS: Published Issue #$($response.number) - $($response.html_url)"
    return $response
} catch {
    Write-Error "Failed to publish GitHub Issue: $_"
    exit 1
}
