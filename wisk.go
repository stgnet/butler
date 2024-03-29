package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (g *glassDeny) Match(name string) int {
	return g.source.Match(name)
}

func (g *glassDeny) Get(keys ...string) interface{} {
	return g.source.Tray().Get(keys...)
}

func (g *glassDeny) Data(keys url.Values) interface{} {
	return nil
}

func DenyWisk(source Glass) Glass {
	return &glassDeny{
		source: source,
	}
}

func RawWisk(source Glass) Glass {
	return source
}
