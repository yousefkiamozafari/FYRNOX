package main

import (
	"runtime"
	"sync"
	"unsafe"
)

var (
	procRegisterHotKey   = user32Mod.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32Mod.NewProc("UnregisterHotKey")
	procGetMessageW      = user32Mod.NewProc("GetMessageW")
)

const (
	modAlt        = 0x0001
	modControl    = 0x0002
	modShift      = 0x0004
	modNoRepeat   = 0x4000
	vkX           = 0x58
	panicHotKeyID = 0x7001
	wmHotKey      = 0x0312
	swHide        = 0
)

type msgStruct struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

type PanicManager struct {
	mu               sync.Mutex
	isPanic          bool
	engine           *VpnEngine
	hotkeyRegistered bool
}

var globalPanic = &PanicManager{}

// InitPanicHotKey registers the global emergency hotkey (Ctrl + Alt + Shift + X)
// on a dedicated Windows thread with a message pump.
func InitPanicHotKey(engine *VpnEngine) {
	globalPanic.mu.Lock()
	globalPanic.engine = engine
	globalPanic.mu.Unlock()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		modifiers := uintptr(modControl | modAlt | modShift | modNoRepeat)
		r, _, err := procRegisterHotKey.Call(0, uintptr(panicHotKeyID), modifiers, uintptr(vkX))
		if r == 0 {
			LogWarn("[PANIC KEY] Failed to register global hotkey Ctrl+Alt+Shift+X: %v", err)
			return
		}

		globalPanic.mu.Lock()
		globalPanic.hotkeyRegistered = true
		globalPanic.mu.Unlock()
		LogInfo("[PANIC KEY] Global stealth panic hotkey registered: Ctrl + Alt + Shift + X")

		var msg msgStruct
		for {
			ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if int32(ret) <= 0 {
				break
			}
			if msg.message == wmHotKey && msg.wParam == uintptr(panicHotKeyID) {
				globalPanic.TogglePanic()
			}
		}

		procUnregisterHotKey.Call(0, uintptr(panicHotKeyID))
	}()
}

func (p *PanicManager) TogglePanic() {
	p.mu.Lock()
	active := p.isPanic
	p.mu.Unlock()

	if active {
		p.RestorePanic()
	} else {
		p.TriggerPanic()
	}
}

func (p *PanicManager) TriggerPanic() {
	p.mu.Lock()
	p.isPanic = true
	eng := p.engine
	p.mu.Unlock()

	LogWarn("[PANIC KEY] Emergency Panic triggered! Disconnecting VPN, hiding window, removing tray icon, flushing DNS cache...")

	// 1. Immediately disconnect VPN and tear down routes
	if eng != nil {
		go func() {
			_ = eng.Disconnect()
		}()
	}

	// 2. Instantly hide GUI window
	hwnd := getGlobalHwnd()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, uintptr(swHide))
	}

	// 3. Remove tray icon completely from taskbar notification area
	if globalTray != nil {
		globalTray.Remove()
	}

	// 4. Flush system DNS resolver cache
	FlushDNSCache()
}

func (p *PanicManager) RestorePanic() {
	p.mu.Lock()
	p.isPanic = false
	p.mu.Unlock()

	LogInfo("[PANIC KEY] Restoring from stealth panic mode...")

	// 1. Restore window
	hwnd := getGlobalHwnd()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, uintptr(swRestore))
		procSetForegroundWindow.Call(hwnd)
	}

	// 2. Restore system tray icon
	if globalTray != nil {
		globalTray.Restore()
	}
}

func (p *PanicManager) IsPanicActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isPanic
}
