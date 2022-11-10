package kube

import (
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"go.uber.org/zap"
)

type Naming struct{}

var _ naming.Naming = (*Naming)(nil)

func New() naming.Naming {
	return &Naming{}
}

func (n *Naming) Deregister() {
	log.Infos("Deregister")
}

func (n *Naming) Register(svc *naming.Service) (err error) {
	log.Infos("Register", zap.Any("svc", svc))
	return
}
