package main

import (
	"fmt"
	"strings"
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

// perform replacement of elements with tags
func (d *Doc) Replace(t *Tag, data interface{}) error {
	found := d.glass.Get("elements", t.Name)
	if found == nil {
		return nil
	}
	fmt.Printf("Replacing %s with %#v\n", t.Name, found)
	NewTag := Tag{Contents: t.Contents}
	brErr := d.ReplaceTag(t, &NewTag, found, data)
	if brErr != nil {
		fmt.Printf("Replace error: %v\n", brErr)
		ErrTag := Tag{Name: "error", Contents: brErr.Error()}
		rErr := d.Replace(&ErrTag, nil)
		if rErr != nil {
			// well this is embarrasing, error processing error
			return rErr
			*t = ErrTag
			return nil
		}
		return nil
	}
	*t = NewTag
	fmt.Printf("Replace %v result: \n%s\n", t.Name, asYaml(t))
	return nil
}

func (d *Doc) ReplaceTag(old *Tag, newtag *Tag, element interface{}, data interface{}) error {
	fmt.Printf("RT element: %#v\n", element)
	switch e := element.(type) {
	case []interface{}:
		// array contains raw content (no name change)
		/*
			newtag.Name = old.Name
			c := []interface{}{}
			for _, i := range e {
				switch f := i.(type) {
				case string:
					sub := Tag{Name: f}
					fmt.Printf("Calling replace on %#v from %#v\n", sub, i)
					rErr := d.Replace(&sub, data)
					if rErr != nil {
						return rErr
					}
					c = append(c, sub)
				default:
					return fmt.Errorf("ReplaceTag: Unexpected []interface type %#v", f)
				}
			}
			newtag.Contents = c
			return nil
		*/
		newtag.Name = old.Name
		newtag.Params = old.Params
		newtag.Contents = e
		return nil
	case map[string]interface{}:
		// map may/should contain tag-like members
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
		return fmt.Errorf("Element replacement failed from %#v", e)

	default:
		return fmt.Errorf("Don't know how to replace %v with element %#v", old.Name, element)
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
/*
func (t *Tag) String() string {
	return t.Convert(&NoDoc)
}
*/

func (t *Tag) Convert(doc *Doc) (string, error) {
	doc.Replace(t, nil)
	fmt.Printf("running Convert on %#v\n", t)
	Contents := t.Contents
	if Contents == nil {
		Contents = ""
	}
	switch content := Contents.(type) {
	case map[string]interface{}:
		set := []string{}
		for name, element := range content {
			te := Tag{Name: name, Contents: element}
			fmt.Printf("Calling convert on string tag %v", te.Name)
			s, cErr := te.Convert(doc)
			if cErr != nil {
				return "", cErr
			}
			set = append(set, s)
		}
		t.Contents = set
		return t.Convert(doc)
	case []interface{}:
		set := []string{}
		for _, element := range content {
			te := Tag{Contents: element}
			fmt.Printf("Calling convert on empty tag element %v", element)
			s, cErr := te.Convert(doc)
			if cErr != nil {
				return "", cErr
			}
			set = append(set, s)
		}
		t.Contents = set
		return t.Convert(doc)

	case []Tag:
		set := []string{}
		for _, element := range content {
			fmt.Printf("Calling convert on tag element %v", element)
			s, cErr := element.Convert(doc)
			if cErr != nil {
				return "", cErr
			}
			set = append(set, s)
		}
		t.Contents = set
		return t.Convert(doc)
	case Tag:
		var cErr error
		fmt.Printf("Calling convert on tag %v contents", t.Name)
		t.Contents, cErr = content.Convert(doc)
		if cErr != nil {
			return "", cErr
		}
		return t.Convert(doc)
	case []string:
		t.Contents = strings.Join(content, "\n")
		fmt.Printf("Calling convert on []strings %v", t.Name)
		return t.Convert(doc)
	case string:
		return ConvertString(t, content), nil
	case nil:
		return ConvertString(t, ""), nil
	default:
		return "", fmt.Errorf("Don't know how to tag type %#v", t.Contents)
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
			fmt.Printf("ConvertString calling Replace\n")
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

func (t *Tag) Find(name string) *Tag {
	switch content := t.Contents.(type) {
	case []Tag:
		for _, target := range content {
			if target.Name == name {
				return &target
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("tag.Find() unexpected type %#v", t.Contents))
	}
}
