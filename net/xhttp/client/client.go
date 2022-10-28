package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/xsuners/mo/net/description"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/status"
)

type Options struct {
	unaryInt description.UnaryClientInterceptor
	// streamInt StreamClientInterceptor
	chainUnaryInts []description.UnaryClientInterceptor
	// chainStreamInts []StreamClientInterceptor

	IP   string `ini-name:"ip" long:"httpc-ip" description:"httpc ip"`
	Port int    `ini-name:"port" long:"httpc-port" description:"httpc port"`

	credentials string
	pkg         string
	service     string
}

// Option configures how we set up the connection.
type Option func(*Options)

func WithUnaryInterceptor(f description.UnaryClientInterceptor) Option {
	return func(o *Options) {
		o.unaryInt = f
	}
}

func WithChainUnaryInterceptor(interceptors ...description.UnaryClientInterceptor) Option {
	return func(o *Options) {
		o.chainUnaryInts = append(o.chainUnaryInts, interceptors...)
	}
}

// IP (e.g., chaining) can be implemented at the caller.
func IP(ip string) Option {
	return func(o *Options) {
		o.IP = ip
	}
}

// Port (e.g., chaining) can be implemented at the caller.
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

func Credentials(crets string) Option {
	return func(o *Options) {
		o.credentials = crets
	}
}

func Package(pkg string) Option {
	return func(o *Options) {
		o.pkg = pkg
	}
}

func Service(svc string) Option {
	return func(o *Options) {
		o.service = svc
	}
}

func defaultDialOptions() Options {
	return Options{}
}

type Client struct {
	dopts Options

	*http.Client
}

var _ description.ClientConnInterface = (*Client)(nil)

// New .
func New(opts ...Option) (*Client, error) {

	c := &Client{
		dopts: defaultDialOptions(),
		Client: &http.Client{
			Timeout: time.Minute,
		},
	}

	for _, opt := range opts {
		opt(&c.dopts)
	}

	chainUnaryClientInterceptors(c)

	// // TODO
	// nc, err := nats.Connect(pub.dopts.urls, pub.dopts.nopts...)
	// if err != nil {
	// 	log.Fatalw("xnats: publisher connect error", "err", err)
	// 	return nil, err
	// }

	// pub.conn = nc
	return c, nil
}

// chainUnaryClientInterceptors chains all unary client interceptors into one.
func chainUnaryClientInterceptors(cc *Client) {
	interceptors := cc.dopts.chainUnaryInts
	// Prepend dopts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
	if cc.dopts.unaryInt != nil {
		interceptors = append([]description.UnaryClientInterceptor{cc.dopts.unaryInt}, interceptors...)
	}
	var chainedInt description.UnaryClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = func(ctx context.Context, method string, req, reply interface{}, cc description.UnaryClient, invoker description.UnaryInvoker, opts ...description.CallOption) error {
			return interceptors[0](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, 0, invoker), opts...)
		}
	}
	cc.dopts.unaryInt = chainedInt
}

// getChainUnaryInvoker recursively generate the chained unary invoker.
func getChainUnaryInvoker(interceptors []description.UnaryClientInterceptor, curr int, finalInvoker description.UnaryInvoker) description.UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc description.UnaryClient, opts ...description.CallOption) error {
		return interceptors[curr+1](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, curr+1, finalInvoker), opts...)
	}
}

// Close .
func (c *Client) Close() {
	// pub.conn.Close()
}

// Invoke .
func (c *Client) Invoke(ctx context.Context, sm string, args interface{}, reply interface{}, opts ...description.CallOption) error {
	if c.dopts.unaryInt != nil {
		return c.dopts.unaryInt(ctx, sm, args, reply, c, invoke, opts...)
	}
	return invoke(ctx, sm, args, reply, c, opts...)
}

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var respool = sync.Pool{
	New: func() interface{} {
		return &response{}
	},
}

func invoke(ctx context.Context, sm string, args interface{}, reply interface{}, cc description.UnaryClient, opts ...description.CallOption) error {

	// TODO
	c, ok := cc.(*Client)
	if !ok {
		return fmt.Errorf("xnats: pub invoke error: cc type (%T) not match", cc)
	}

	co := copool.Get().(*CallOptions)
	defer copool.Put(co)

	for _, o := range opts {
		o.Apply(co)
	}

	argsb, err := json.Marshal(args)
	if err != nil {
		return err
	}

	url := "http://" + c.dopts.IP + ":" + strconv.Itoa(c.dopts.Port) + sm

	fmt.Println(url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(argsb))
	if err != nil {
		return err
	}
	rsp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		st := new(spb.Status)
		err = json.Unmarshal(body, st)
		if err != nil {
			return err
		}
		return status.ErrorProto(st)
	}

	err = json.Unmarshal(body, reply)
	if err != nil {
		return err
	}

	return nil
}

// NewStream begins a streaming RPC.
func (c *Client) NewStream(ctx context.Context, desc *description.StreamDesc, method string, opts ...description.CallOption) (cs description.ClientStream, err error) {
	return
}
