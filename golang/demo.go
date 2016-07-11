package main

import (
	"fmt"

	"github.com/dailing/levlog"
	"github.com/dailing/redisSemaphore/golang/redisema"
)

func main() {
	fmt.Println("start")
	test := redisema.NewResiSemaphore("a")
	test.Init(10)
	err := test.Init(10)
	levlog.E(err)
	fmt.Println("finished")
}
