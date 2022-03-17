package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// a tray represents all the contents
// (config, fs, and post-processed) of
// one directory path
type Tray interface {
	Name() string                   // just the directory name
	Path() string                   // full path to directory on filesystem
	Glass(string) (Glass, error)    // locate a file, drink, or ersatz
	Pass(string) (Tray, error)      // locate a sub-directory
	Get(keys ...string) interface{} // obtain config
	FileMatch(string) (int, string) // return index and type matching name
}

type trayDir struct {
	path   string      // full filesystem path to directory
	info   fs.FileInfo // Name(), etc
	config interface{} // raw contents from butler.yaml
	parent Tray

	dirs   []Tray  // sub-directories as other trays
	files  []Glass // actual files on filesystem for change detection
	drinks []Glass // files or data sources available to pour
}

func (t *trayDir) FileMatch(name string) (int, string) {
	filematch := t.Get("filematch")
	if filematch == nil {
		log.Infof("Tray %s has no filematch configuration", t.Path())
		return 0, ""
	}
	for i, m := range filematch.([]interface{}) {
		switch r := m.(type) {
		case map[string]interface{}:
			for patterns, wisk := range r {
				for _, pattern := range strings.Split(patterns, ",") {
					pat := strings.TrimSpace(pattern)
					match, mErr := filepath.Match(pat, name)
					if mErr != nil {
						log.Errorf("Pattern error in filematch %d %s: %v", i, pat, mErr)
						continue
					}
					if match {
						switch wisk.(type) {
						case string:
							return i + 1, wisk.(string)
						default:
							log.Errorf("Unexpected wisk type in filematch %d %s: %#v", i, pat, wisk)
							continue
						}
					}
				}
			}
		default:
			log.Errorf("Unknown filematch type: %#v", m)
		}
	}
	log.Warningf("filematch had no match for '%s'", name)
	return 0, ""
}

func (t *trayDir) Get(keys ...string) interface{} {
	// log.Infof("%s: looking for key %#v", t.Path(), keys)
	data := t.config
	if data == nil {
		if t.parent != nil {
			return t.parent.Get(keys...)
		}
		// TODO: also check default configs
		log.Errorf("%s: has no config data, no parent", t.Path())
		return nil
	}
	for _, key := range keys {
		switch d := data.(type) {
		case map[string]interface{}:
			data = d[key]
			if data == nil {
				// try again with parent tray
				if t.parent != nil {
					// log.Infof("%s: did not find %#v checking parent %s", t.Path(), keys, t.parent.Path())
					return t.parent.Get(keys...)
				}
				// TODO: also check default configs
				log.Errorf("%s: Get did not find key '%s'", t.Path(), key)
				return nil
			}
			continue
		default:
			// log.Infof("%s: has config=%#v", t.Path(), t.config)
			log.Errorf("%s: Get unknown type '%#v'", t.Path(), data)
			return nil
		}
	}
	// log.Infof("%s: found key %#v data=%#v", t.Path(), keys, data)
	return data
}

func (t *trayDir) loadConfig() {
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

func (t *trayDir) Name() string {
	return t.info.Name()
}

func (t *trayDir) Path() string {
	return t.path
}

func (t *trayDir) Pass(name string) (Tray, error) {
	for _, dir := range t.dirs {
		if dir.Name() == name {
			return dir, nil
		}
	}
	return nil, fmt.Errorf("Directory not found '%s'", name)
}

func (t *trayDir) Glass(name string) (Glass, error) {
	var match Glass
	best := 0
	for _, drink := range t.drinks {
		prio := drink.Match(name)
		// log.Infof("Glass search %s in drink %s has %d", name, drink.Name(), prio)
		if prio <= 0 {
			continue
		}
		if best <= 0 || prio < best {
			match = drink
			best = prio
		}
	}
	if match != nil {
		return match, nil
	}
	return nil, fmt.Errorf("File '%s' not found in tray %s", name, t.Path())
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
	tray.parent = parent

	tray.loadConfig()

	glasses := []Glass{}
	files, rdErr := ioutil.ReadDir(path)
	if rdErr != nil {
		// log the error
		return nil
	}
	dirs := []string{}
	drinks := []Glass{}
	for _, file := range files {
		// log.Infof("Processing file %s/%s", tray.Path(), file.Name())
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		} else {
			prio, wiskName := tray.FileMatch(file.Name())
			glassfile := Blow(tray, &file, prio)
			glasses = append(glasses, glassfile)

			if prio > 0 {
				wisk, exists := Wisks[wiskName]
				if exists {
					drinks = append(drinks, wisk(glassfile))
					// log.Infof("File %s/%s wisk %s added to drinks", tray.Path(), file.Name(), wiskName)
				} else {
					// log.Infof("File %s/%s has unknown wisk %s", tray.Path(), file.Name(), wiskName)
				}
			} else {
				// log.Infof("File %s/%s has no match", tray.Path(), file.Name())
			}
		}
	}
	tray.files = glasses
	tray.drinks = drinks

	tray.dirs = []Tray{}
	for _, dir := range dirs {
		tray.dirs = append(tray.dirs, Cast(filepath.Join(tray.path, dir), tray))
	}

	return tray
}
