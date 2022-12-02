package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
	// "io"
	"io/ioutil"
)

func Walk(data interface{}, keys ...string) interface{} {
	for _, key := range keys {
		switch d := data.(type) {
		case map[string]interface{}:
			data := d[key]
			if data == nil {
				log.Errorf("Walk did not find key '%s'", key)
				return nil
			}
			continue
		default:
			log.Errorf("Walk unknown type '%#v'", data)
			return nil
		}
	}
	return data
}

func getYaml(file string) interface{} {
	products, pErr := ioutil.ReadFile(file)
	if pErr != nil {
		panic(fmt.Errorf("Cant read '%s': %v", file, pErr))
	}

	var data interface{}

	yErr := yaml.Unmarshal(products, &data)
	if yErr != nil {
		panic(fmt.Errorf("Cant understand yaml '%s': %v", file, yErr))
	}
	return data
}

func asYaml(d interface{}) string {
	out, err := yaml.Marshal(d)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

/*
func renderYaml(w io.Writer, file string) {
	els := getYaml("elements.yaml")

	// elements, eExists := (els.(map[string]interface{}))["elements"]
	elements := Walk(els, "elements")

	doc := Doc{Elements: elements.(map[string]interface{})}

	data := getYaml(file)

	json, jErr := json.MarshalIndent(els, "", "  ")
	if jErr != nil {
		log.Fatal(jErr)
	}
	fmt.Println(string(json))
	fmt.Printf("data: %#v\n", data)

	html := Walk(data, "html") // data.GetAny("html") // (map[string]interface{}))["html"]
	/*
		if !htmlExists {
			log.Fatal("no html section")
		}
	tag := Tag{
		Name:     "html",
		Params:   Params{"lang": "en"},
		Contents: html,
	}
	io.WriteString(w, "<!DOCTYPE html>\n"+tag.Convert(&doc))
}
*/
