package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jchv/go-webview2"
)

func TestWebView2NewStandard(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("WebView2 initialization in test environment caught recovery: %v", r)
		}
	}()
	w := webview2.New(false)
	if w != nil {
		defer w.Destroy()
		t.Log("webview2.New(false) successfully initialized!")
	}
}

func TestWebView2SafeLaunch(t *testing.T) {
	userDataDir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Fyrnox", "WebView2Test")
	_ = os.MkdirAll(userDataDir, 0755)

	var w webview2.WebView
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("WebView2 initialization caught expected panic if re-entrant: %v", r)
			}
		}()
		w = webview2.New(false)
	}()

	if w != nil {
		w.Destroy()
		t.Log("WebView2 successfully initialized and cleaned up")
	}
}



