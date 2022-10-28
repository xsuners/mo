package none

import (
	"github.com/xsuners/mo/naming"
)

type Naming struct{}

var _ naming.Naming = (*Naming)(nil)

func New() naming.Naming {
	return &Naming{}
}

func (n *Naming) Deregister() {}

func (n *Naming) Register(svc *naming.Service) (err error) {
	return
}
