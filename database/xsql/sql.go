package xsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type options struct {
	connMaxLifetime time.Duration
	maxIdleConns    int
	maxOpenConns    int
	username        string
	password        string
	name            string
	ip              string
	port            int
	driver          string
}

var defaultOptions = options{
	username:     "root",
	password:     "123456",
	ip:           "127.0.0.1",
	driver:       "mysql",
	port:         3306,
	maxIdleConns: 3,
	maxOpenConns: 3,
}

// Option .
type Option func(o *options)

// ConnMaxLifetime .
func ConnMaxLifetime(d time.Duration) Option {
	return func(o *options) {
		o.connMaxLifetime = d
	}
}

// MaxIdleConns .
func MaxIdleConns(num int) Option {
	return func(o *options) {
		o.maxIdleConns = num
	}
}

// MaxOpenConns .
func MaxOpenConns(num int) Option {
	return func(o *options) {
		o.maxOpenConns = num
	}
}

// IP .
func IP(ip string) Option {
	return func(o *options) {
		o.ip = ip
	}
}

// Port .
func Port(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// Username .
func Username(username string) Option {
	return func(o *options) {
		o.username = username
	}
}

// Password .
func Password(password string) Option {
	return func(o *options) {
		o.password = password
	}
}

// Name .
func Name(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// Driver .
func Driver(driver string) Option {
	return func(o *options) {
		o.driver = driver
	}
}

// Database is database clienr
type Database struct {
	*sql.DB
	opts options
	dsn  string
}

// New create a sql client .
func New(opts ...Option) (db *Database, cf func(), err error) {
	db = &Database{
		opts: defaultOptions,
	}
	for _, o := range opts {
		o(&db.opts)
	}
	db.dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true",
		db.opts.username,
		db.opts.password,
		db.opts.ip,
		db.opts.port,
		db.opts.name,
	)
	db.DB, err = sql.Open(db.opts.driver, db.dsn)
	if err != nil {
		log.Fatal("unable to use data source name", err)
	}
	db.DB.SetConnMaxLifetime(db.opts.connMaxLifetime)
	db.DB.SetMaxIdleConns(db.opts.maxIdleConns)
	db.DB.SetMaxOpenConns(db.opts.maxOpenConns)
	cf = func() {
		db.DB.Close()
	}
	return
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
