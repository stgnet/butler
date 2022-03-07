package main

import (
	"fmt"
	"io"
	"net/http"
)

// wisk constructs a glass that targets a source file(s) or structurered data
// and includes a definition for optionally converting to another type
// for example, yaml -> html
// the conversion itelf may not happen until Pour()

var Wisks = map[string]func(Glass) Glass{
	"deny": DenyWisk,
	"raw":  RawWisk,
	"ssi":  SSIWisk,
}

type glassDeny struct {
	source Glass
}

func (g *glassDeny) Name() string {
	return g.source.Name()
}
func (g *glassDeny) Path() string {
	return g.source.Path()
}
func (g *glassDeny) Tray() Tray {
	return g.source.Tray()
}
func (g *glassDeny) Type() string {
	return g.source.Type()
}

func (g *glassDeny) Pour(w io.Writer) error {
	return Broke{Status: http.StatusForbidden, Problem: fmt.Errorf("access denied to file '%s'", g.Name())}
}

func DenyWisk(source Glass) Glass {
	return &glassDeny{
		source: source,
	}
}

func RawWisk(source Glass) Glass {
	return source
}
