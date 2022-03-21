package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
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
	out  io.Writer
	buf  []byte
	vars map[string]string
	tray Tray
}

func (w *writeSSI) Error(c string, e error) {
	w.out.Write([]byte(fmt.Sprintf("[an error occurred while processing %s: %v\n", c, e)))
	log.Errorf("SSI %s: %v", c, e)
}

func (w *writeSSI) Command(cmd string) {
	c := strings.SplitN(cmd, " ", 2)
	params := splitkvp(c[1])
	// log.Infof("SSI %s %#v", c[0], params)
	switch c[0] {
	case "set":
		if w.vars == nil {
			w.vars = make(map[string]string)
		}
		key := params["var"]
		value := params["value"]
		w.vars[key] = value

	case "echo":
		if w.vars == nil {
			return
		}
		key := params["var"]
		value, exists := w.vars[key]
		if exists {
			w.out.Write([]byte(value))
		}

	case "include":
		virtual := params["virtual"]
		// log.Infof("including %#v", virtual)
		tray := w.tray
		if virtual[0] == '/' && root != nil {
			virtual = virtual[1:]
			// root is host tray which has root as parent
			for {
				parent := tray.Parent()
				if parent == nil || parent == root {
					break
				}
				log.Infof("Searching for host tray %s -> %s", tray.Path(), parent.Path())
				tray = parent
			}
		}
		paths := splitpath(virtual)
		dirs := paths[0 : len(paths)-1]
		filename := paths[len(paths)-1]
		for _, dirname := range dirs {
			log.Infof("dirname = %#v", dirname)
			sub, tErr := tray.Pass(dirname)
			if tErr != nil {
				w.Error(cmd, tErr)
				return
			}
			tray = sub
		}
		glass, gErr := tray.Glass(filename)
		if gErr != nil {
			w.Error(cmd, gErr)
			return
		}
		glass.Pour(w)
	default:
		w.Error(cmd, fmt.Errorf("unexpected"))
	}
}

func (w *writeSSI) Write(p []byte) (int, error) {
	b := append(w.buf, p...)
	i := 0
next:
	for {
		for i < len(b)-8 {
			if b[i+0] == '<' && b[i+1] == '!' && b[i+2] == '-' && b[i+3] == '-' && b[i+4] == '#' {
				n, wErr := w.out.Write(b[:i])
				if wErr != nil {
					return n, wErr
				}
				b = b[i:]
				i = 0
				for i < len(b)-3 {
					if b[i+0] == '-' && b[i+1] == '-' && b[i+2] == '>' {
						w.Command(string(b[5:i]))
						b = b[i+3:]
						i = 0
						continue next
					}
					i++
				}
				// log.Infof("return 2 with buf '%s'", string(b))
				w.buf = b
				return len(p), nil
			}
			i++
		}
		n, wwErr := w.out.Write(b)
		if wwErr != nil {
			return n, wwErr
		}
		w.buf = []byte{}
		// log.Infof("return 1")
		return len(p), nil
	}
}

func (g *glassSSI) Pour(w io.Writer) error {
	return g.source.Pour(&writeSSI{out: w, tray: g.Tray()})
}

func (g *glassSSI) Match(name string) int {
	return g.source.Match(name)
}

func SSIWisk(source Glass) Glass {
	return &glassSSI{source: source}
}
