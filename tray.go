package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
)

// a tray represents all the contents
// (config, fs, and post-processed) of
// one directory path
type Tray interface {
	Path() string
	Glass(string) (Glass, error)
}

type trayDir struct {
	path   string      // full filesystem path to directory
	info   fs.FileInfo // Name(), etc
	config interface{} // raw contents from butler.yaml
	parent *Tray

	files []Glass // actual files on filesystem for change detection
	// Scripts // cached output from scripts for processing into output
	cache []Glass // copies of static or generated output cached for possible reuse
}

func (t trayDir) Path() string {
	return t.path
}

func (t trayDir) Glass(name string) (Glass, error) {
	for _, file := range t.files {
		if file.Name() == name {
			return file, nil
		}
	}
	return nil, fmt.Errorf("File not found '%s'", name)
}

func Cast(path string) Tray {
	stat, sErr := os.Stat(path)
	if sErr != nil {
		// log the error
		return nil
	}

	// go ahead and create tray now so glasses can point back to it
	tray := new(trayDir)
	tray.path = path
	tray.info = stat
	// files: glasses,

	glasses := []Glass{}
	files, rdErr := ioutil.ReadDir(path)
	if rdErr != nil {
		// log the error
		return nil
	}
	for _, file := range files {
		glasses = append(glasses, Blow(tray, &file))
	}
	tray.files = glasses

	return tray
}
