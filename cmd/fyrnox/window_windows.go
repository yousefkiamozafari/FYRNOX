package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
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
)

const (
	gwlStyle           = ^uintptr(15) // -16
	wsCaption          = 0x00C00000
	wsThickFrame       = 0x00040000
	wsMinimizeBox      = 0x00020000
	wsMaximizeBox      = 0x00010000
	swpNoMove          = 0x0002
	swpNoSize          = 0x0001
	swpNoZOrder        = 0x0004
	swpFrameChanged    = 0x0020
	wmNCLButtonDown    = 0x00A1
	htCaption          = 2
	swMinimize         = 6
	swMaximize         = 3
	swRestore          = 9
	wmClose            = 0x0010
	wmSetIcon          = 0x0080
	iconSmall          = 0
	iconBig            = 1
	imageIcon          = 1
	lrDefaultSize      = 0x0040
	lrShared           = 0x8000
	lrLoadFromFile     = 0x0010
	appUserModelID     = "FYRNOX.Client.App.v1"
)

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
