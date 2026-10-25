$ErrorActionPreference = "Stop"
$repoRoot = "C:\Users\LouiNoor\.gemini\antigravity\scratch\FoxyVPN-Windows"
Set-Location $repoRoot

$env:GOCACHE = "C:\Users\LouiNoor\.gemini\antigravity\scratch\gocache"
$env:GOPATH = "C:\Users\LouiNoor\.gemini\antigravity\scratch\gopath"
$env:GOOS = "windows"

Write-Host "=== 1. Building amd64 (x64) binaries ==="
$env:GOARCH = "amd64"
go build -ldflags "-s -w -H windowsgui" -o "$repoRoot\build\Fyrnox-x64.exe" ./cmd/fyrnox
go build -ldflags "-s -w -H windowsgui" -o "$repoRoot\build\Uninstall.exe" ./cmd/uninstaller

Write-Host "=== 2. Building 386 (x86) binaries ==="
$env:GOARCH = "386"
go build -ldflags "-s -w -H windowsgui" -o "$repoRoot\build\Fyrnox-x86.exe" ./cmd/fyrnox
go build -ldflags "-s -w -H windowsgui" -o "$repoRoot\build\Uninstall-x86.exe" ./cmd/uninstaller

Write-Host "=== 3. Updating root Fyrnox.exe with new amd64 binary ==="
Copy-Item -Force "$repoRoot\build\Fyrnox-x64.exe" "$repoRoot\Fyrnox.exe"

Write-Host "=== 4. Updating portable x64 directory ==="
New-Item -ItemType Directory -Force "$repoRoot\build\portable\Fyrnox-Portable" | Out-Null
Copy-Item -Force "$repoRoot\build\Fyrnox-x64.exe" "$repoRoot\build\portable\Fyrnox-Portable\Fyrnox.exe"
Copy-Item -Force "$repoRoot\build\Uninstall.exe" "$repoRoot\build\portable\Fyrnox-Portable\Uninstall.exe"
Copy-Item -Force "$repoRoot\icon.ico" "$repoRoot\build\portable\Fyrnox-Portable\icon.ico"
Copy-Item -Force "$repoRoot\README.md" "$repoRoot\build\portable\Fyrnox-Portable\README.md"
Copy-Item -Force "$repoRoot\build\wintun.dll" "$repoRoot\build\portable\Fyrnox-Portable\wintun.dll"

Write-Host "=== 5. Updating portable x86 directory ==="
New-Item -ItemType Directory -Force "$repoRoot\build\portable-x86\Fyrnox-Portable" | Out-Null
Copy-Item -Force "$repoRoot\build\Fyrnox-x86.exe" "$repoRoot\build\portable-x86\Fyrnox-Portable\Fyrnox.exe"
Copy-Item -Force "$repoRoot\build\Uninstall-x86.exe" "$repoRoot\build\portable-x86\Fyrnox-Portable\Uninstall.exe"
Copy-Item -Force "$repoRoot\icon.ico" "$repoRoot\build\portable-x86\Fyrnox-Portable\icon.ico"
Copy-Item -Force "$repoRoot\README.md" "$repoRoot\build\portable-x86\Fyrnox-Portable\README.md"

Write-Host "=== 6. Packaging Portable ZIPs ==="
$zipX64 = "$repoRoot\build\Fyrnox-v1.0.4-Windows-x64-Portable.zip"
if (Test-Path $zipX64) { Remove-Item $zipX64 }
Compress-Archive -Path "$repoRoot\build\portable\Fyrnox-Portable\*" -DestinationPath $zipX64 -CompressionLevel Optimal

$zipX86 = "$repoRoot\build\Fyrnox-v1.0.4-Windows-x86-Portable.zip"
if (Test-Path $zipX86) { Remove-Item $zipX86 }
Compress-Archive -Path "$repoRoot\build\portable-x86\Fyrnox-Portable\*" -DestinationPath $zipX86 -CompressionLevel Optimal

Write-Host "=== 7. Building x64 Installer ==="
Copy-Item -Force "$repoRoot\build\Fyrnox-x64.exe" "$repoRoot\cmd\installer\payload\Fyrnox.exe"
Copy-Item -Force "$repoRoot\build\Uninstall.exe" "$repoRoot\cmd\installer\payload\Uninstall.exe"
Copy-Item -Force "$repoRoot\icon.ico" "$repoRoot\cmd\installer\payload\icon.ico"
Copy-Item -Force "$repoRoot\README.md" "$repoRoot\cmd\installer\payload\README.md"
Copy-Item -Force "$repoRoot\build\wintun.dll" "$repoRoot\cmd\installer\payload\wintun.dll"

$env:GOARCH = "amd64"
New-Item -ItemType Directory -Force "$repoRoot\build\installer" | Out-Null
go build -ldflags "-s -w -H windowsgui" -o "$repoRoot\build\installer\Fyrnox-Setup-v1.0.4.exe" ./cmd/installer

Write-Host "=== 8. Cleaning and updating build/dist directory ==="
New-Item -ItemType Directory -Force "$repoRoot\build\dist" | Out-Null
Get-ChildItem "$repoRoot\build\dist" -File | Where-Object { $_.Name -match "v1\.0\.[0-3]" } | Remove-Item -Force

Copy-Item -Force "$repoRoot\build\installer\Fyrnox-Setup-v1.0.4.exe" "$repoRoot\build\dist\Fyrnox-Setup-v1.0.4.exe"
Copy-Item -Force "$repoRoot\build\installer\Fyrnox-Setup-v1.0.4.exe" "$repoRoot\build\dist\Fyrnox-Setup-Windows-11-10-x64.exe"
Copy-Item -Force $zipX64 "$repoRoot\build\dist\Fyrnox-v1.0.4-Windows-x64-Portable.zip"
Copy-Item -Force $zipX86 "$repoRoot\build\dist\Fyrnox-v1.0.4-Windows-x86-Portable.zip"

Write-Host "=== 9. Building x86 Installer for build/dist ==="
Copy-Item -Force "$repoRoot\build\Fyrnox-x86.exe" "$repoRoot\cmd\installer\payload\Fyrnox.exe"
Copy-Item -Force "$repoRoot\build\Uninstall-x86.exe" "$repoRoot\cmd\installer\payload\Uninstall.exe"
Copy-Item -Force "$repoRoot\build\portable-x86\Fyrnox-Portable\wintun.dll" "$repoRoot\cmd\installer\payload\wintun.dll"

$env:GOARCH = "386"
go build -ldflags "-s -w -H windowsgui" -o "$repoRoot\build\dist\Fyrnox-Setup-Windows-7-8-10-x86.exe" ./cmd/installer

# Restore x64 payload in cmd/installer/payload
Copy-Item -Force "$repoRoot\build\Fyrnox-x64.exe" "$repoRoot\cmd\installer\payload\Fyrnox.exe"
Copy-Item -Force "$repoRoot\build\Uninstall.exe" "$repoRoot\cmd\installer\payload\Uninstall.exe"
Copy-Item -Force "$repoRoot\build\wintun.dll" "$repoRoot\cmd\installer\payload\wintun.dll"

Write-Host "=== Build and Packaging Complete! ==="
