package main

import (
	"fmt"
	"io/ioutil"
	"log"
	// "os"
	"io"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	jsoniter "github.com/json-iterator/go"
	// "gopkg.in/yaml.v2"
)

var indent = "  "

var id_counter = 0

func getId(prefix string) string {
	id_counter += 1
	return fmt.Sprintf("%s%d", prefix, id_counter)
}

var dont_break = []string{
	"li",
	"a",
}

var dont_close = []string{
	"area",
	"base",
	"br",
	"col",
	"command",
	"embed",
	"hr",
	"img",
	"input",
	"keygen",
	"link",
	"meta",
	"param",
	"source",
	"track",
	"wbr",
}

var dont_self_close = []string{
	"script",
	"i",
	"iframe",
	"div",
	"span",
	"title",
}

func isIn(value string, list []string) bool {
	for _, entry := range list {
		if entry == value {
			return true
		}
	}
	return false
}

type Params map[string]interface{}

func NewClass(classes []string) Params {
	return Params{"class": classes}
}

type Tag struct {
	Name     string
	Params   Params
	Contents interface{}
}

type Doc struct {
	Elements map[string]interface{}
}

var NoDoc = Doc{}

// perform replacement of elements with tags
func (d *Doc) Replace(t *Tag) error {
	found, exists := d.Elements[t.Name]
	if !exists {
		return nil
	}
	fmt.Printf("Replacing %s with %#v\n", t.Name, found)
	NewTag := Tag{}
	brErr := ReplaceTag(t, &NewTag, found)
	if brErr != nil {
		fmt.Printf("Replace error: %v\n", brErr)
		ErrTag := Tag{Name: "error", Contents: brErr.Error()}
		rErr := d.Replace(&ErrTag)
		if rErr != nil {
			// well this is embarrasing, error processing error
			return rErr
			*t = ErrTag
			return nil
		}
		return nil
	}
	*t = NewTag
	fmt.Printf("Replace result: %#v\n", t)
	return nil
}

func ReplaceTag(old *Tag, newtag *Tag, element interface{}) error {
	//fmt.Printf("element: %#v\n", element)
	switch e := element.(type) {
	case map[string]interface{}:
		name, nameExists := e["name"]
		if !nameExists {
			name, nameExists = e["tag"]
		}
		params, paramsExists := e["params"]
		if paramsExists {
			newtag.Params = params.(map[string]interface{})
		}
		contents, contentsExists := e["contents"]
		if contentsExists {
			newtag.Contents = ReplaceVar(contents, old.Contents)
		}
		if nameExists {
			(*newtag).Name = name.(string)
			return nil
		}
		return fmt.Errorf("Element repalcement failed from %#v", e)

	default:
		return fmt.Errorf("Don't know how to replace with element %#v", element)
	}
}

func ReplaceVar(template interface{}, data interface{}) interface{} {
	switch t := template.(type) {
	case string:
		return ReplaceStringVar(t, data)
	default:
		return fmt.Sprintf("Error: unknown template %#v", template)
	}
}

func ReplaceStringVar(t string, data interface{}) interface{} {
	if t == "$" {
		return data
	}
	if t[0] == '$' {
		value, valueExists := (data.(map[string]interface{}))[t[1:]]
		if valueExists {
			return value
		}
	}
	return t
}

// smartly generate neatly formatted nested tags
func (t *Tag) String() string {
	return t.Convert(&NoDoc)
}

func (t *Tag) Convert(doc *Doc) string {
	doc.Replace(t)
	// fmt.Printf("    %#v\n", t)
	Contents := t.Contents
	if Contents == nil {
		Contents = ""
	}
	switch content := Contents.(type) {
	case map[string]interface{}:
		set := []string{}
		for name, element := range content {
			te := Tag{Name: name, Contents: element}
			set = append(set, te.Convert(doc))
		}
		t.Contents = set
		return t.Convert(doc)
	case []interface{}:
		set := []string{}
		for _, element := range content {
			te := Tag{Contents: element}
			set = append(set, te.Convert(doc))
		}
		t.Contents = set
		return t.Convert(doc)

	case []Tag:
		set := []string{}
		for _, element := range content {
			set = append(set, element.Convert(doc))
		}
		t.Contents = set
		return t.Convert(doc)
	case Tag:
		t.Contents = content.Convert(doc)
		return t.Convert(doc)
	case []string:
		t.Contents = strings.Join(content, "\n")
		return t.Convert(doc)
	case string:
		return ConvertString(t, content)
	case nil:
		return ConvertString(t, "")
	default:
		panic(fmt.Sprintf("Don't know how to tag type %#v", t.Contents))
	}
}

func ConvertString(t *Tag, content string) string {
	if len(t.Name) == 0 || t.Name == "-" || t.Name == "+" {
		return content
	}
	html := "<" + t.Name
	tag := strings.Split(t.Name, " ")[0]
	for name, value := range t.Params {
		html += " " + name + "=\"" + fmt.Sprintf("%v", value) + "\""
	}
	if len(content) == 0 && !isIn(tag, dont_self_close) {
		html += "/>"
		return html
	}
	html += ">"
	if len(content) > 0 {
		need_break := strings.Contains(content, "\n") || len(content) > 40
		if need_break && !isIn(tag, dont_break) {
			// note: this wil indent a <pre> in content -- to avoid:
			// process individual lines to add indent with <pre detection
			html += "\n" + indent + strings.Replace(content, "\n", "\n"+indent, -1) + "\n"
		} else {
			html += content
		}
	}
	if !isIn(tag, dont_close) {
		html += "</" + tag + ">"
	}
	return html
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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

func renderYaml(w io.Writer, file string) {
	els := getYaml("elements.yaml")

	elements, eExists := (els.(map[string]interface{}))["elements"]
	if !eExists {
		panic(fmt.Errorf("elements.yaml does not contain elements section"))
	}

	doc := Doc{Elements: elements.(map[string]interface{})}

	data := getYaml(file)

	json, jErr := json.MarshalIndent(els, "", "  ")
	if jErr != nil {
		log.Fatal(jErr)
	}
	fmt.Println(string(json))
	fmt.Printf("data: %#v\n", data)

	html, htmlExists := (data.(map[string]interface{}))["html"]
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

func renderFile(w io.Writer, file string) {
	ext := filepath.Ext(file)
	if ext == ".yaml" {
		renderYaml(w, file)
	} else {
		panic(fmt.Errorf("No method to handle %s", ext))
	}
}
