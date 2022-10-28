package generator

import (
	"github.com/xsuners/mo/internal/generator/api"
	"github.com/xsuners/mo/internal/generator/server"
)

type Command struct {
	Server server.Command `command:"server"`
	Api    api.Command    `command:"api"`
}
