package read

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

type atmpl struct {
	tpl  *xtpl.Template
	name string
}

func Agg() template.Templater {
	t := &atmpl{
		name: "read.aggregate",
	}
	tpl, err := xtpl.New(t.Name()).Parse(`tpl`)
	if err != nil {
		panic(err)
	}
	t.tpl = tpl
	return t
}

func (t *atmpl) Name() string {
	return t.name
}

func (t *atmpl) Execute(data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
