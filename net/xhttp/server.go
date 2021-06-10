package xhttp

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
)

// Handler .
type Handler func(ctx context.Context, service, method string, data []byte, interceptor description.UnaryServerInterceptor) (interface{}, error)

type options struct {
	unknownServiceHandler Handler
	// ip                    string
	port int
}

var defaultOptions = options{}

// Option sets server options.
type Option func(*options)

// UnknownServiceHandler .
func UnknownServiceHandler(handler Handler) Option {
	return func(o *options) {
		o.unknownServiceHandler = handler
	}
}

// // IP .
// func IP(ip string) Option {
// 	return func(o *options) {
// 		o.ip = ip
// 	}
// }

// Port .
func Port(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// Server .
type Server struct {
	server *gin.Engine
	opts   *options
}

// New .
func New(opt ...Option) (s *Server, cf func()) {
	opts := defaultOptions
	for _, opt := range opt {
		opt(&opts)
	}
	s = &Server{
		opts:   &opts,
		server: gin.Default(),
	}
	cf = func() {
		log.Info("xhttp is closing...")
		log.Info("xhttp is closed.")
	}
	return
}

// Server .
func (s *Server) Server() *gin.Engine {
	return s.server
}

// Start .
func (s *Server) Start() (err error) {
	err = s.server.Run(fmt.Sprintf(":%d", s.opts.port))
	if err != nil {
		return
	}
	return
}
