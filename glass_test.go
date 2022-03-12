package main

import (
	"os"
	"testing"
)

func TestGlass(t *testing.T) {
	t.Log("testing Glass")

	stat, sErr := os.Stat("www/testfile")
	if sErr != nil {
		t.Errorf("stat failed: %v", sErr)
	}

	glass := Blow(nil, &stat, 0)

	glass.Pour(os.Stdout)
}
