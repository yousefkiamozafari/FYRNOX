package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

//go:embed payload/Fyrnox.exe
var fyrnoxExe []byte

//go:embed payload/Uninstall.exe
var uninstallExe []byte

//go:embed payload/icon.ico
var iconIco []byte

//go:embed payload/README.md
var readmeMd []byte

//go:embed payload/wintun.dll
var wintunDll []byte

var (
	user32     = syscall.NewLazyDLL("user32.dll")
	procMsgBox = user32.NewProc("MessageBoxW")
)

const (
	mbOK           = 0x00000000
	mbYesNo        = 0x00000004
	mbIconQuestion = 0x00000020
	mbIconInfo     = 0x00000040
	idYes          = 6
)

func showMsg(title, msg string, uType uint) int {
	tPtr, _ := syscall.UTF16PtrFromString(title)
	mPtr, _ := syscall.UTF16PtrFromString(msg)
	ret, _, _ := procMsgBox.Call(0, uintptr(unsafe.Pointer(mPtr)), uintptr(unsafe.Pointer(tPtr)), uintptr(uType))
	return int(ret)
}

func createShortcut(targetExe, shortcutPath, iconPath, workDir string) error {
	script := fmt.Sprintf(`$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%s'); $s.TargetPath = '%s'; $s.WorkingDirectory = '%s'; $s.IconLocation = '%s'; $s.Save()`,
		shortcutPath, targetExe, workDir, iconPath)
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Run()
}

func main() {
	silent := flag.Bool("silent", false, "Run silent installation")
	flag.BoolVar(silent, "s", false, "Run silent installation")
	flag.Parse()

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, _ := os.UserHomeDir()
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	installDir := filepath.Join(localAppData, "Programs", "Fyrnox")

	if !*silent {
		welcomeText := fmt.Sprintf("Welcome to Fyrnox for Windows v1.0.4 Setup!\n"+
			"Fast, Subscription-Free Firefox VPN Client\n"+
			"Powered by Yousef Kia\n"+
			"GitHub: https://github.com/yousefkiamozafari\n\n"+
			"Installation Directory:\n%s\n\n"+
			"Would you like to continue with the installation?\n\n"+
			"به برنامه نصب Fyrnox خوش آمدید! آیا مایل به شروع نصب هستید؟", installDir)
		resp := showMsg("Fyrnox Setup v1.0.4", welcomeText, mbYesNo|mbIconQuestion)
		if resp != idYes {
			os.Exit(0)
		}
	}

	// 1. Terminate any running Fyrnox silently
	cmdKill := exec.Command("taskkill", "/F", "/IM", "Fyrnox.exe")
	cmdKill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	_ = cmdKill.Run()

	// 2. Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		if !*silent {
			showMsg("Setup Error", fmt.Sprintf("Failed to create installation directory: %v", err), mbOK)
		}
		os.Exit(1)
	}

	// 3. Write files
	targetExe := filepath.Join(installDir, "Fyrnox.exe")
	targetUninstaller := filepath.Join(installDir, "Uninstall.exe")
	targetIcon := filepath.Join(installDir, "icon.ico")
	targetReadme := filepath.Join(installDir, "README.md")
	targetWintun := filepath.Join(installDir, "wintun.dll")

	_ = os.WriteFile(targetExe, fyrnoxExe, 0755)
	_ = os.WriteFile(targetUninstaller, uninstallExe, 0755)
	_ = os.WriteFile(targetIcon, iconIco, 0644)
	_ = os.WriteFile(targetReadme, readmeMd, 0644)
	_ = os.WriteFile(targetWintun, wintunDll, 0644)

	// 4. Create Desktop Shortcut
	home, _ := os.UserHomeDir()
	if home != "" {
		desktopPath := filepath.Join(home, "Desktop", "Fyrnox.lnk")
		_ = createShortcut(targetExe, desktopPath, targetIcon, installDir)

		oneDriveDesktop := filepath.Join(home, "OneDrive", "Desktop")
		if _, err := os.Stat(oneDriveDesktop); err == nil {
			_ = createShortcut(targetExe, filepath.Join(oneDriveDesktop, "Fyrnox.lnk"), targetIcon, installDir)
		}
	}

	// 5. Create Start Menu Shortcut & folder
	appData := os.Getenv("APPDATA")
	if appData != "" {
		startMenuDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Fyrnox")
		_ = os.MkdirAll(startMenuDir, 0755)
		_ = createShortcut(targetExe, filepath.Join(startMenuDir, "Fyrnox.lnk"), targetIcon, installDir)
		_ = createShortcut(targetUninstaller, filepath.Join(startMenuDir, "Uninstall Fyrnox.lnk"), targetIcon, installDir)
	}

	// 6. Register in Windows Add/Remove Programs (Control Panel / Settings Apps)
	uninstallRegKey := `Software\Microsoft\Windows\CurrentVersion\Uninstall\Fyrnox`
	if k, _, err := registry.CreateKey(registry.CURRENT_USER, uninstallRegKey, registry.SET_VALUE); err == nil {
		_ = k.SetStringValue("DisplayName", "Fyrnox")
		_ = k.SetStringValue("DisplayVersion", "1.0.4")
		_ = k.SetStringValue("Publisher", "Powered by Yousef Kia")
		_ = k.SetStringValue("DisplayIcon", targetIcon)
		_ = k.SetStringValue("UninstallString", fmt.Sprintf("\"%s\"", targetUninstaller))
		_ = k.SetStringValue("InstallLocation", installDir)
		_ = k.SetStringValue("URLInfoAbout", "https://github.com/yousefkiamozafari")
		_ = k.SetDWordValue("NoModify", 1)
		_ = k.SetDWordValue("NoRepair", 1)
		_ = k.SetDWordValue("EstimatedSize", uint32(len(fyrnoxExe)+len(uninstallExe))/1024)
		_ = k.Close()
	}

	if !*silent {
		launchResp := showMsg("Fyrnox Setup Complete",
			"Fyrnox v1.0.4 has been installed successfully!\n\n"+
				"Shortcuts have been created on your Desktop and Start Menu.\n\n"+
				"Would you like to launch Fyrnox now?\n\n"+
				"برنامه با موفقیت نصب شد! آیا مایلید اکنون Fyrnox اجرا شود؟",
			mbYesNo|mbIconInfo)

		if launchResp == idYes {
			launchCmd := exec.Command(targetExe)
			launchCmd.Dir = installDir
			_ = launchCmd.Start()
		}
	}
}
