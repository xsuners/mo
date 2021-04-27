package config

type options struct {
	localPath  string
	remoteAddr string
}

var defaultOptions = options{
	localPath: "../conf/conf.json",
}

// Option sets server options.
type Option func(*options)

// LocalPath returns a Option that will set TLS credentials for server
// connections.
func LocalPath(path string) Option {
	return func(o *options) {
		o.localPath = path
	}
}

// RemoteAddr .
func RemoteAddr(addr string) Option {
	return func(o *options) {
		o.remoteAddr = addr
	}
}
