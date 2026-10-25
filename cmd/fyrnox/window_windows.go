package main

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2"
	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

var (
	user32Mod                                  = syscall.NewLazyDLL("user32.dll")
	shell32Mod                                 = syscall.NewLazyDLL("shell32.dll")
	kernel32Mod                                = syscall.NewLazyDLL("kernel32.dll")
	procGetWindowLongW                         = user32Mod.NewProc("GetWindowLongW")
	procSetWindowLongW                         = user32Mod.NewProc("SetWindowLongW")
	procSetWindowPos                           = user32Mod.NewProc("SetWindowPos")
	procShowWindow                             = user32Mod.NewProc("ShowWindow")
	procIsZoomed                               = user32Mod.NewProc("IsZoomed")
	procPostMessageW                           = user32Mod.NewProc("PostMessageW")
	procSendMessageW                           = user32Mod.NewProc("SendMessageW")
	procReleaseCapture                         = user32Mod.NewProc("ReleaseCapture")
	procSetCurrentProcessExplicitAppUserModelID = shell32Mod.NewProc("SetCurrentProcessExplicitAppUserModelID")
	procGetModuleHandleW                       = kernel32Mod.NewProc("GetModuleHandleW")
	procLoadImageW                             = user32Mod.NewProc("LoadImageW")
	procLoadIconW                              = user32Mod.NewProc("LoadIconW")
	procDwmExtendFrameIntoClientArea           = modDwmapi.NewProc("DwmExtendFrameIntoClientArea")
	procSetWindowCompositionAttribute          = user32Mod.NewProc("SetWindowCompositionAttribute")
)

const (
	gwlStyle                = ^uintptr(15) // -16
	wsCaption               = 0x00C00000
	wsThickFrame            = 0x00040000
	wsMinimizeBox           = 0x00020000
	wsMaximizeBox           = 0x00010000
	swpNoMove               = 0x0002
	swpNoSize               = 0x0001
	swpNoZOrder             = 0x0004
	swpFrameChanged         = 0x0020
	wmNCLButtonDown         = 0x00A1
	htCaption               = 2
	swMinimize              = 6
	swMaximize              = 3
	swRestore               = 9
	wmClose                 = 0x0010
	wmSetIcon               = 0x0080
	iconSmall               = 0
	iconBig                 = 1
	imageIcon               = 1
	lrDefaultSize           = 0x0040
	lrShared                = 0x8000
	lrLoadFromFile          = 0x0010
	appUserModelID          = "FYRNOX.Client.App.v1"
	dwmwaSystemBackdropType = 38
	dwmsbtTransientWindow   = 3 // Desktop Acrylic
	wcaAccentPolicy         = 19
	accentEnableAcrylic     = 4
	accentEnableBlurBehind  = 3
)

type margins struct {
	cxLeftWidth    int32
	cxRightWidth   int32
	cyTopHeight    int32
	cyBottomHeight int32
}

type accentPolicy struct {
	AccentState   uint32
	AccentFlags   uint32
	GradientColor uint32
	AnimationId   uint32
}

type windowCompositionAttributeData struct {
	Attrib uint32
	PvData uintptr
	CbData uintptr
}

type iUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

var globalHwnd uintptr

func setGlobalHwnd(hwnd uintptr) {
	globalHwnd = hwnd
}

func getGlobalHwnd() uintptr {
	return globalHwnd
}

// initAppUserModelID explicitly registers FYRNOX with the Windows taskbar manager
func initAppUserModelID() {
	if procSetCurrentProcessExplicitAppUserModelID.Find() == nil {
		appIDPtr, err := syscall.UTF16PtrFromString(appUserModelID)
		if err == nil {
			_, _, _ = procSetCurrentProcessExplicitAppUserModelID.Call(uintptr(unsafe.Pointer(appIDPtr)))
		}
	}
}

