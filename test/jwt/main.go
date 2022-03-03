package main

import (
	"fmt"

	"github.com/xsuners/mo/metadata"
	"github.com/xsuners/mo/misc/jwt"
)

func main() {
	token, err := jwt.New(&metadata.Metadata{
		Appid:  10000,
		Id:     100001,
		Device: 1,
		Ints:   map[string]int64{"hello": 1},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(token)

	md, err := jwt.Parse(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(md)
}
