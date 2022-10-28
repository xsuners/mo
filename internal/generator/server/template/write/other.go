package write

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

type otmpl struct {
	tpl  *xtpl.Template
	name string
}

func Oth() template.Templater {
	t := &otmpl{
		name: "write.other",
	}
	tpl, err := xtpl.New(t.Name()).Parse(`tpl`)
	if err != nil {
		panic(err)
	}
	t.tpl = tpl
	return t
}

func (t *otmpl) Name() string {
	return t.name
}

func (t *otmpl) Execute(data any) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := t.tpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
