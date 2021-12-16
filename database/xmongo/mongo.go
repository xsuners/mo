package xmongo

import (
	"context"

	"github.com/xsuners/mo/log"
	"go.mongodb.org/mongo-driver/mongo"
	moptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Options struct {
	mopts []*moptions.ClientOptions
	// connMaxLifetime time.Duration
	// maxIdleConns    int
	// maxOpenConns    int
	URL string `ini-name:"url" long:"mongo-url" description:"mongo url"`
	DB  string `ini-name:"db" long:"mongo-db" description:"mongo db"`
}

var defaultOptions = Options{
	// maxIdleConns: 3,
	// maxOpenConns: 3,
}

// Option .
type Option interface {
	apply(*Options)
}

// EmptyOption .
type EmptyOption struct{}

func (EmptyOption) apply(*Options) {}

type funcOption struct {
	f func(*Options)
}

func (fdo *funcOption) apply(do *Options) {
	fdo.f(do)
}

func newFuncOption(f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// URL .
func URL(url string) Option {
	return newFuncOption(func(o *Options) {
		o.URL = url
	})
}

// DB .
func DB(db string) Option {
	return newFuncOption(func(o *Options) {
		o.DB = db
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
	return newFuncOption(func(o *Options) {
		o.mopts = append(o.mopts, mopts...)
	})
}

// Database is database clienr
type Database struct {
	*mongo.Database

	opts   Options
	client *mongo.Client
}

// New create a sql client .
func New(opts ...Option) (db *Database, cf func(), err error) {
	dopts := defaultOptions
	for _, o := range opts {
		o.apply(&dopts)
	}
	dopts.mopts = append(dopts.mopts, moptions.Client().ApplyURI(dopts.URL))
	db = new(Database)
	if db.client, err = mongo.Connect(context.Background(), dopts.mopts...); err != nil {
		log.Errors("xmongo: connect error", zap.Error(err))
		return
	}
	db.opts = dopts
	db.Database = db.client.Database(db.opts.DB)
	cf = func() {
		db.client.Disconnect(context.TODO())
	}
	return
}
