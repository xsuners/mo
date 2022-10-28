package nats

type Options struct {
	Urls        string `ini-name:"urls" long:"natslc-urls" description:"nats urls"`
	Credentials string `ini-name:"credentials" long:"natslc-credentials" description:"nats credentials"`
	Members     int    `ini-name:"members" long:"natslc-members" description:"nats members"`
	LogPath     string `ini-name:"logPath" long:"natslc-logPath" description:"nats logPath"`
}

var defaultOptions = Options{
	Members: 1,
	LogPath: "/tmp/graft.log",
}

type Option func(*Options)

func Urls(urls string) Option {
	return func(o *Options) {
		o.Urls = urls
	}
}

func Members(num int) Option {
	return func(o *Options) {
		o.Members = num
	}
}

func LogPath(path string) Option {
	return func(o *Options) {
		o.LogPath = path
	}
}