// bindWindowIcon binds the official high-resolution FYRNOX icon to the window handle and taskbar
func bindWindowIcon(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	hMod, _, _ := procGetModuleHandleW.Call(0)

	// 1. Try loading icon resource ID 1 from embedded PE resources
	hIconBig, _, _ := procLoadImageW.Call(hMod, uintptr(1), imageIcon, 32, 32, lrShared)
	hIconSmall, _, _ := procLoadImageW.Call(hMod, uintptr(1), imageIcon, 16, 16, lrShared)

	if hIconBig == 0 {
		hIconBig, _, _ = procLoadIconW.Call(hMod, uintptr(1))
	}
	if hIconSmall == 0 {
		hIconSmall, _, _ = procLoadIconW.Call(hMod, uintptr(1))
	}

	// 2. Fallback to icon.ico on disk if running from source/dev directory
	if hIconBig == 0 || hIconSmall == 0 {
		candidates := []string{"icon.ico"}
		if exePath, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "icon.ico"))
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				if p, err := syscall.UTF16PtrFromString(path); err == nil {
					if hIconBig == 0 {
						hIconBig, _, _ = procLoadImageW.Call(0, uintptr(unsafe.Pointer(p)), imageIcon, 32, 32, lrLoadFromFile)
					}
					if hIconSmall == 0 {
						hIconSmall, _, _ = procLoadImageW.Call(0, uintptr(unsafe.Pointer(p)), imageIcon, 16, 16, lrLoadFromFile)
					}
					break
				}
			}
		}
	}

	// Bind icons to window handle via WM_SETICON
	if hIconBig != 0 {
		procSendMessageW.Call(hwnd, uintptr(wmSetIcon), uintptr(iconBig), hIconBig)
	}
	if hIconSmall != 0 {
		procSendMessageW.Call(hwnd, uintptr(wmSetIcon), uintptr(iconSmall), hIconSmall)
	}
	LogInfo("Bound multi-resolution FYRNOX icon to window HWND 0x%X (big: 0x%X, small: 0x%X)", hwnd, hIconBig, hIconSmall)
}

// applyFramelessStyle removes native title bar while keeping sizing frame, Aero Snap, and minimize/maximize animation
func applyFramelessStyle(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	r, _, _ := procGetWindowLongW.Call(hwnd, gwlStyle)
	currentStyle := uint32(r)
	if currentStyle == 0 {
		return
	}

	// Remove caption bar, ensure sizing frame and box controls remain for OS window manager
	newStyle := (currentStyle &^ wsCaption) | wsThickFrame | wsMinimizeBox | wsMaximizeBox
	procSetWindowLongW.Call(hwnd, gwlStyle, uintptr(newStyle))
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, uintptr(swpNoMove|swpNoSize|swpNoZOrder|swpFrameChanged))
}

// handleWindowDrag triggers native OS window dragging (with full Aero Snap support)
func handleWindowDrag(hwnd uintptr) {
	if hwnd == 0 {
		hwnd = globalHwnd
	}
	if hwnd == 0 {
		return
	}
	procReleaseCapture.Call()
	procSendMessageW.Call(hwnd, uintptr(wmNCLButtonDown), uintptr(htCaption), 0)
}

// handleWindowMinimize minimizes the window
func handleWindowMinimize(hwnd uintptr) {
	if hwnd == 0 {
		hwnd = globalHwnd
	}
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, uintptr(swMinimize))
}

// handleWindowMaximize toggles between maximized and restored states
func handleWindowMaximize(hwnd uintptr) {
	if hwnd == 0 {
		hwnd = globalHwnd
	}
	if hwnd == 0 {
		return
	}
	zoomed, _, _ := procIsZoomed.Call(hwnd)
	if zoomed != 0 {
		procShowWindow.Call(hwnd, uintptr(swRestore))
	} else {
		procShowWindow.Call(hwnd, uintptr(swMaximize))
	}
}

// handleWindowClose requests clean window closure
func handleWindowClose(hwnd uintptr) {
	if hwnd == 0 {
		hwnd = globalHwnd
	}
	if hwnd == 0 {
		return
	}
	procPostMessageW.Call(hwnd, uintptr(wmClose), 0, 0)
}

