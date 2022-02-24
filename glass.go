package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// glass describes a file, either on disk or cached output
type Glass interface {
	fs.FileInfo     // file Name(), Size(), Mode(), ModTime()
	Pour(io.Writer) // write contents to stream
	Path() string   // full path to file
}

type glassfile struct {
	info fs.FileInfo
	tray Tray
}

func (g *glassfile) IsDir() bool {
	return false
}
func (g *glassfile) ModTime() time.Time {
	return g.info.ModTime()
}
func (g *glassfile) Mode() fs.FileMode {
	return g.info.Mode()
}
func (g *glassfile) Name() string {
	return g.info.Name()
}
func (g *glassfile) Size() int64 {
	return g.info.Size()
}
func (g *glassfile) Sys() interface{} {
	return g.info.Sys()
}
func (g *glassfile) Path() string {
	if g.tray == nil {
		// hopefully this means that it's in the current directory?
		return g.info.Name()
	}
	return filepath.Join(g.tray.Path(), g.info.Name())
}

func (g *glassfile) Pour(w io.Writer) {
	// io.WriteString(w, "test\n")
	in, inErr := os.Open(g.Path())
	if inErr != nil {
		panic(inErr)
	}
	defer in.Close()
	buf := make([]byte, 1024)
	for {
		// read a chunk
		r, rErr := in.Read(buf)
		if rErr != nil && rErr != io.EOF {
			panic(rErr)
		}
		if r == 0 {
			break
		}

		// write a chunk
		_, wErr := w.Write(buf[:r])
		if wErr != nil {
			panic(wErr)
		}
	}
}

func Blow(tray Tray, info *fs.FileInfo) Glass {
	glass := new(glassfile)
	glass.info = *info
	glass.tray = tray
	return glass
}
