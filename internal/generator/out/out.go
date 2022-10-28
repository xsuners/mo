package out

type Outer interface {
	Out(path string, data []byte) error
}
