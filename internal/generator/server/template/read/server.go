package read

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

var tpl = ``

type tmpl struct {
	tpl  *xtpl.Template
	name string
}

func New() template.Templater {
	t := &tmpl{
		name: "read.server",
	}
	tpl, err := xtpl.New(t.Name()).Parse(tpl)
	if err != nil {
		panic(err)
	}
	t.tpl = tpl
	return t
}

func (t *tmpl) Name() string {
	return t.name
}

func (t *tmpl) Execute(data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
