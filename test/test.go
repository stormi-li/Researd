package main

import (
	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "118.25.196.166:6379",
	})
	client := researd.NewClient(redisClient)
	client.Register("server", "lll:333", 3)
}
