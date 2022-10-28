package render

import (
	"github.com/xsuners/mo/internal/generator/server/spec"
)

type Render interface {
	Rend(s *spec.Spec) error
}
