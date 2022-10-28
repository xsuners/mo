package xxorm

type Options struct {
	MaxIdleConns int    `ini-name:"maxIdleConns" long:"xorm-maxIdleConns" description:"database maxIdleConns"`
	MaxOpenConns int    `ini-name:"maxOpenConns" long:"xorm-maxOpenConns" description:"database maxOpenConns"`
	Username     string `ini-name:"username" long:"xorm-username" description:"database username"`
	Password     string `ini-name:"password" long:"xorm-password" description:"database password"`
	Name         string `ini-name:"name" long:"xorm-name" description:"database name"`
	IP           string `ini-name:"ip" long:"xorm-ip" description:"database ip"`
	Port         int    `ini-name:"port" long:"xorm-port" description:"database port"`
	Driver       string `ini-name:"driver" long:"xorm-driver" description:"database driver"`
}

var defaultOptions = Options{
	Username:     "root",
	Password:     "123456",
	IP:           "127.0.0.1",
	Driver:       "mysql",
	Port:         3306,
	MaxIdleConns: 3,
	MaxOpenConns: 3,
}

// Option .
type Option func(o *Options)

// MaxIdleConns .
func MaxIdleConns(num int) Option {
	return func(o *Options) {
		o.MaxIdleConns = num
	}
}

// MaxOpenConns .
func MaxOpenConns(num int) Option {
	return func(o *Options) {
		o.MaxOpenConns = num
	}
}

// IP .
func IP(ip string) Option {
	return func(o *Options) {
		o.IP = ip
	}
}

// Port .
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// Username .
func Username(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// Password .
func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// Name .
func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// Driver .
func Driver(driver string) Option {
	return func(o *Options) {
		o.Driver = driver
	}
}
