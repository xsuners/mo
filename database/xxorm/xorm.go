package xxorm

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/xsuners/mo/database"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

type Engine struct {
	*xorm.Engine
	opts *Options
}

func dsn(opts Options) string {
	return opts.Username + ":" + opts.Password + "@tcp(" + opts.IP + ":" + fmt.Sprintf("%d", opts.Port) + ")/" + opts.Name + "?charset=utf8&loc=Local&parseTime=True"
}

// New .
func New(opt ...Option) (*Engine, func(), error) {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	engine, err := xorm.NewEngine(opts.Driver, dsn(opts))
	if err != nil {
		panic(err)
	}

	if opts.MaxIdleConns > 0 {
		engine.SetMaxIdleConns(opts.MaxIdleConns)
	}
	if opts.MaxOpenConns > 0 {
		engine.SetMaxOpenConns(opts.MaxOpenConns)
	}

	r := &Engine{
		Engine: engine,
		opts:   &opts,
	}
	return r, func() {
		engine.Close()
	}, nil
}

var (
	_engine     *Engine
	_engineOnce sync.Once
)

// Singleton .
func Singleton(opt ...Option) (*Engine, func(), error) {

	_engineOnce.Do(func() {
		opts := defaultOptions
		for _, o := range opt {
			o(&opts)
		}
		engine, err := xorm.NewEngine(opts.Driver, dsn(opts))
		if err != nil {
			panic(err)
		}

		if opts.MaxIdleConns > 0 {
			engine.SetMaxIdleConns(opts.MaxIdleConns)
		}
		if opts.MaxOpenConns > 0 {
			engine.SetMaxOpenConns(opts.MaxOpenConns)
		}

		_engine = &Engine{
			Engine: engine,
			opts:   &opts,
		}
	})
	return _engine, func() {
		_engine.Close()
	}, nil
}

// type Session struct {
// 	*xorm.Session
// }

func (e *Engine) Session(ctx context.Context, qs ...database.Q) *xorm.Session {
	s := e.Engine.Context(ctx)
	for _, q := range qs {
		q(func(c *database.Condition) {
			switch c.Op {
			case database.Operation_eq:
				s.Where(c.Key+" = ?", c.Value)
			case database.Operation_gt:
				s.Where(c.Key+" > ?", c.Value)
			case database.Operation_lt:
				s.Where(c.Key+" < ?", c.Value)
			case database.Operation_in:
				s.In(c.Key, c.Value)
			}
		})
	}
	return s
}

func (e *Engine) Tx(ctx context.Context, fns ...func(tx *xorm.Session) error) (err error) {
	tx := e.Engine.Context(ctx)
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return
	}
	for _, fn := range fns {
		if err = fn(tx); err != nil {
			tx.Rollback()
			return
		}
	}
	if err = tx.Commit(); err != nil {
		return
	}
	return
}

type Tabler interface {
	bean() any
	name() string
}

type table struct {
	obj any
	tbl string
}

var _ Tabler = (*table)(nil)

func (t *table) bean() any {
	return t.obj
}
func (t *table) name() string {
	return t.tbl
}

func Table(bean any, name ...string) Tabler {
	t := &table{
		obj: bean,
	}
	if len(name) > 0 {
		t.tbl = name[0]
	}
	return t
}

func (e *Engine) Synchronize(namespace string, ts ...Tabler) error {
	var kvs []string
	var beans []any
	for _, m := range ts {
		if m.name() != "" {
			// fmt.Println(reflect.TypeOf(Table{}).Name())
			// fmt.Println(reflect.ValueOf(&Table{}).Type().Kind() == reflect.Ptr)
			// fmt.Println(reflect.ValueOf(&Table{}).Elem().Type().Name())
			obj := reflect.TypeOf(m.bean()).Name()
			if reflect.ValueOf(m.bean()).Type().Kind() == reflect.Ptr {
				obj = reflect.ValueOf(m.bean()).Elem().Type().Name()
			}
			kvs = append(kvs, obj, m.name())
		}
		beans = append(beans, m.bean())
	}
	e.SetTableMapper(newXmapper(namespace, kvs...))
	if err := e.Engine.Sync2(beans...); err != nil {
		return err
	}
	return nil
}

type xmapper struct {
	mapper    names.Mapper
	namespace string
	mm        map[string]string
}

var _ names.Mapper = (*xmapper)(nil)

var m = &xmapper{
	mapper: names.GonicMapper{},
	mm:     make(map[string]string),
}

func newXmapper(namespace string, kvs ...string) names.Mapper {
	// m := &xmapper{
	// 	mapper:    names.GonicMapper{},
	// 	mm:        make(map[string]string),
	// 	namespace: namespace,
	// }
	if m.namespace != "" {
		m.namespace += "_"
	}
	if len(kvs)%2 != 0 {
		panic("kvs num not even")
	}
	for i := 0; i < len(kvs); i += 2 {
		m.mm[kvs[i]] = kvs[i+1]
		m.mm[kvs[i+1]] = kvs[i]
	}
	return m
}

func (m *xmapper) Obj2Table(name string) string {
	n, ok := m.mm[name]
	if ok {
		return m.namespace + m.mapper.Obj2Table(n)
	}
	return m.namespace + m.mapper.Obj2Table(name)
}

func (m *xmapper) Table2Obj(name string) string {
	n, ok := m.mm[name[len(m.namespace):]]
	if ok {
		return m.mapper.Table2Obj(n)
	}
	return m.mapper.Table2Obj(name[len(m.namespace):])
}
