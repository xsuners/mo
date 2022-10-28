package write

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

type rtmpl struct {
	tpl  *xtpl.Template
	name string
}

func Req() template.Templater {
	t := &rtmpl{
		name: "write.request",
	}
	tpl, err := xtpl.New(t.Name()).Parse(`tpl`)
	if err != nil {
		panic(err)
	}
	t.tpl = tpl
	return t
}

func (t *rtmpl) Name() string {
	return t.name
}

func (t *rtmpl) Execute(data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
