package xsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	Mysql    = "mysql"
	Postgres = "postgres"
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

type SQL interface {
	Exec(ctx context.Context, query string, args ...interface{}) (af, id int64, err error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	One(ctx context.Context, query string, args ...any) *sql.Row
}

type Database interface {
	SQL
	Tx(ctx context.Context, fns ...func(tx SQL) error) error
}

// database is database clienr
type database struct {
	*sql.DB
	opts Options
	dsn  string
}

var _ Database = (*database)(nil)

// New create a sql client .
func New(opts ...Option) (Database, func(), error) {
	db := &database{
		opts: defaultOptions,
	}
	for _, o := range opts {
		o(&db.opts)
	}
	switch db.opts.Driver {
	case Mysql:
		db.dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&loc=Local&parseTime=True",
			db.opts.Username,
			db.opts.Password,
			db.opts.IP,
			db.opts.Port,
			db.opts.Name,
		)
	case Postgres:
		db.dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
			db.opts.Username,
			db.opts.Password,
			db.opts.IP,
			db.opts.Port,
			db.opts.Name,
		)
	default:
		panic(fmt.Sprintf("unknown driver %s", db.opts.Driver))
	}
	var err error
	db.DB, err = sql.Open(db.opts.Driver, db.dsn)
	if err != nil {
		log.Fatal("unable to use data source name", err)
	}
	db.DB.SetConnMaxLifetime(db.opts.ConnMaxLifetime)
	db.DB.SetMaxIdleConns(db.opts.MaxIdleConns)
	db.DB.SetMaxOpenConns(db.opts.MaxOpenConns)
	return db, func() {
		db.DB.Close()
	}, nil
}

func (db *database) Exec(ctx context.Context, query string, args ...interface{}) (af, id int64, err error) {
	ret, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return
	}
	if db.opts.Driver == Mysql {
		af, err = ret.RowsAffected()
		if err != nil {
			return
		}
	}
	id, err = ret.LastInsertId()
	if err != nil {
		return
	}
	return
}

func (db *database) One(ctx context.Context, query string, args ...any) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *database) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *database) Tx(ctx context.Context, fns ...func(tx SQL) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	for _, fn := range fns {
		if err = fn(&Tx{tx}); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
