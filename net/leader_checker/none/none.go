package none

import (
	"github.com/xsuners/mo/net/leader_checker"
)

type none struct{}

var _ leader_checker.Checker = (*none)(nil)

func (*none) IsLeader() bool {
	return false
}

func New() leader_checker.Checker {
	return &none{}
}
