package xsql

import (
	"database/sql"
	"log"
	"time"
)

// Config .
type Config struct {
	DSN    string `json:"dsn"`
	Driver string `json:"driver"`
}

type options struct {
	connMaxLifetime time.Duration
	maxIdleConns    int
	maxOpenConns    int
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

// Database is database clienr
type Database struct {
	*sql.DB
	opts options
}

// New create a sql client .
func New(c *Config, opts ...Option) *Database {
	dopts := defaultOptions
	for _, o := range opts {
		o.apply(&dopts)
	}
	pool, err := sql.Open(c.Driver, c.DSN)
	if err != nil {
		log.Fatal("unable to use data source name", err)
	}
	pool.SetConnMaxLifetime(dopts.connMaxLifetime)
	pool.SetMaxIdleConns(dopts.maxIdleConns)
	pool.SetMaxOpenConns(dopts.maxOpenConns)
	db := &Database{}
	db.opts = dopts
	db.DB = pool
	return db
}

// Close close the connection.
func (db *Database) Close() {
	db.DB.Close()
}
