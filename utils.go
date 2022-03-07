package main

import (
	"os"
	"strings"
)

func splitpath(s string) []string {
	result := strings.Split(s, string(os.PathSeparator))
	if result[0] == "" {
		result = result[1:]
	}
	l := len(result) - 1
	if result[l] == "" {
		result = result[:l]
	}
	return result
}
