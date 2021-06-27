package xmongo

import (
	"context"

	"github.com/xsuners/mo/log"
	"go.mongodb.org/mongo-driver/mongo"
	moptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type options struct {
	mopts []*moptions.ClientOptions
	// connMaxLifetime time.Duration
	// maxIdleConns    int
	// maxOpenConns    int
	url string
	db  string
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

// URL .
func URL(url string) Option {
	return newFuncOption(func(o *options) {
		o.url = url
	})
}

// DB .
func DB(db string) Option {
	return newFuncOption(func(o *options) {
		o.db = db
	})
}

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
func New(opts ...Option) (db *Database, cf func(), err error) {
	dopts := defaultOptions
	for _, o := range opts {
		o.apply(&dopts)
	}
	dopts.mopts = append(dopts.mopts, moptions.Client().ApplyURI(dopts.url))
	db = new(Database)
	if db.client, err = mongo.Connect(context.Background(), dopts.mopts...); err != nil {
		log.Errors("xmongo: connect error", zap.Error(err))
		return
	}
	db.opts = dopts
	db.Database = db.client.Database(db.opts.db)
	cf = func() {
		db.client.Disconnect(context.TODO())
	}
	return
}
