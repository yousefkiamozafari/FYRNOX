package main

import (
	"net/http"
	"time"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

var (
	modDwmapi                         = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttribute         = modDwmapi.NewProc("DwmSetWindowAttribute")
	modUser32                         = syscall.NewLazyDLL("user32.dll")
	procSetProcessDpiAwarenessContext = modUser32.NewProc("SetProcessDpiAwarenessContext")
	modShcore                         = syscall.NewLazyDLL("shcore.dll")
	procSetProcessDpiAwareness        = modShcore.NewProc("SetProcessDpiAwareness")
	procGetConsoleWindow              = kernel32Mod.NewProc("GetConsoleWindow")
)

func hideConsoleWindow() {
	if procGetConsoleWindow.Find() == nil && procShowWindow.Find() == nil {
		if hwnd, _, _ := procGetConsoleWindow.Call(); hwnd != 0 {
			const swHide = 0
			procShowWindow.Call(hwnd, swHide)
		}
	}
}

const (
	dwmwaUseImmersiveDarkModeBefore20H1 = 19
	dwmwaUseImmersiveDarkMode           = 20
	dwmwaCaptionColor                   = 35
	dwmwaTextColor                      = 36
	dpiAwarenessContextPerMonitorAwareV2 = ^uintptr(3) // -4
)

func enableHighDpiAwareness() {
	if procSetProcessDpiAwarenessContext.Find() == nil {
		procSetProcessDpiAwarenessContext.Call(dpiAwarenessContextPerMonitorAwareV2)
		return
	}
	if procSetProcessDpiAwareness.Find() == nil {
		procSetProcessDpiAwareness.Call(2) // PROCESS_PER_MONITOR_DPI_AWARE
		return
	}
	procSetProcessDPIAware := modUser32.NewProc("SetProcessDPIAware")
	if procSetProcessDPIAware.Find() == nil {
		procSetProcessDPIAware.Call()
	}
}

func enableDarkTitleBar(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	var dark int32 = 1
	// Set dark mode for native title bar (Windows 10 2004+ and Windows 11)
	r1, _, _ := procDwmSetWindowAttribute.Call(hwnd, uintptr(dwmwaUseImmersiveDarkMode), uintptr(unsafe.Pointer(&dark)), unsafe.Sizeof(dark))
	if r1 != 0 {
		// Fallback for Windows 10 1809-1909
		procDwmSetWindowAttribute.Call(hwnd, uintptr(dwmwaUseImmersiveDarkModeBefore20H1), uintptr(unsafe.Pointer(&dark)), unsafe.Sizeof(dark))
	}

	// On Windows 11: Set caption background color to deep obsidian #120A08 (BGR: 0x00080A12)
	var captionColor uint32 = 0x00080A12
	procDwmSetWindowAttribute.Call(hwnd, uintptr(dwmwaCaptionColor), uintptr(unsafe.Pointer(&captionColor)), unsafe.Sizeof(captionColor))

	// On Windows 11: Set caption text color to high-contrast #F8FAFC (BGR: 0x00FCFAF8)
	var textColor uint32 = 0x00FCFAF8
	procDwmSetWindowAttribute.Call(hwnd, uintptr(dwmwaTextColor), uintptr(unsafe.Pointer(&textColor)), unsafe.Sizeof(textColor))
}

func terminateOldInstances(exeName string) {
	currentPid := os.Getpid()
	if exeName == "" {
		exeName = "Fyrnox.exe"
	}
	// Terminate any other orphaned instances silently without console window
	cmd := newSilentCmd("taskkill", "/F", "/FI", fmt.Sprintf("PID ne %d", currentPid), "/IM", exeName)
	_ = cmd.Run()
}

func main() {
	hideConsoleWindow()
	initAppUserModelID()
	enableHighDpiAwareness()

	// Ensure working directory and DLL search path point to the executable directory
	var exeName string = "Fyrnox.exe"
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		_ = windows.SetDllDirectory(exeDir)
		_ = os.Chdir(exeDir)
		exeName = filepath.Base(exePath)
	}

	portFlag := flag.Int("port", 18080, "Dashboard HTTP and API port")
	headlessFlag := flag.Bool("headless", false, "Run in background without opening window")
	cliFlag := flag.Bool("cli", false, "Run in console mode")
	elevatedFlag := flag.Bool("elevated", false, "Internal flag indicating process was elevated via UAC")
	serverFlag := flag.String("server", "", "Server hostname to auto-select")
	modeFlag := flag.String("mode", "", "Mode to activate (tun or proxy)")
	restoreFlag := flag.Bool("restore", false, "Restore window from stealth panic mode")
	flag.Parse()

	if *restoreFlag {
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Post(fmt.Sprintf("http://127.0.0.1:%d/api/panic/restore", *portFlag), "application/json", nil)
		if err == nil {
			resp.Body.Close()
			fmt.Println("FYRNOX stealth panic restored.")
			return
		}
		fmt.Printf("Could not reach running FYRNOX instance: %v\n", err)
		return
	}

	// Clean up any old orphaned instances from previous sessions
	terminateOldInstances(exeName)

	// Clean up any old orphaned proxy settings from crashed sessions
	CleanOrphanedProxy()
	CleanupStaleKillSwitchRules()

	fmt.Println("==================================================")
	fmt.Println("           Fyrnox for Windows v1.0.4              ")
	fmt.Println("    Fast, Subscription-Free Firefox VPN Client    ")
	fmt.Println("==================================================")

	// Backup current system proxy before any modifications
	_, _ = BackupCurrentProxySettings()

	engine := NewVpnEngine()
	if *serverFlag != "" {
		_ = engine.SelectServer(*serverFlag)
	}
	if *modeFlag == "tun" || (*elevatedFlag && IsAdmin()) {
		settings := engine.GetSettings()
		settings.EnableTunMode = true
		engine.UpdateSettings(settings)
		LogInfo("Elevation initialized with Wintun Layer-3 TUN mode enabled")
	}

	apiServer := NewApiServer(engine, *portFlag)
	if err := apiServer.Start(); err != nil {
		LogError("Failed to start API server: %v", err)
		os.Exit(1)
	}

	// Initialize global emergency panic hotkey (Ctrl + Alt + Shift + X)
	InitPanicHotKey(engine)

	// Trap termination signals to cleanly restore system proxy and routes
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		LogInfo("Shutting down Fyrnox...")
		_ = globalKillSwitch.Disarm()
		_ = engine.Disconnect()
		_ = RestoreSystemProxy()
		LogInfo("Cleanup complete. Goodbye!")
		os.Exit(0)
	}()

	dashboardURL := fmt.Sprintf("http://127.0.0.1:%d", *portFlag)
	LogInfo("Fyrnox is ready! Hosting UI at %s", dashboardURL)

	if !*headlessFlag && !*cliFlag {
		launchAppWindow(dashboardURL, engine)
		LogInfo("Native desktop window closed; cleaning up proxy and exiting...")
		if globalTray != nil {
			globalTray.Remove()
		}
		_ = globalKillSwitch.Disarm()
		_ = engine.Disconnect()
		_ = RestoreSystemProxy()
		LogInfo("Shutdown complete.")
		os.Exit(0)
	} else {
		// Keep running in headless or cli mode
		select {}
	}
}

