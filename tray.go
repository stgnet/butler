package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// a tray represents all the contents
// (config, fs, and post-processed) of
// one directory path
type Tray interface {
	Name() string                   // just the directory name
	Path() string                   // full path to directory on filesystem
	Glass(string) (Glass, error)    // locate a file or ersatz stream
	Pass(string) (Tray, error)      // locate a sub-directory
	Get(keys ...string) interface{} // obtain config
}

type trayDir struct {
	path   string      // full filesystem path to directory
	info   fs.FileInfo // Name(), etc
	config interface{} // raw contents from butler.yaml
	parent *Tray

	dirs   []Tray  // sub-directories as other trays
	files  []Glass // actual files on filesystem for change detection
	drinks []Glass // files or data sources available to pour
}

func (t trayDir) Get(keys ...string) interface{} {
	data := t.config
	for _, key := range keys {
		switch d := data.(type) {
		case map[string]interface{}:
			data := d[key]
			if data == nil {
				log.Errorf("%s: Get did not find key '%s'", t.Path(), key)
				return nil
			}
			continue
		default:
			log.Errorf("%s: Get unknown type '%#v'", t.Path(), data)
			return nil
		}
	}
	return data
}

func (t trayDir) loadConfig() {
	file, fErr := os.Open(filepath.Join(t.Path(), "butler.yaml"))
	if fErr != nil {
		t.config = nil
		return
	}
	defer file.Close()
	dErr := yaml.NewDecoder(file).Decode(&t.config)
	if dErr != nil {
		log.Errorf("%s/butler.yaml: %v", t.Path(), dErr)
	}
}

func (t trayDir) Name() string {
	return t.info.Name()
}

func (t trayDir) Path() string {
	return t.path
}

func (t trayDir) Pass(name string) (Tray, error) {
	for _, dir := range t.dirs {
		if dir.Name() == name {
			return dir, nil
		}
	}
	return nil, fmt.Errorf("Directory not found '%s'", name)
}

var extOrder = []string{
	"html",
	"yaml",
}

func (t trayDir) Glass(name string) (Glass, error) {
	var match Glass
	for _, file := range t.files {
		if file.Name() == name {
			return file, nil
		}
		for _, ext := range extOrder {
			if file.Name() == name+"."+ext {
				match = file
				break
			}
		}
	}
	if match != nil {
		return match, nil
	}
	return nil, fmt.Errorf("File not found '%s'", name)
}

func Cast(path string, parent Tray) Tray {
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
	dirs := []string{}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		} else {
			glasses = append(glasses, Blow(tray, &file))
		}
	}
	tray.files = glasses

	tray.dirs = []Tray{}
	for _, dir := range dirs {
		tray.dirs = append(tray.dirs, Cast(filepath.Join(tray.path, dir), tray))
	}

	tray.loadConfig()

	return tray
}
