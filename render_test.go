package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

type glasstest struct {
	t *testing.T
}

func (g *glasstest) Data(keys url.Values) interface{} {
	records, rErr := ioutil.ReadFile("render_test_data.yaml")
	if rErr != nil {
		g.t.Errorf("ReadFile render_test_data.yaml: %v", rErr)
		return nil
	}
	var data interface{}
	yErr := yaml.Unmarshal(records, &data)
	if yErr != nil {
		g.t.Errorf("yaml.Unmarshal render_test_data: %v", yErr)
		return nil
	}
	return data
}

func (g *glasstest) Get(keys ...string) interface{} {
	records, rErr := ioutil.ReadFile("render_test_elem.yaml")
	if rErr != nil {
		g.t.Errorf("ReadFile render_test_elem.yaml: %v", rErr)
		return nil
	}
	var data interface{}
	yErr := yaml.Unmarshal(records, &data)
	if yErr != nil {
		g.t.Errorf("yaml.Unmarshal render_test_elem: %v", yErr)
		return nil
	}

	if data == nil {
		return nil
	}
	for _, key := range keys {
		switch d := data.(type) {
		case map[string]interface{}:
			data = d[key]
			if data == nil {
				g.t.Logf("%s: Get did not find key '%s'", "render_test_elem", key)
				return nil
			}
			continue
		default:
			g.t.Errorf("%s: Get unknown type '%#v'", "render_test_elem", data)
			return nil
		}
	}
	// log.Infof("%s: found key %#v data=%#v", t.Path(), keys, data)
	return data

}

func (g *glasstest) Match(match string) int {
	return 0
}

func (g *glasstest) Name() string {
	return "test"
}
func (g *glasstest) Path() string {
	return "none"
}
func (g *glasstest) Pour(w io.Writer) error {
	return fmt.Errorf("pour test not implemented")
}
func (g *glasstest) Tray() Tray {
	return nil
}

func (g *glasstest) Type() string {
	return "text/html"
}

func TestRender(t *testing.T) {
	t.Log("testing Render")

	glass := &glasstest{t: t}

	rErr := Render(os.Stdout, glass)

	if rErr != nil {
		t.Fatalf("Render failed: %v", rErr)
	}
}
