package description

// Options .
type Options interface {
	Value() interface{}
}

// A CallOption sets options such as credentials, codec and keepalive parameters, etc.
type CallOption interface {
	Apply(Options)
}

// FuncOption .
type FuncOption struct {
	f func(Options)
}

// Apply .
func (fdo *FuncOption) Apply(do Options) {
	fdo.f(do)
}

// NewFuncOption .
func NewFuncOption(f func(Options)) *FuncOption {
	return &FuncOption{
		f: f,
	}
}
