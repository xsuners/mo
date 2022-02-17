package description

import "github.com/xsuners/mo/naming"

type Server interface {
	Register(ss interface{}, sds ...*ServiceDesc)
	Serve() error
	Naming(naming naming.Naming) error
}
