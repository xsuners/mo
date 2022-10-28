package repository

import (
	"bytes"
	xtpl "text/template"

	"github.com/xsuners/mo/internal/generator/server/template"
)

var tpl = `import (
	"context"

	"github.com/xsuners/api_go/im/user_message/v1/repository"
	"github.com/xsuners/mo/database/xsql"
)

type server struct {
	sql xsql.SQL
}

func New(sql xsql.SQL) repository.RepositoryServer {
	return &server{
		sql: sql,
	}
}

func (s *server) Create(ctx context.Context, in *repository.Message) (*repository.Message, error) {
	return nil, nil
}

func (s *server) Update(ctx context.Context, in *repository.Option) (*repository.Message, error) {
	return nil, nil
}

func (s *server) Delete(ctx context.Context, in *repository.Query) (*repository.Message, error) {
	return nil, nil
}

func (s *server) Get(ctx context.Context, in *repository.Query) (*repository.Message, error) {
	return nil, nil
}

func (s *server) List(ctx context.Context, in *repository.Query) (*repository.Messages, error) {
	return nil, nil
}
`

type tmpl struct {
	tpl  *xtpl.Template
	name string
}

func New() template.Templater {
	t := &tmpl{
		name: "repository",
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
