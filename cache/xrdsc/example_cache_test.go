package xrdsc

import (
	"context"
	"fmt"
	"time"

	"github.com/xsuners/mo/database/xredis"
)

type Object struct {
	Str string
	Num int
}

func Example_basicUsage() {
	// ring := redis.NewRing(&redis.RingOptions{
	// 	Addrs: map[string]string{
	// 		"server1": ":6379",
	// 		"server2": ":6380",
	// 	},
	// })

	// mycache := New(&Options{
	// 	Redis:      ring,
	// 	LocalCache: NewTinyLFU(1000, time.Minute),
	// })

	rds, _, err := xredis.New()
	if err != nil {
		panic(err)
	}
	mycache := New(rds, Local(NewTinyLFU(1000, time.Minute)))

	ctx := context.TODO()
	key := "mykey"
	obj := &Object{
		Str: "mystring",
		Num: 42,
	}

	if err := mycache.Set(&Item{
		Ctx:   ctx,
		Key:   key,
		Value: obj,
		TTL:   time.Hour,
	}); err != nil {
		panic(err)
	}

	var wanted Object
	if err := mycache.Get(ctx, key, &wanted); err == nil {
		fmt.Println(wanted)
	}

	// Output: {mystring 42}
}

func Example_advancedUsage() {
	// ring := redis.NewRing(&redis.RingOptions{
	// 	Addrs: map[string]string{
	// 		"server1": ":6379",
	// 		"server2": ":6380",
	// 	},
	// })

	// mycache := New(&Options{
	// 	Redis:      ring,
	// 	LocalCache: NewTinyLFU(1000, time.Minute),
	// })

	rds, _, err := xredis.New()
	if err != nil {
		panic(err)
	}
	mycache := New(rds, Local(NewTinyLFU(1000, time.Minute)))
	obj := new(Object)
	err = mycache.Once(&Item{
		Key:   "mykey",
		Value: obj, // destination
		Do: func(*Item) (interface{}, error) {
			return &Object{
				Str: "mystring",
				Num: 42,
			}, nil
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(obj)
	// Output: &{mystring 42}
}
