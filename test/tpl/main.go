package main

import (
	"bytes"
	"fmt"
	"text/template"
)

func main() {
	templ := template.Must(template.New("").ParseGlob("../../internal/generator/template/v1/*.proto"))
	if true {
		buf := bytes.NewBuffer([]byte{})
		err := templ.ExecuteTemplate(buf, "read.proto", "")
		if err != nil {
			panic(err)
		}
		fmt.Println(buf.String())
	}
	// templ.Execute(w, r.Data)
	// return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

// func test() {
// 	config

// 	p := parser.New(config)
// 	s := p.Parse()

// 	o := output.New()
// 	r := rander.New(o)
// 	r.Render(s)
// }
