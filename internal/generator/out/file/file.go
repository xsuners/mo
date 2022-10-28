package file

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/xsuners/mo/internal/generator/out"
)

type file struct{}

var _ out.Outer = (*file)(nil)

func New() out.Outer {
	return &file{}
}

func (f *file) Out(path string, data []byte) error {
	ps := strings.Split(path, "/")
	fo := strings.Join(ps[:len(ps)-1], "/")
	err := os.MkdirAll(fo, 0777)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(path, data, 0666); err != nil {
		return err
	}
	return nil
}
