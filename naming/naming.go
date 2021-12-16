package naming

type Protocol string

const (
	WS    Protocol = "ws"
	WSS   Protocol = "wss"
	TCP   Protocol = "tcp"
	TCPS  Protocol = "tcps"
	HTTP  Protocol = "http"
	HTTPS Protocol = "https"
	GRPC  Protocol = "grpc"
	GRPCS Protocol = "grpcs"
)

type Options struct {
	Name string `json:"name" ini-name:"name" long:"naming-name" description:"naming name"`
	Host string `json:"host" ini-name:"host" long:"naming-host" description:"naming host"`
	Port int32  `json:"port" ini-name:"port" long:"naming-port" description:"naming port"`
}

// Service .
type Service struct {
	Name     string
	Protocol Protocol
	IP       string
	Port     int
	Tag      []string
}

type Naming interface {
	Deregister()
	Register(svc *Service) (err error)
}