// enableDesktopAcrylic enables native Windows Desktop Acrylic blur-behind via DWM
func enableDesktopAcrylic(hwnd uintptr) {
	if hwnd == 0 {
		return
	}

	// 1. Extend frame into entire client area so glass effect encompasses the whole window
	m := margins{-1, -1, -1, -1}
	procDwmExtendFrameIntoClientArea.Call(hwnd, uintptr(unsafe.Pointer(&m)))

	// 2. Windows 11 (22H2+): Set DWMWA_SYSTEMBACKDROP_TYPE to DWMSBT_TRANSIENTWINDOW (Acrylic = 3)
	var backdropType uint32 = dwmsbtTransientWindow
	r, _, _ := procDwmSetWindowAttribute.Call(
		hwnd,
		uintptr(dwmwaSystemBackdropType),
		uintptr(unsafe.Pointer(&backdropType)),
		unsafe.Sizeof(backdropType),
	)

	// 3. Fallback for Windows 10 & Windows 11 21H2 via SetWindowCompositionAttribute
	if r != 0 && procSetWindowCompositionAttribute.Find() == nil {
		policy := accentPolicy{
			AccentState:   accentEnableAcrylic,
			AccentFlags:   2,
			GradientColor: 0x99100D0B, // Translucent dark obsidian tint (ARGB)
		}
		data := windowCompositionAttributeData{
			Attrib: wcaAccentPolicy,
			PvData: uintptr(unsafe.Pointer(&policy)),
			CbData: unsafe.Sizeof(policy),
		}
		procSetWindowCompositionAttribute.Call(hwnd, uintptr(unsafe.Pointer(&data)))
	}
	LogInfo("Native Windows Desktop Acrylic blur-behind activated on HWND 0x%X", hwnd)
}

// enableWebView2Transparency configures WebView2 controller background to 100% transparent (Alpha=0)
func enableWebView2Transparency(w webview2.WebView) {
	if w == nil {
		return
	}

	apply := func() bool {
		defer func() {
			_ = recover()
		}()

		ifaceVal := *(*[2]unsafe.Pointer)(unsafe.Pointer(&w))
		dataPtr := ifaceVal[1]
		if dataPtr == nil {
			return false
		}

		ptrSize := unsafe.Sizeof(uintptr(0))
		// webview struct: hwnd (0) + mainthread (1) + browser.tab (2) + browser.data (3)
		browserDataPtr := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(dataPtr) + 3*ptrSize))
		if browserDataPtr == nil {
			return false
		}

		chromium := (*edge.Chromium)(browserDataPtr)
		controller := chromium.GetController()
		if controller == nil {
			return false
		}

		// IID_ICoreWebView2Controller2 = {c979903e-d57a-4dc4-a638-bf1b3713f04b}
		iid := windows.GUID{
			Data1: 0xc979903e,
			Data2: 0xd57a,
			Data3: 0x4dc4,
			Data4: [8]byte{0xa6, 0x38, 0xbf, 0x1b, 0x37, 0x13, 0xf0, 0x4b},
		}

		vtbl := *(**iUnknownVtbl)(unsafe.Pointer(controller))
		if vtbl == nil || vtbl.QueryInterface == 0 {
			return false
		}

		var controller2 *edge.ICoreWebView2Controller2
		res, _, _ := syscall.SyscallN(
			vtbl.QueryInterface,
			uintptr(unsafe.Pointer(controller)),
			uintptr(unsafe.Pointer(&iid)),
			uintptr(unsafe.Pointer(&controller2)),
		)

		if res == 0 && controller2 != nil {
			_ = controller2.PutDefaultBackgroundColor(edge.COREWEBVIEW2_COLOR{A: 0, R: 0, G: 0, B: 0})
			LogInfo("WebView2 controller background set to 100%% transparent (Alpha=0)")
			return true
		}
		return false
	}

	if !apply() {
		go func() {
			for i := 0; i < 25; i++ {
				time.Sleep(50 * time.Millisecond)
				success := false
				w.Dispatch(func() {
					if apply() {
						success = true
					}
				})
				if success {
					break
				}
			}
		}()
	}
}
