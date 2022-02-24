package main

import (
	"os"
	"testing"
)

func TestTray(t *testing.T) {
	t.Log("testing Tray")
	tray := Cast(".")
	if tray == nil {
		t.Errorf("Cast failed")
	}

	_, mustErr := tray.Glass("does_not_exist.nofile")
	if mustErr == nil {
		t.Fatalf("did not receive error for nonexistant file")
	}
	glass, gErr := tray.Glass("tray_test.go")
	if gErr != nil {
		t.Fatalf("failed to locate file: %v", gErr)
	}

	glass.Pour(os.Stdout)
}
