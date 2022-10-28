package write

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

type subimpl struct {
	tpl  *xtpl.Template
	name string
}

func Sub() template.Templater {
	t := &subimpl{
		name: "write.request",
	}
	tpl, err := xtpl.New(t.Name()).Parse(`tpl`)
	if err != nil {
		panic(err)
	}
	t.tpl = tpl
	return t
}

func (t *subimpl) Name() string {
	return t.name
}

func (t *subimpl) Execute(data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
