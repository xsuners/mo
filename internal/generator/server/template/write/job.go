package write

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

type jtmpl struct {
	tpl  *xtpl.Template
	name string
}

func Job() template.Templater {
	t := &jtmpl{
		name: "write.job",
	}
	tpl, err := xtpl.New(t.Name()).Parse(`tpl`)
	if err != nil {
		panic(err)
	}
	t.tpl = tpl
	return t
}

func (t *jtmpl) Name() string {
	return t.name
}

func (t *jtmpl) Execute(data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
