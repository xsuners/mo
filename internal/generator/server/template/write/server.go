package write

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

var tpl = `import (
	v1 "github.com/xsuners/api_go/im/user_message/v1"
)

type Server interface {
	v1.RequestServer
}

type server struct{}

var _ Server = (*server)(nil)

func New() Server {
	return &server{}
}
`

type tmpl struct {
	tpl  *xtpl.Template
	name string
}

func New() template.Templater {
	t := &tmpl{
		name: "write.server",
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
