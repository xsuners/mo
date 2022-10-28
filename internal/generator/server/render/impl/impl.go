package impl

import (
	"fmt"

	"github.com/xsuners/mo/internal/generator/out"
	"github.com/xsuners/mo/internal/generator/server/render"
	"github.com/xsuners/mo/internal/generator/server/spec"
	"github.com/xsuners/mo/internal/generator/server/template"
)

type impl struct {
	out   out.Outer
	tmpls map[string]template.Templater
}

var _ render.Render = (*impl)(nil)

func New(out out.Outer, tmpls ...template.Templater) render.Render {
	i := &impl{
		out:   out,
		tmpls: make(map[string]template.Templater),
	}
	for _, t := range tmpls {
		i.tmpls[t.Name()] = t
	}
	return i
}

func (i *impl) Rend(sp *spec.Spec) error {
	type out struct {
		tmpl string
		path string
		data any
	}
	var outs []*out
	if svc := sp.Aggregation; svc != nil {
		if rep := svc.Repository; rep != nil {
			outs = append(outs, &out{tmpl: rep.Template, path: sp.Dist + "/" + rep.Path, data: rep})
		}
	}
	if svr := sp.Repository; svr != nil {
		if svc := svr.Repository; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
	}
	if svr := sp.Write; svr != nil {
		if svc := svr.Server; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		if svc := svr.Request; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		if svc := svr.Command; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		if svc := svr.Sub; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		if svc := svr.Job; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		for _, svc := range svr.Others {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
	}
	if svr := sp.Read; svr != nil {
		if svc := svr.Server; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		if svc := svr.Inquiry; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		if svc := svr.Aggregate; svc != nil {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
		for _, svc := range svr.Others {
			outs = append(outs, &out{tmpl: svc.Template, path: sp.Dist + "/" + svc.Path, data: svc})
		}
	}
	for _, o := range outs {
		if err := i.rend(o.tmpl, o.path, o.data); err != nil {
			return err
		}
	}
	return nil
}

func (i *impl) rend(tmpl, path string, data any) error {
	t, ok := i.tmpls[tmpl]
	if !ok {
		return fmt.Errorf("tmpl: %s not found", tmpl)
	}
	out, err := t.Execute(data)
	if err != nil {
		return err
	}
	return i.out.Out(path, out)
}
