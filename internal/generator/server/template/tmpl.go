package template

type Templater interface {
	Name() string
	Execute(data any) ([]byte, error)
}
