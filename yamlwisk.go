package main

import (
	// "fmt"
	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

type glassYaml struct {
	source Glass
	butler interface{} // butler: section of yaml
	data   interface{} // everything else
}

func (g *glassYaml) Name() string {
	return g.source.Name()
}
func (g *glassYaml) Path() string {
	return g.source.Path()
}
func (g *glassYaml) Tray() Tray {
	return g.source.Tray()
}
func (g *glassYaml) Type() string {
	return "text/html"
}

/*
func (w *writeYaml) Error(c string, e error) {
	w.out.Write([]byte(fmt.Sprintf("[an error occurred while processing %s: %v\n", c, e)))
	log.Errorf("Yaml error %s: %v", c, e)
}
*/

func (g *glassYaml) Pour(w io.Writer) error {
	return Render(w, g)
}

func (g *glassYaml) Match(name string) int {
	return g.source.Match(name)
}

func (g *glassYaml) Get(keys ...string) interface{} {
	// TODO: check this yaml for butler section
	return g.Tray().Get(keys...)
}

func (g *glassYaml) Data(keys url.Values) interface{} {
	// TODO: get the data from yaml and remove butler section
	file, fErr := os.Open(filepath.Join(g.Tray().Path(), g.Name()))
	if fErr != nil {
		log.Errorf("Error opening %s: %v", g.Name(), fErr)
		return nil
	}
	defer file.Close()
	var data interface{}
	dErr := yaml.NewDecoder(file).Decode(&data)
	if dErr != nil {
		log.Errorf("Decode error %s: %v", g.Name(), dErr)
	}
	return data
}

func YamlWisk(source Glass) Glass {
	glass := new(glassYaml)
	glass.source = source
	return glass
}
