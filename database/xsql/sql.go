package xsql

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type options struct {
	connMaxLifetime time.Duration
	maxIdleConns    int
	maxOpenConns    int
	dsn             string
	driver          string
}

var defaultOptions = options{
	maxIdleConns: 3,
	maxOpenConns: 3,
}

// Option .
type Option interface {
	apply(*options)
}

// EmptyOption .
type EmptyOption struct{}

func (EmptyOption) apply(*options) {}

type funcOption struct {
	f func(*options)
}

func (fdo *funcOption) apply(do *options) {
	fdo.f(do)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// ConnMaxLifetime .
func ConnMaxLifetime(d time.Duration) Option {
	return newFuncOption(func(o *options) {
		o.connMaxLifetime = d
	})
}

// MaxIdleConns .
func MaxIdleConns(num int) Option {
	return newFuncOption(func(o *options) {
		o.maxIdleConns = num
	})
}

// MaxOpenConns .
func MaxOpenConns(num int) Option {
	return newFuncOption(func(o *options) {
		o.maxOpenConns = num
	})
}

// DSN .
func DSN(dsn string) Option {
	return newFuncOption(func(o *options) {
		o.dsn = dsn
	})
}

// Driver .
func Driver(driver string) Option {
	return newFuncOption(func(o *options) {
		o.driver = driver
	})
}

// Database is database clienr
type Database struct {
	*sql.DB
	opts options
}

// New create a sql client .
func New(opts ...Option) (*Database, func(), error) {
	dopts := defaultOptions
	for _, o := range opts {
		o.apply(&dopts)
	}
	pool, err := sql.Open(dopts.driver, dopts.dsn)
	if err != nil {
		log.Fatal("unable to use data source name", err)
	}
	pool.SetConnMaxLifetime(dopts.connMaxLifetime)
	pool.SetMaxIdleConns(dopts.maxIdleConns)
	pool.SetMaxOpenConns(dopts.maxOpenConns)
	db := &Database{}
	db.opts = dopts
	db.DB = pool
	return db, func() {
		db.DB.Close()
	}, nil
}

func (db *Database) Exec(ctx context.Context, query string, args ...interface{}) (af, id int64, err error) {
	ret, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return
	}
	af, err = ret.RowsAffected()
	if err != nil {
		return
	}
	id, err = ret.LastInsertId()
	if err != nil {
		return
	}
	return
}
