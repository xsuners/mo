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
	proxyHandler          func(c *gin.Context)
	middlewares           []gin.HandlerFunc
	// ip                    string
	// port int
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

// ProxyHandler .
func ProxyHandler(handler func(c *gin.Context)) Option {
	return func(o *options) {
		o.proxyHandler = handler
	}
}

// Use .
func Use(middlewares ...gin.HandlerFunc) Option {
	return func(o *options) {
		o.middlewares = append(o.middlewares, middlewares...)
	}
}

// // IP .
// func IP(ip string) Option {
// 	return func(o *options) {
// 		o.ip = ip
// 	}
// }

// // Port .
// func Port(port int) Option {
// 	return func(o *options) {
// 		o.port = port
// 	}
// }

// Server .
type Server struct {
	*gin.Engine

	opts *options
}

// New .
func New(opt ...Option) (s *Server, cf func()) {
	opts := defaultOptions
	for _, opt := range opt {
		opt(&opts)
	}
	s = &Server{
		opts:   &opts,
		Engine: gin.Default(),
	}
	cf = func() {
		log.Info("xhttp is closing...")
		log.Info("xhttp is closed.")
	}
	return
}

// // Server .
// func (s *Server) Server() *gin.Engine {
// 	return s.server
// }

// Serve .
func (s *Server) Serve(port int) (err error) {
	s.Use(s.opts.middlewares...)
	if s.opts.proxyHandler != nil {
		s.POST("/rpc/:service/:method", s.opts.proxyHandler)
	}
	err = s.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
	return
}
