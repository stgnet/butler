package main

import (
	"fmt"
	"io"
)

type Doc struct {
	glass Glass // glass for obtaining content rules
	head  []Tag // content to add to end of head
	body  []Tag // content to add to end of body
}

// wrap content on the page with html head/body/etc according to element rules
func (d *Doc) DistillHtml(content interface{}) Tag {
	// get html tag

	return Tag{}
}

// distill data records into tag structure via element rules
func (d *Doc) DistillData(name string, records interface{}) Tag {
	return Tag{}
}

func Render(w io.Writer, g Glass) error {
	// build html from glass and write it

	// get the data to display for this glass
	data := g.Data(nil)
	fmt.Fprintf(w, "data: %#v", data)

	doc := Doc{glass: g}
	doc.DistillData("name", nil)
	// for each data as name -> record
	// dataTag := doc.DistillData(name, record)
	return nil
}
