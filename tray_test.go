package main

import (
	"os"
	"testing"
)

func TestTray(t *testing.T) {
	t.Log("testing Tray")
	tray := Cast("./www", nil)
	if tray == nil {
		t.Errorf("Cast failed")
	}

	_, mustErr := tray.Glass("does_not_exist.nofile")
	if mustErr == nil {
		t.Fatalf("did not receive error for nonexistant file")
	}
	glass, gErr := tray.Glass("index.html")
	if gErr != nil {
		t.Fatalf("failed to locate file: %v", gErr)
	}

	glass.Pour(os.Stdout)
}

func TestTrayConfig(t *testing.T) {
	t.Log("testing www target tray")
	tray := Cast("./www", nil)
	if tray == nil {
		t.Errorf("Cast failed")
	}

	result := tray.Get("bogus")
	if result == nil {
		t.Fatalf("could not get bogus test string")
	}

	if result != "test string" {
		t.Errorf("bogus has wrong value: %#v", result)
	}

	// make sure the html type is found
	p, wisk := tray.FileMatch("test.html")
	if p <= 0 {
		t.Errorf("test.html not found in filematch")
	}
	if wisk != "raw" {
		t.Errorf("test.html matched wisk '%s' instead of 'raw'", wisk)
	}

	// make sure the comma separated image file matches are found
	p, wisk = tray.FileMatch("test.gif")
	if p <= 0 {
		t.Errorf("test.gif not found in filematch")
	}
	if wisk != "raw" {
		t.Errorf("test.html matched wisk '%s' instead of 'raw'", wisk)
	}

}
