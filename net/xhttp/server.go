package xhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/misc/uhttp"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"google.golang.org/grpc/status"
)

// Handler .
type Handler func(ctx context.Context, service, method string, data []byte, interceptor description.UnaryServerInterceptor) (interface{}, error)

type Options struct {
	unknownServiceHandler Handler
	proxyer               func(c *gin.Context)
	exporter              func(c *gin.Context)
	importer              func(c *gin.Context)
	callbacker            func(c *gin.Context)
	middlewares           []gin.HandlerFunc
	pre                   func(engine *gin.Engine)
	unaryInt              description.UnaryServerInterceptor
	chainUnaryInts        []description.UnaryServerInterceptor
	// ip                    string
	Port int
}

var defaultOptions = Options{
	Port: 8000,
}

// Option sets server options.
type Option func(*Options)

// UnknownServiceHandler .
func UnknownServiceHandler(handler Handler) Option {
	return func(o *Options) {
		o.unknownServiceHandler = handler
	}
}

// Proxyer .
func Proxyer(handler func(c *gin.Context)) Option {
	return func(o *Options) {
		o.proxyer = handler
	}
}

// Exporter .
func Exporter(handler func(c *gin.Context)) Option {
	return func(o *Options) {
		o.exporter = handler
	}
}

// Importer .
func Importer(handler func(c *gin.Context)) Option {
	return func(o *Options) {
		o.importer = handler
	}
}

// Callbacker .
func Callbacker(handler func(c *gin.Context)) Option {
	return func(o *Options) {
		o.callbacker = handler
	}
}

// Use .
func Use(middlewares ...gin.HandlerFunc) Option {
	return func(o *Options) {
		o.middlewares = append(o.middlewares, middlewares...)
	}
}

// Pre .
func Pre(fn func(engine *gin.Engine)) Option {
	return func(o *Options) {
		o.pre = fn
	}
}

func UnaryInterceptor(i description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		if o.unaryInt != nil {
			panic("The unary server interceptor was already set and may not be reset.")
		}
		o.unaryInt = i
	}
}

func ChainUnaryInterceptor(interceptors ...description.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	}
}

// Port .
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// Server .
type Server struct {
	*gin.Engine

	opts     *Options
	mu       sync.Mutex
	services map[string]*description.ServiceInfo
}

// New .
func New(opt ...Option) (description.Server, func()) {
	opts := defaultOptions
	for _, opt := range opt {
		opt(&opts)
	}
	s := &Server{
		opts:     &opts,
		Engine:   gin.Default(),
		services: make(map[string]*description.ServiceInfo),
	}
	chainUnaryServerInterceptors(s)
	return s, func() {
		log.Info("xhttp is closing...")
		log.Info("xhttp is closed.")
	}
}

// chainUnaryServerInterceptors chains all unary server interceptors into one.
func chainUnaryServerInterceptors(s *Server) {
	// Prepend opts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
	interceptors := s.opts.chainUnaryInts
	if s.opts.unaryInt != nil {
		interceptors = append([]description.UnaryServerInterceptor{s.opts.unaryInt}, s.opts.chainUnaryInts...)
	}

	var chainedInt description.UnaryServerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = func(ctx context.Context, req interface{}, info *description.UnaryServerInfo, handler description.UnaryHandler) (interface{}, error) {
			return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
		}
	}

	s.opts.unaryInt = chainedInt
}

// getChainUnaryHandler recursively generate the chained UnaryHandler
func getChainUnaryHandler(interceptors []description.UnaryServerInterceptor, curr int, info *description.UnaryServerInfo, finalHandler description.UnaryHandler) description.UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}

	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}

// Register .
func (s *Server) Register(ss interface{}, sds ...*description.ServiceDesc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sd := range sds {
		err := description.Register(&s.services, sd, ss)
		if err != nil {
			log.Fatalw("xnats register service error", "err", err)
		}
	}
}

func (s *Server) Check(c *gin.Context) {}

// Serve .
func (s *Server) Serve() (err error) {
	s.Engine.Use(uhttp.Cors())
	s.Use(s.opts.middlewares...)

	if s.opts.pre != nil {
		s.opts.pre(s.Engine)
	}

	if s.opts.callbacker != nil {
		s.Any("/cb/:service/:method", s.opts.callbacker)
	}

	if s.opts.proxyer != nil {
		s.POST("/rpc/:service/:method", s.opts.proxyer)
	}

	if s.opts.exporter != nil {
		s.POST("/data/:service/:method", s.opts.exporter)
	}

	if s.opts.exporter != nil {
		s.POST("/download/:service/:method", s.opts.exporter)
	}

	if s.opts.importer != nil {
		s.POST("/upload/:service/:method", s.opts.importer)
	}

	for sname, service := range s.services {
		for mname, m := range service.Methods() {
			s.POST("/"+sname+"/"+mname, s.wrap(service.Service(), m.Handler))
		}
	}

	// for consul health check
	s.GET("/", s.Check)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opts.Port))
	if err != nil {
		return err
	}
	s.opts.Port = lis.Addr().(*net.TCPAddr).Port
	err = s.RunListener(lis)

	// err = s.Run(fmt.Sprintf(":%d", s.opts.Port))
	if err != nil {
		return
	}
	return
}

// func response(c *gin.Context, code int, message string, data interface{}) {
// 	out := map[string]interface{}{
// 		"code":    code,
// 		"message": message,
// 	}
// 	if data != nil {
// 		out["data"] = data
// 	}
// 	c.JSON(http.StatusOK, out)
// }

func (s *Server) wrap(svc interface{}, handler interface{}) gin.HandlerFunc {
	if handler == nil {
		panic("handler can not be nil")
	}
	// TODO 校验handler的函数签名
	f := reflect.ValueOf(handler)
	return func(c *gin.Context) {
		dec := func(req interface{}) error {
			if err := c.BindJSON(req); err != nil { // 解析请求参数
				return err
			}
			return nil
		}
		o := f.Call([]reflect.Value{
			reflect.ValueOf(svc),
			reflect.ValueOf(c.Request.Context()),
			reflect.ValueOf(dec),
			reflect.ValueOf(s.opts.unaryInt)}) // 调用handler
		if !o[1].IsNil() { // err != nil
			st := status.Convert(o[1].Interface().(error))
			c.JSON(uhttp.Code2Status(st.Code()), st.Proto())
			// response(c, 1, o[1].Interface().(error).Error(), nil) // 错误响应
		} else {
			c.JSON(http.StatusOK, o[0].Interface())
			// response(c, 0, "成功", o[0].Interface()) // 成功响应
		}
	}
}

func (s *Server) Naming(nm naming.Naming) error {
	for name := range s.services {
		ins := &naming.Service{
			Name:     name,
			Protocol: naming.HTTP,
			IP:       ip.Internal(),
			Port:     s.opts.Port,
			Tag:      []string{"http"},
		}
		if err := nm.Register(ins); err != nil {
			return err
		}
	}
	return nil
}
