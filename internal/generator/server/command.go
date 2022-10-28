package server

import (
	"strings"

	"github.com/xsuners/mo/internal/generator/out/file"
	"github.com/xsuners/mo/internal/generator/server/render/impl"
	"github.com/xsuners/mo/internal/generator/server/spec"
	"github.com/xsuners/mo/internal/generator/server/template/aggregation"
	"github.com/xsuners/mo/internal/generator/server/template/read"
	"github.com/xsuners/mo/internal/generator/server/template/repository"
	"github.com/xsuners/mo/internal/generator/server/template/write"
)

type Command struct {
	Dist  string `short:"d" long:"dist" description:"Dist dir path" default:"./"`
	Proto string `short:"p" long:"proto" description:"Proto dir path" default:"./"`
}

func (cmd *Command) Execute(args []string) error {
	s, err := spec.Load(cmd.Proto)
	if err != nil {
		return err
	}
	s.Dist = strings.TrimSuffix(cmd.Dist, "/")

	o := file.New()
	i := impl.New(o,
		// aggregation tmpls
		aggregation.New(),
		// repository tmpls
		repository.New(),
		// write tmpls
		write.New(),
		write.Job(),
		write.Cmd(),
		write.Sub(),
		write.Oth(),
		// read tmpls
		read.New(),
		read.Inq(),
		read.Agg(),
		read.Oth(),
	)

	err = i.Rend(s)
	if err != nil {
		return err
	}

	return nil
}
