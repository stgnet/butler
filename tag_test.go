package main

import (
	"testing"
)

func OFF_TestTagConvert(t *testing.T) {
	t.Log("testing tag convert")

	glass := &glasstest{t: t}
	doc := &Doc{glass: glass}
	html := &Tag{Name: "html"}

	t.Log("Calling convert on html")
	out, cErr := html.Convert(doc)
	if cErr != nil {
		t.Fatalf("Error convert: %v", cErr)
	}
	t.Logf("Convert output: %s", out)
}

func OFF_TestTagReplace(t *testing.T) {
	t.Log("testing Tag Replace")
	glass := &glasstest{t: t}

	doc := &Doc{glass: glass}

	html := &Tag{Name: "html"}

	rErr := doc.Replace(html, nil)
	if rErr != nil {
		t.Fatalf("TagReplace failed: %v", rErr)
	}

	t.Logf("Result after Tag Replace: %#v", html)
	/*
		text, cErr := html.Convert(doc)
		if cErr != nil {
			t.Errorf("html convert resulted in error: %#v", cErr)
		}
		t.Logf("As html: %s", text)
	*/
}

func OFF_TestTagFind(t *testing.T) {
	t.Log("testing Tag")

	tag := Tag{
		Name: "one",
		Contents: []Tag{
			Tag{
				Name:     "two",
				Contents: "end",
			},
		},
	}

	result := tag.Find("two")

	if result == nil {
		t.Fatalf("Tag.find failed")
	}
	if result.Name != "two" {
		t.Fatalf("Tag.find didn't find two, found: %#v", result)
	}

	result2 := tag.Find("three")
	if result2 != nil {
		t.Fatalf("Tag.find unexpected found %#v", result2)
	}
}
