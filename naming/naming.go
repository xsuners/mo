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
