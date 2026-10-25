package main

import (
	"testing"
)

func TestNewSilentCmd(t *testing.T) {
	cmd := newSilentCmd("route", "print")
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}
	if cmd.SysProcAttr == nil {
		t.Fatal("expected SysProcAttr to be non-nil")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Errorf("expected HideWindow to be true")
	}
	const expectedCreationFlags = 0x08000000 // CREATE_NO_WINDOW
	if cmd.SysProcAttr.CreationFlags&expectedCreationFlags != expectedCreationFlags {
		t.Errorf("expected CreationFlags to have CREATE_NO_WINDOW (0x08000000), got 0x%X", cmd.SysProcAttr.CreationFlags)
	}
}
