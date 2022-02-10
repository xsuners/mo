package xhttp

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
)

// Handler .
type Handler func(ctx context.Context, service, method string, data []byte, interceptor description.UnaryServerInterceptor) (interface{}, error)

type Options struct {
	unknownServiceHandler Handler
	proxyHandler          func(c *gin.Context)
	middlewares           []gin.HandlerFunc
	pre                   func(engine *gin.Engine)
	// ip                    string
	// port int
}

var defaultOptions = Options{}

// Option sets server options.
type Option func(*Options)

// UnknownServiceHandler .
func UnknownServiceHandler(handler Handler) Option {
	return func(o *Options) {
		o.unknownServiceHandler = handler
	}
}

// ProxyHandler .
func ProxyHandler(handler func(c *gin.Context)) Option {
	return func(o *Options) {
		o.proxyHandler = handler
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

	opts     *Options
	mu       sync.Mutex
	services map[string]*description.ServiceInfo
}

// New .
func New(opt ...Option) (s *Server, cf func()) {
	opts := defaultOptions
	for _, opt := range opt {
		opt(&opts)
	}
	s = &Server{
		opts:     &opts,
		Engine:   gin.Default(),
		services: make(map[string]*description.ServiceInfo),
	}
	cf = func() {
		log.Info("xhttp is closing...")
		log.Info("xhttp is closed.")
	}
	return
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
func (s *Server) Serve(port int) (err error) {
	s.Use(s.opts.middlewares...)
	if s.opts.pre != nil {
		s.opts.pre(s.Engine)
	}

	if s.opts.proxyHandler != nil {
		s.POST("/rpc/:service/:method", s.opts.proxyHandler)
	}

	for sname, service := range s.services {
		for mname, m := range service.Methods() {
			s.POST("/"+sname+"/"+mname, wrap(m))
		}
	}

	// for consul health check
	s.GET("/", s.Check)

	err = s.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		return
	}
	return
}

func response(c *gin.Context, code int, message string, data interface{}) {
	out := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if data != nil {
		out["data"] = data
	}
	c.JSON(http.StatusOK, out)
}

func wrap(handler interface{}) gin.HandlerFunc {
	if handler == nil {
		panic("handler can not be nil")
	}
	// TODO 校验handler的函数签名
	f := reflect.ValueOf(handler)
	return func(c *gin.Context) {
		req := reflect.New(f.Type().In(1).Elem())       // 调用反射创建对象
		if err := c.Bind(req.Interface()); err != nil { // 解析请求参数
			response(c, 1, "请求参数不合法", nil)
			return
		}
		o := f.Call([]reflect.Value{reflect.ValueOf(c.Request.Context()), req}) // 调用handler
		if !o[1].IsNil() {                                                      // err != nil
			response(c, 1, o[1].Interface().(error).Error(), nil) // 错误响应
		} else {
			response(c, 0, "成功", o[0].Interface()) // 成功响应
		}
	}
}
