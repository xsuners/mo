package xsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Options struct {
	ConnMaxLifetime time.Duration `ini-name:"connMaxLifetime" long:"sql-connMaxLifetime" description:"database connMaxLifetime"`
	MaxIdleConns    int           `ini-name:"maxIdleConns" long:"sql-maxIdleConns" description:"database maxIdleConns"`
	MaxOpenConns    int           `ini-name:"maxOpenConns" long:"sql-maxOpenConns" description:"database maxOpenConns"`
	Username        string        `ini-name:"username" long:"sql-username" description:"database username"`
	Password        string        `ini-name:"password" long:"sql-password" description:"database password"`
	Name            string        `ini-name:"name" long:"sql-name" description:"database name"`
	IP              string        `ini-name:"ip" long:"sql-ip" description:"database ip"`
	Port            int           `ini-name:"port" long:"sql-port" description:"database port"`
	Driver          string        `ini-name:"driver" long:"sql-driver" description:"database driver"`
}

var defaultOptions = Options{
	Username:     "root",
	Password:     "123456",
	IP:           "127.0.0.1",
	Driver:       "mysql",
	Port:         3306,
	MaxIdleConns: 3,
	MaxOpenConns: 3,
}

// Option .
type Option func(o *Options)

// ConnMaxLifetime .
func ConnMaxLifetime(d time.Duration) Option {
	return func(o *Options) {
		o.ConnMaxLifetime = d
	}
}

// MaxIdleConns .
func MaxIdleConns(num int) Option {
	return func(o *Options) {
		o.MaxIdleConns = num
	}
}

// MaxOpenConns .
func MaxOpenConns(num int) Option {
	return func(o *Options) {
		o.MaxOpenConns = num
	}
}

// IP .
func IP(ip string) Option {
	return func(o *Options) {
		o.IP = ip
	}
}

// Port .
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// Username .
func Username(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// Password .
func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// Name .
func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// Driver .
func Driver(driver string) Option {
	return func(o *Options) {
		o.Driver = driver
	}
}

// Database is database clienr
type Database struct {
	*sql.DB
	opts Options
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
		db.opts.Username,
		db.opts.Password,
		db.opts.IP,
		db.opts.Port,
		db.opts.Name,
	)
	db.DB, err = sql.Open(db.opts.Driver, db.dsn)
	if err != nil {
		log.Fatal("unable to use data source name", err)
	}
	db.DB.SetConnMaxLifetime(db.opts.ConnMaxLifetime)
	db.DB.SetMaxIdleConns(db.opts.MaxIdleConns)
	db.DB.SetMaxOpenConns(db.opts.MaxOpenConns)
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
