// Package env get env & app config, all the public field must after init()
// finished and flag.Parse().
package env

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// deploy env.
const (
	EnvDev  = "dev"
	EnvFat  = "fat"
	EnvUat  = "uat"
	EnvPre  = "pre"
	EnvProd = "prod"
)

// env default value.
const (
	_dc  = "HZ"
	_dev = EnvDev
)

// env configuration.
var (
	// DC avaliable zone where app at.
	DC string
	// Hostname machine hostname.
	Hostname string
	// Env deploy env where app at.
	Env string
	// Color is the identification of different experimental group in one caster cluster.
	Color string
)

func init() {
	var err error
	Hostname = os.Getenv("HOSTNAME")
	if Hostname == "" {
		Hostname, err = os.Hostname()
		if err != nil {
			Hostname = strconv.Itoa(int(time.Now().UnixNano()))
		}
	}
	addFlag(flag.CommandLine)
}

func addFlag(fs *flag.FlagSet) {
	fs.StringVar(&DC, "env.dc", defaultString("ENV_DC", _dc), "avaliable zone. or use DC env variable, value: sh001/sh002 etc.")
	fs.StringVar(&Env, "env.env", defaultString("ENV_ENV", _dev), "deploy env. or use ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	fs.StringVar(&Color, "env.color", os.Getenv("ENV_COLOR"), "color is the identification of different experimental group.")
}

func defaultString(env, value string) string {
	v := os.Getenv(env)
	if v == "" {
		return value
	}
	return v
}
