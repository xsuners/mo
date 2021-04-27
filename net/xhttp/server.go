package xhttp

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xsuners/mo/net/description"
)

// Config .
type Config struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// Handler .
type Handler func(ctx context.Context, service, method string, data []byte, interceptor description.UnaryServerInterceptor) (interface{}, error)

type options struct {
	// tlsCfg         *tls.Config
	// onconnect      func(connection.Conn)
	// onclose        func(connection.Conn)
	// workerSize     int // numbers of worker go-routines
	// bufferSize     int // size of buffered channel
	// maxConnections int
	unknownServiceHandler Handler
}

var defaultOptions = options{
	// bufferSize:     256,
	// workerSize:     10000,
	// maxConnections: 1000,
}

// ServerOption sets server options.
type ServerOption func(*options)

// UnknownServiceHandler .
func UnknownServiceHandler(handler Handler) ServerOption {
	return func(o *options) {
		o.unknownServiceHandler = handler
	}
}

// Server .
type Server struct {
	conf   *Config
	server *gin.Engine
}

// New .
func New(c *Config) *Server {
	s := &Server{
		conf: c,
	}
	s.server = gin.Default()
	return s
}

// Server .
func (s *Server) Server() *gin.Engine {
	return s.server
}

// Start .
func (s *Server) Start() (err error) {
	err = s.server.Run(fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		log.Panic(err)
	}
	return
}

// Stop .
func (s *Server) Stop(ctx context.Context) (err error) {
	return
}
