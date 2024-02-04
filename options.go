package mo

import (
	"fmt"
	"os"

	"github.com/skyasker/go-flags"
	"github.com/xsuners/mo/database/xmongo"
	"github.com/xsuners/mo/database/xredis"
	"github.com/xsuners/mo/database/xsql"
	"github.com/xsuners/mo/database/xxorm"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming/consul"
	"github.com/xsuners/mo/net/xcron"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xgrpc/client"
	"github.com/xsuners/mo/net/xhttp"
	hc "github.com/xsuners/mo/net/xhttp/client"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xnats/publisher"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
)

type Optioner interface {
	V() bool
	C() string
}

type Options struct {
	// Application options
	Version bool           `json:"version" ini-name:"version" short:"v" long:"version" description:"execution version"`
	Config  flags.Filename `json:"config" ini-name:"config" short:"c" long:"config" description:"config file path" default:"/etc/conf"`

	// Server options
	Log    log.Options    `json:"log" group:"log"`
	Consul consul.Options `json:"consul" group:"consul"`
	// Naming naming.Options `json:"naming" group:"naming"`

	// Storage options
	SQL   xsql.Options   `json:"sql" group:"sql"`
	XORM  xxorm.Options  `json:"xorm" group:"xorm"`
	Mongo xmongo.Options `json:"mongo" group:"mongo"`
	Redis xredis.Options `json:"redis" group:"redis"`

	// Network server options
	WS   xws.Options   `json:"ws" group:"ws"`
	TCP  xtcp.Options  `json:"tcp" group:"tcp"`
	NATS xnats.Options `json:"nats" group:"nats"`
	GRPC xgrpc.Options `json:"grpc" group:"grpc"`
	HTTP xhttp.Options `json:"http" group:"http"`
	CRON xcron.Options `json:"cron" group:"cron"`

	// Network client options
	GRPCC client.Options    `json:"grpcc" group:"grpcc"`
	NATSC publisher.Options `json:"natsc" group:"natsc"`
	HTTPC hc.Options        `json:"httpc" group:"httpc"`
}

var _ Optioner = (*Options)(nil)

func (o *Options) V() bool {
	return o.Version
}

func (o *Options) C() string {
	return string(o.Config)
}

func Parse(option Optioner) {
	parser := flags.NewParser(option, flags.Default|flags.IgnoreUnknown)
	ini := flags.NewIniParser(parser)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			fmt.Println(err)
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
	if option.V() {
		fmt.Print(BuildInfo())
		os.Exit(0)
	}
	if err := ini.ParseFile(option.C()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
