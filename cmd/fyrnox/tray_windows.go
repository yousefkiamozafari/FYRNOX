package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	procShellNotifyIconW     = shell32Mod.NewProc("Shell_NotifyIconW")
	procCreatePopupMenu      = user32Mod.NewProc("CreatePopupMenu")
	procAppendMenuW          = user32Mod.NewProc("AppendMenuW")
	procTrackPopupMenuEx     = user32Mod.NewProc("TrackPopupMenuEx")
	procDestroyMenu          = user32Mod.NewProc("DestroyMenu")
	procGetCursorPos         = user32Mod.NewProc("GetCursorPos")
	procSetForegroundWindow  = user32Mod.NewProc("SetForegroundWindow")
)

const (
	nimAdd        = 0x00000000
	nimModify     = 0x00000001
	nimDelete     = 0x00000002
	nifMessage    = 0x00000001
	nifIcon       = 0x00000002
	nifTip        = 0x00000004
	mfString      = 0x00000000
	mfSeparator   = 0x00000800
	mfDisabled    = 0x00000002
	mfChecked     = 0x00000008
	tpmRightalign = 0x0008
	tpmBottomalign = 0x0020
	tpmReturncmd  = 0x0100
	wmLButtonUp   = 0x0202
	wmRButtonUp   = 0x0205
	wmTrayMessage = 0x0400 + 101 // WM_USER + 101
)

type notifyIconDataW struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uTimeoutOrVersion uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type point struct {
	x int32
	y int32
}

type TrayManager struct {
	hwnd      uintptr
	hIcon     uintptr
	nid       notifyIconDataW
	engine    *VpnEngine
	installed bool
}

var globalTray *TrayManager

// InitSystemTray registers FYRNOX in the Windows system notification area (Tray)
func InitSystemTray(hwnd uintptr, engine *VpnEngine) *TrayManager {
	if hwnd == 0 {
		return nil
	}

	hMod, _, _ := procGetModuleHandleW.Call(0)
	hIcon, _, _ := procLoadImageW.Call(hMod, uintptr(1), imageIcon, 16, 16, lrShared)
	if hIcon == 0 {
		hIcon, _, _ = procLoadIconW.Call(hMod, uintptr(1))
	}
	if hIcon == 0 {
		candidates := []string{"icon.ico"}
		if exePath, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "icon.ico"))
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				if p, err := syscall.UTF16PtrFromString(path); err == nil {
					hIcon, _, _ = procLoadImageW.Call(0, uintptr(unsafe.Pointer(p)), imageIcon, 16, 16, lrLoadFromFile)
					break
				}
			}
		}
	}

	t := &TrayManager{
		hwnd:   hwnd,
		hIcon:  hIcon,
		engine: engine,
	}

	t.nid.cbSize = uint32(unsafe.Sizeof(t.nid))
	t.nid.hWnd = hwnd
	t.nid.uID = 1001
	t.nid.uFlags = nifMessage | nifIcon | nifTip
	t.nid.uCallbackMessage = wmTrayMessage
	t.nid.hIcon = hIcon

	tip, _ := syscall.UTF16FromString("FYRNOX Desktop VPN")
	copy(t.nid.szTip[:], tip)

	r, _, _ := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&t.nid)))
	if r != 0 {
		t.installed = true
		LogInfo("System tray icon successfully attached.")
	}

	globalTray = t
	return t
}

// UpdateTrayStatus updates tooltip and tray icon state based on VPN connection
func (t *TrayManager) UpdateTrayStatus(state VpnState, server string) {
	if !t.installed {
		return
	}

	var tipText string
	switch state {
	case StateConnected:
		tipText = fmt.Sprintf("FYRNOX: Connected (%s)", server)
	case StateConnecting:
		tipText = fmt.Sprintf("FYRNOX: Connecting to %s...", server)
	case StateDisconnecting:
		tipText = "FYRNOX: Disconnecting..."
	default:
		tipText = "FYRNOX: Disconnected"
	}

	tip, _ := syscall.UTF16FromString(tipText)
	for i := range t.nid.szTip {
		t.nid.szTip[i] = 0
	}
	copy(t.nid.szTip[:], tip)

	_, _, _ = procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&t.nid)))
}

