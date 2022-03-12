package main

import (
	"io"
)

type glassSSI struct {
	source Glass
}

func (g *glassSSI) Name() string {
	return g.source.Name()
}
func (g *glassSSI) Path() string {
	return g.source.Path()
}
func (g *glassSSI) Tray() Tray {
	return g.source.Tray()
}
func (g *glassSSI) Type() string {
	return "text/html"
}

type writeSSI struct {
	out io.Writer
	buf []byte
}

func (w *writeSSI) Command(c string) {
	w.out.Write([]byte("### " + c + " ###"))
}

func (w *writeSSI) Write(p []byte) (int, error) {
	b := append(w.buf, p...)
	i := 0
next:
	for {
		for i < len(b)-5 {
			if b[i+0] == '<' && b[i+1] == '!' && b[i+2] == '-' && b[i+3] == '-' && b[i+4] == '#' {
				n, wErr := w.out.Write(p[:i])
				if wErr != nil {
					return n, wErr
				}
				b = b[i:]
				i = 0
				break
			}
			i++
		}
		i = 5
		for i < len(b)-3 {
			if b[i+0] == '-' && b[i+1] == '-' && b[i+2] == '>' {
				w.Command(string(b[5 : i-5]))
				b = b[i+3:]
				i = 0
				continue next
			}
			i++
		}
		return len(p), nil
	}
}

func (g *glassSSI) Pour(w io.Writer) error {
	return g.source.Pour(&writeSSI{out: w})
}

func (g *glassSSI) Match(name string) int {
	return g.source.Match(name)
}

func SSIWisk(source Glass) Glass {
	return &glassSSI{source: source}
}
