package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

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

func main() {
	silent := flag.Bool("silent", false, "Run silent uninstallation")
	flag.BoolVar(silent, "s", false, "Run silent uninstallation")
	flag.Parse()

	if !*silent {
		resp := showMsg("Fyrnox Uninstall",
			"Are you sure you want to completely uninstall Fyrnox and remove all its components?\n\nآیا از حذف کامل برنامه Fyrnox اطمینان دارید؟",
			mbYesNo|mbIconQuestion)
		if resp != idYes {
			os.Exit(0)
		}
	}

	// 1. Terminate running Fyrnox process silently
	cmdKill := exec.Command("taskkill", "/F", "/IM", "Fyrnox.exe")
	cmdKill.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	_ = cmdKill.Run()

	// 2. Restore Windows proxy in registry cleanly and refresh system cache
	if k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE); err == nil {
		_ = k.SetDWordValue("ProxyEnable", 0)
		_ = k.SetStringValue("ProxyServer", "")
		_ = k.Close()
	}
	wininet := syscall.NewLazyDLL("wininet.dll")
	procInternetSetOption := wininet.NewProc("InternetSetOptionW")
	procInternetSetOption.Call(0, 39, 0, 0) // INTERNET_OPTION_SETTINGS_CHANGED
	procInternetSetOption.Call(0, 37, 0, 0) // INTERNET_OPTION_REFRESH

	// 3. Remove Desktop shortcut
	home, _ := os.UserHomeDir()
	if home != "" {
		desktopShortcut := filepath.Join(home, "Desktop", "Fyrnox.lnk")
		_ = os.Remove(desktopShortcut)
		oneDriveDesktop := filepath.Join(home, "OneDrive", "Desktop", "Fyrnox.lnk")
		_ = os.Remove(oneDriveDesktop)
	}

	// 4. Remove Start Menu folder
	appData := os.Getenv("APPDATA")
	if appData != "" {
		startMenuFolder := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Fyrnox")
		_ = os.RemoveAll(startMenuFolder)
		startMenuShortcut := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Fyrnox.lnk")
		_ = os.Remove(startMenuShortcut)
	}

	// 5. Remove Windows Uninstall Registry entries
	_ = registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Uninstall\Fyrnox`)

	// 6. Get executable directory
	exePath, err := os.Executable()
	if err == nil {
		installDir := filepath.Dir(exePath)
		// Clean files in installDir
		_ = os.Remove(filepath.Join(installDir, "Fyrnox.exe"))
		_ = os.Remove(filepath.Join(installDir, "icon.ico"))
		_ = os.Remove(filepath.Join(installDir, "README.md"))
		_ = os.Remove(filepath.Join(installDir, "wintun.dll"))

		// Schedule self-deletion of uninstaller and install directory silently via hidden cmd
		cmdStr := fmt.Sprintf("timeout /t 2 /nobreak >nul & del /f /q \"%s\" >nul 2>&1 & rmdir /s /q \"%s\" >nul 2>&1", exePath, installDir)
		cmdDel := exec.Command("cmd.exe", "/c", cmdStr)
		cmdDel.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		cmdDel.Start()
	}

	if !*silent {
		showMsg("Fyrnox Uninstall",
			"Fyrnox has been successfully uninstalled from your computer.\n\nبرنامه با موفقیت از سیستم شما حذف شد.",
			mbOK|mbIconInfo)
	}
}
