package main

import (
	"errors"
	"strings"
	"testing"
	"golang.org/x/sys/windows"
)

func TestBuildElevationArgs(t *testing.T) {
	args := buildElevationArgs("us.m1.fastly-masque.net", "tun")
	if !strings.Contains(args, "-elevated") {
		t.Errorf("expected args to contain -elevated, got %q", args)
	}
	if !strings.Contains(args, "-server us.m1.fastly-masque.net") {
		t.Errorf("expected args to contain server flag, got %q", args)
	}
	if !strings.Contains(args, "-mode tun") {
		t.Errorf("expected args to contain mode flag, got %q", args)
	}
}

func TestIsUACCancelled(t *testing.T) {
	if !isUACCancelled(windows.ERROR_CANCELLED) {
		t.Errorf("expected windows.ERROR_CANCELLED to be identified as cancelled")
	}
	if !isUACCancelled(ErrUACCancelled) {
		t.Errorf("expected ErrUACCancelled to be identified as cancelled")
	}
	if isUACCancelled(errors.New("some other error")) {
		t.Errorf("expected random error not to be identified as cancelled")
	}
	if isUACCancelled(nil) {
		t.Errorf("expected nil error not to be identified as cancelled")
	}
}
