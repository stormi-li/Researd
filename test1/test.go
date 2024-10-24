package main

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "118.25.196.166:6379",
	})
	client := researd.NewClient(redisClient)
	client.SetNamespace("a")

	client.Connect("server", func(addr string) {
		fmt.Println(addr)
	})
}