// ShowTrayContextMenu renders the native Win32 popup menu on right-click
func (t *TrayManager) ShowTrayContextMenu() {
	if t.hwnd == 0 {
		return
	}

	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(t.hwnd)

	state := StateDisconnected
	if t.engine != nil {
		state = t.engine.GetStatus().State
	}

	// Menu items IDs
	const (
		idTitle       = 201
		idConnect     = 202
		idModeProxy   = 203
		idModeTun     = 204
		idModeGame    = 205
		idShowDash    = 206
		idExit        = 207
	)

	// 1. Title Item
	titleStr, _ := syscall.UTF16PtrFromString("FYRNOX Desktop v1.0.4")
	procAppendMenuW.Call(hMenu, mfString|mfDisabled, idTitle, uintptr(unsafe.Pointer(titleStr)))
	procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)

	// 2. Connect / Disconnect Action
	var connLabel string
	if state == StateConnected {
		connLabel = "Disconnect"
	} else if state == StateConnecting {
		connLabel = "Connecting..."
	} else {
		connLabel = "Connect Now"
	}
	connStr, _ := syscall.UTF16PtrFromString(connLabel)
	procAppendMenuW.Call(hMenu, mfString, idConnect, uintptr(unsafe.Pointer(connStr)))
	procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)

	// 3. Routing Modes
	var proxyFlag uintptr = mfString
	var tunFlag uintptr = mfString
	var gameFlag uintptr = mfString

	status := t.engine.GetStatus()
	if !status.TunActive {
		proxyFlag |= mfChecked
	} else {
		tunFlag |= mfChecked
	}
	if globalGamer.GetStats().Active {
		gameFlag |= mfChecked
	}

	pStr, _ := syscall.UTF16PtrFromString("System Proxy Mode")
	procAppendMenuW.Call(hMenu, proxyFlag, idModeProxy, uintptr(unsafe.Pointer(pStr)))

	tStr, _ := syscall.UTF16PtrFromString("Wintun Layer-3 TUN")
	procAppendMenuW.Call(hMenu, tunFlag, idModeTun, uintptr(unsafe.Pointer(tStr)))

	gStr, _ := syscall.UTF16PtrFromString("⚡ Gaming Accelerator")
	procAppendMenuW.Call(hMenu, gameFlag, idModeGame, uintptr(unsafe.Pointer(gStr)))

	procAppendMenuW.Call(hMenu, mfSeparator, 0, 0)

	// 4. Open & Exit
	openStr, _ := syscall.UTF16PtrFromString("Open FYRNOX Window")
	procAppendMenuW.Call(hMenu, mfString, idShowDash, uintptr(unsafe.Pointer(openStr)))

	exitStr, _ := syscall.UTF16PtrFromString("Exit FYRNOX")
	procAppendMenuW.Call(hMenu, mfString, idExit, uintptr(unsafe.Pointer(exitStr)))

	cmd, _, _ := procTrackPopupMenuEx.Call(
		hMenu,
		tpmRightalign|tpmBottomalign|tpmReturncmd,
		uintptr(pt.x),
		uintptr(pt.y),
		t.hwnd,
		0,
	)

	// Handle selection
	switch cmd {
	case idConnect:
		if state == StateConnected {
			go t.engine.Disconnect()
		} else if state == StateDisconnected {
			go t.engine.Connect()
		}
	case idModeProxy:
		t.engine.SetTunMode(false)
	case idModeTun:
		t.engine.SetTunMode(true)
	case idModeGame:
		stats := globalGamer.GetStats()
		globalGamer.SetGamingMode(!stats.Active)
	case idShowDash:
		procShowWindow.Call(t.hwnd, uintptr(swRestore))
		procSetForegroundWindow.Call(t.hwnd)
	case idExit:
		_ = t.engine.Disconnect()
		t.Remove()
		os.Exit(0)
	}
}

// Remove cleans up notification tray icon on exit
func (t *TrayManager) Remove() {
	if t.installed {
		_, _, _ = procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&t.nid)))
		t.installed = false
	}
}
