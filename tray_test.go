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
	root = tray

	_, mustErr := tray.Glass("does_not_exist.nofile")
	if mustErr == nil {
		t.Fatalf("did not receive error for nonexistant file")
	}
	glass, gErr := tray.Glass("ssitest.html")
	if gErr != nil {
		t.Fatalf("failed to locate file: %v", gErr)
	}

	t.Log("Pouring ssitest.html")
	glass.Pour(os.Stdout)
}

func TestTrayConfig(t *testing.T) {
	t.Log("testing www target tray")
	tray := Cast("./www", nil)
	if tray == nil {
		t.Errorf("Cast failed")
	}
	root = tray

	// also check subdir operation
	sub, sErr := tray.Pass("include")
	if sErr != nil {
		t.Fatalf("Can't get include subdir: %v", sErr)
	}

	// make sure test key exists
	result := tray.Get("bogus")
	if result == nil {
		t.Fatalf("could not get bogus test string")
	}

	if result != "test string" {
		t.Errorf("bogus has wrong value: %#v", result)
	}

	// make sure test key shows up in sub
	result = sub.Get("bogus")
	if result == nil {
		t.Fatalf("could not get test string from include sub tray")
	}
	if result != "test string" {
		t.Errorf("bogus has wrong value in include: %#v", result)
	}

	// make sure the html type is found
	p, wisk := tray.FileMatch("ssitest.html")
	if p <= 0 {
		t.Errorf("ssitest.html not found in filematch")
	}
	if wisk != "ssi" {
		t.Errorf("ssitest.html matched wisk '%s' instead of 'ssi'", wisk)
	}

	// make sure that the html type can be found in a subdir
	p, wisk = sub.FileMatch("head.html")
	if p <= 0 {
		t.Errorf("head.html not found in filematch")
	}
	if wisk != "ssi" {
		t.Errorf("head.html matched wisk '%s' instead of 'ssi'", wisk)
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
