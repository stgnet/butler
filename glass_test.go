package main

import (
	"os"
	"testing"
)

func TestGlass(t *testing.T) {
	t.Log("testing Glass")

	stat, sErr := os.Stat("glass_test.go")
	if sErr != nil {
		t.Errorf("stat failed: %v", sErr)
	}

	glass := Blow(nil, &stat)

	glass.Pour(os.Stdout)
}
