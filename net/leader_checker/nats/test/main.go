package main

import (
	"time"

	"github.com/xsuners/mo/net/leader_checker/nats"
)

func main() {
	_, cf, err := nats.New("order")
	if err != nil {
		panic(err)
	}
	defer cf()

	time.Sleep(time.Minute)
}
