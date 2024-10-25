package main

import (
	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

var redisAddr = "your redis addr"

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	client := researd.NewClient(redisClient)
	client.Register("server", "lll:333", 3) //参数为服务名，地址，权重
}
