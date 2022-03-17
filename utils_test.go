package main

import (
	"testing"
)

func Test_splitpath(t *testing.T) {
	check := []string{
		"/one/two",
		"/one/two/",
		"one/two",
		"one/two/",
		"one",
	}

	for _, c := range check {
		t.Logf("%s: %#v", c, splitpath(c))
	}
}

func Test_splitkvp(t *testing.T) {
	result := splitkvp("one=1 two=\"do ah\" three=\"not = four\" five= nada")
	t.Logf("splitkvp = %#v", result)
}
