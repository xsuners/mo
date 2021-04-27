package xmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	moptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Config .
type Config struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
}

type options struct {
	mopts []*moptions.ClientOptions
	// connMaxLifetime time.Duration
	// maxIdleConns    int
	// maxOpenConns    int
}

var defaultOptions = options{
	// maxIdleConns: 3,
	// maxOpenConns: 3,
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

// // ConnMaxLifetime .
// func ConnMaxLifetime(d time.Duration) Option {
// 	return newFuncOption(func(o *options) {
// 		o.connMaxLifetime = d
// 	})
// }

// // MaxIdleConns .
// func MaxIdleConns(num int) Option {
// 	return newFuncOption(func(o *options) {
// 		o.maxIdleConns = num
// 	})
// }

// // MaxOpenConns .
// func MaxOpenConns(num int) Option {
// 	return newFuncOption(func(o *options) {
// 		o.maxOpenConns = num
// 	})
// }

// MongoOptions .
func MongoOptions(mopts []*moptions.ClientOptions) Option {
	return newFuncOption(func(o *options) {
		o.mopts = append(o.mopts, mopts...)
	})
}

// Database is database clienr
type Database struct {
	*mongo.Database

	opts   options
	client *mongo.Client
}

// New create a sql client .
func New(c *Config, opts ...Option) *Database {
	var err error
	dopts := defaultOptions
	for _, o := range opts {
		o.apply(&dopts)
	}
	dopts.mopts = append(dopts.mopts, moptions.Client().ApplyURI(c.URI))
	db := &Database{}
	if db.client, err = mongo.Connect(context.Background(), dopts.mopts...); err != nil {
		panic(err)
	}
	db.opts = dopts
	db.Database = db.client.Database(c.Database)
	return db
}

// Close close the connection.
func (db *Database) Close() {
	db.client.Disconnect(context.TODO())
}
