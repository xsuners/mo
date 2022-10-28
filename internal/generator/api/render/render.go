package render

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/xsuners/mo/internal/generator/api/spec"
	v1 "github.com/xsuners/mo/internal/generator/api/template/v1"
	"github.com/xsuners/mo/internal/generator/api/template/v1/aggregation"
	"github.com/xsuners/mo/internal/generator/api/template/v1/command"
	"github.com/xsuners/mo/internal/generator/api/template/v1/event"
	"github.com/xsuners/mo/internal/generator/api/template/v1/repository"
	"github.com/xsuners/mo/internal/generator/api/template/v1/types"
)

type Generator interface {
	Generate(s *spec.Spec) error
}

func New(opt ...Option) Generator {
	g := &generator{
		opts: defaultOptions,
	}
	for _, o := range opt {
		o(&g.opts)
	}
	g.gens = append(g.gens,
		newAgg(),
		newRep(),
		newEvt(),
		newTyp(),
		newRea(),
		newWri(),
		newCmd(),
	)
	return g
}

type generator struct {
	opts Options

	gens []generater
}

var _ Generator = (*generator)(nil)

func (g *generator) Generate(s *spec.Spec) error {
	var err error
	for _, w := range g.gens {
		if err = w.gen(s); err != nil {
			return err
		}
	}
	return nil
}

type generater interface {
	gen(s *spec.Spec) error
}

type base struct {
	folder string
	name   string
	tpl    *template.Template
	check  func(s *spec.Spec) bool
}

func newAgg() generater {
	tpl, err := template.New("aggregation.proto").Parse(aggregation.Tpl)
	if err != nil {
		panic(err)
	}
	return &base{
		name:   "aggregation",
		folder: "aggregation",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Aggregation
		},
	}
}

func newRep() generater {
	tpl, err := template.New("repository.proto").Parse(repository.Tpl)
	if err != nil {
		panic(err)
	}
	return &base{
		folder: "repository",
		name:   "repository",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Repository
		},
	}
}

func newEvt() generater {
	tpl, err := template.New("event.proto").Parse(event.Tpl)
	if err != nil {
		panic(err)
	}
	return &base{
		folder: "event",
		name:   "event",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Event
		},
	}
}

func newRea() generater {
	tpl, err := template.New("read.proto").Parse(v1.Read)
	if err != nil {
		panic(err)
	}
	return &base{
		folder: "",
		name:   "read",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Read
		},
	}
}

func newWri() generater {
	tpl, err := template.New("write.proto").Parse(v1.Write)
	if err != nil {
		panic(err)
	}
	return &base{
		folder: "",
		name:   "write",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Write
		},
	}
}

func newTyp() generater {
	tpl, err := template.New("types.proto").Parse(types.Tpl)
	if err != nil {
		panic(err)
	}
	return &base{
		folder: "types",
		name:   "types",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Types
		},
	}
}

func newCmd() generater {
	tpl, err := template.New("command.proto").Parse(command.Tpl)
	if err != nil {
		panic(err)
	}
	return &base{
		folder: "command",
		name:   "command",
		tpl:    tpl,
		check: func(s *spec.Spec) bool {
			return s.Command
		},
	}
}

func (b *base) gen(s *spec.Spec) error {
	if !b.check(s) {
		return nil
	}
	buf := bytes.NewBuffer([]byte{})
	err := b.tpl.Execute(buf, s)
	if err != nil {
		return err
	}
	err = os.MkdirAll(s.Dist+"/"+s.Path+"/v1/"+b.folder, 0777)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(s.Dist+"/"+s.Path+"/v1/"+b.folder+"/"+b.name+".proto", buf.Bytes(), 0666); err != nil {
		return err
	}
	return nil
}
