package main

import (
	"os"
	"strings"
	"unicode"
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

func splitkvp(s string) map[string]string {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}

	kvp := map[string]string{}
	for _, field := range strings.FieldsFunc(s, f) {
		pair := strings.SplitN(field, "=", 2)
		value := ""
		if len(pair) == 2 {
			value = pair[1]
		}
		if len(value) > 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		kvp[pair[0]] = value
	}
	return kvp
}
