package api

import (
	"fmt"
	"strings"

	"github.com/xsuners/mo/internal/generator/api/render"
	"github.com/xsuners/mo/internal/generator/api/spec"
)

type Command struct {
	Dist        string `short:"d" long:"dist" description:"Dist file path" default:"./"`
	Package     string `short:"p" long:"package" description:"Package"`
	Model       string `short:"m" long:"model" description:"Model name"`
	Plural      string `short:"M" long:"plural" description:"Model plural form"`
	Aggregation bool   `short:"a" long:"agg" description:"Whether generate aggregation"`
	Repository  bool   `short:"r" long:"repo" description:"Whether generate repository"`
	Event       bool   `short:"e" long:"event" description:"Whether generate event"`
	Read        bool   `short:"i" long:"read" description:"Whether generate read"`
	Write       bool   `short:"w" long:"write" description:"Whether generate write"`
	Types       bool   `short:"t" long:"types" description:"Whether generate types"`
	Command     bool   `short:"c" long:"command" description:"Whether generate command"`
	All         bool   `short:"A" long:"all" description:"Whether generate all"`
	Base        bool   `short:"B" long:"base" description:"Whether generate base"`
}

func (cmd *Command) Execute(args []string) error {
	spec := spec.Spec{
		Dist:        cmd.Dist,
		Package:     cmd.Package,
		Path:        strings.ReplaceAll(cmd.Package, ".", "/"),
		Model:       cmd.Model,
		Models:      cmd.Plural,
		Types:       cmd.All || cmd.Types,
		Event:       cmd.All || cmd.Event,
		Command:     cmd.All || cmd.Command,
		Aggregation: cmd.All || cmd.Aggregation,
		Repository:  cmd.All || cmd.Base || cmd.Repository,
		Read:        cmd.All || cmd.Base || cmd.Read,
		Write:       cmd.All || cmd.Base || cmd.Write,
	}
	if spec.Package == "" {
		fmt.Println("package can't be \"\"")
	}
	if spec.Model == "" {
		ps := strings.Split(spec.Package, ".")
		m := ps[len(ps)-1]
		spec.Model = upper(end(m))
	}
	if spec.Models == "" {
		spec.Models = spec.Model + "s"
	}

	g := render.New()
	if err := g.Generate(&spec); err != nil {
		panic(err)
	}
	return nil
}

func upper(in string) (out string) {
	ps := strings.Split(in, "_")
	for i, p := range ps {
		ps[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(ps, "")
}

func end(in string) string {
	ps := strings.Split(in, "_")
	return ps[len(ps)-1]
}
