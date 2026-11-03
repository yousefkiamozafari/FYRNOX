//go:build windows

package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEmbeddedWintunDLLValidity(t *testing.T) {
	if len(embeddedWintunDLL) == 0 {
		t.Fatalf("embeddedWintunDLL is empty for GOARCH=%s", runtime.GOARCH)
	}

	// 1. Verify DOS Header Magic 'MZ'
	if len(embeddedWintunDLL) < 64 || embeddedWintunDLL[0] != 0x4D || embeddedWintunDLL[1] != 0x5A {
		t.Fatalf("embeddedWintunDLL is not a valid Windows PE DLL (missing MZ header)")
	}

	// 2. Locate PE Header
	peOffset := binary.LittleEndian.Uint32(embeddedWintunDLL[0x3C:0x40])
	if int(peOffset)+24 > len(embeddedWintunDLL) {
		t.Fatalf("invalid PE header offset: %d", peOffset)
	}

	// 3. Verify PE Signature 'PE\0\0'
	if embeddedWintunDLL[peOffset] != 'P' || embeddedWintunDLL[peOffset+1] != 'E' ||
		embeddedWintunDLL[peOffset+2] != 0 || embeddedWintunDLL[peOffset+3] != 0 {
		t.Fatalf("invalid PE signature at offset %d", peOffset)
	}

	// 4. Verify Architecture Machine Type
	machine := binary.LittleEndian.Uint16(embeddedWintunDLL[peOffset+4 : peOffset+6])
	switch runtime.GOARCH {
	case "amd64":
		if machine != 0x8664 {
			t.Errorf("expected machine 0x8664 (x64) for GOARCH=amd64, got 0x%04X", machine)
		}
	case "386":
		if machine != 0x014C {
			t.Errorf("expected machine 0x014C (x86) for GOARCH=386, got 0x%04X", machine)
		}
	case "arm64":
		if machine != 0xAA64 {
			t.Errorf("expected machine 0xAA64 (ARM64) for GOARCH=arm64, got 0x%04X", machine)
		}
	}
}

func TestEnsureWintunDLLExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	targetDll := filepath.Join(tmpDir, "wintun.dll")

	// Verify extraction creates file matching embedded byte length
	if err := os.WriteFile(targetDll, embeddedWintunDLL, 0644); err != nil {
		t.Fatalf("failed to write test wintun.dll: %v", err)
	}

	fi, err := os.Stat(targetDll)
	if err != nil {
		t.Fatalf("stat failed on extracted wintun.dll: %v", err)
	}

	if fi.Size() != int64(len(embeddedWintunDLL)) {
		t.Errorf("extracted file size %d does not match embedded size %d", fi.Size(), len(embeddedWintunDLL))
	}
}
