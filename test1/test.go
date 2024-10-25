package main

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

var redisAddr = "your redis addr"

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	client := researd.NewClient(redisClient)
	client.Discover("server", func(addr string) {
		fmt.Println(addr)
	}) 
}