// launchAppWindow hosts a native desktop window via embedded WebView2
func launchAppWindow(targetURL string, engine *VpnEngine) {
	LogInfo("Initializing native embedded desktop window via WebView2...")

	userDataDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Fyrnox", "WebView2")
	_ = os.MkdirAll(userDataDir, 0755)

	var w webview2.WebView
	func() {
		defer func() {
			if r := recover(); r != nil {
				LogWarn("WebView2 customized initialization recovered from panic: %v", r)
				w = nil
			}
		}()

		w = webview2.NewWithOptions(webview2.WebViewOptions{
			Debug:     false,
			AutoFocus: true,
			DataPath:  userDataDir,
			WindowOptions: webview2.WindowOptions{
				Title:  "FYRNOX — Premium Desktop VPN",
				Width:  1240,
				Height: 780,
				IconId: 0,
			},
		})
	}()

	if w == nil {
		LogWarn("WebView2 customized initialization returned nil; attempting standard constructor...")
		func() {
			defer func() {
				if r := recover(); r != nil {
					LogWarn("WebView2 standard constructor recovered from panic: %v", r)
					w = nil
				}
			}()
			w = webview2.New(false)
		}()
	}

	if w != nil {
		defer w.Destroy()
		hwnd := uintptr(w.Window())
		if hwnd != 0 {
			setGlobalHwnd(hwnd)
			enableDarkTitleBar(hwnd)
			enableDesktopAcrylic(hwnd)
			applyFramelessStyle(hwnd)
			bindWindowIcon(hwnd)
			InitSystemTray(hwnd, engine)
		}
		enableWebView2Transparency(w)

		_ = w.Bind("nativeWindowDrag", func() {
			handleWindowDrag(hwnd)
		})
		_ = w.Bind("nativeWindowMinimize", func() {
			handleWindowMinimize(hwnd)
		})
		_ = w.Bind("nativeWindowMaximize", func() {
			handleWindowMaximize(hwnd)
		})
		_ = w.Bind("nativeWindowClose", func() {
			handleWindowClose(hwnd)
		})

		w.SetTitle("FYRNOX")
		w.SetSize(1240, 780, webview2.HintNone)
		if hwnd != 0 {
			applyFramelessStyle(hwnd)
		}
		w.Navigate(targetURL)
		LogInfo("Native frameless Liquid Glass window initialized.")
		w.Run()
		LogInfo("Native desktop window closed by user.")
		return
	}

	LogError("Failed to initialize embedded WebView2 window. Falling back to default browser...")
	// Show friendly message box to inform user
	msgPtr, _ := syscall.UTF16PtrFromString("Microsoft Edge WebView2 Runtime is required for the embedded window.\n\nOpening FYRNOX in your default browser at " + targetURL)
	titlePtr, _ := syscall.UTF16PtrFromString("FYRNOX Desktop Notice")
	windows.MessageBox(0, msgPtr, titlePtr, windows.MB_ICONINFORMATION|windows.MB_OK)

	// Launch default web browser
	urlPtr, _ := syscall.UTF16PtrFromString(targetURL)
	openVerb, _ := syscall.UTF16PtrFromString("open")
	_ = windows.ShellExecute(0, openVerb, urlPtr, nil, nil, windows.SW_SHOWNORMAL)

	// Block so background server stays alive while user interacts via browser
	select {}
}
