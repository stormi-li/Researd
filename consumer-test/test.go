package main

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

var redisAddr = "118.25.196.166:3934"
var password = "12982397StrongPassw0rd"

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
	})
	client := researd.NewClient(redisClient, "researd-namespace", researd.MQ)
	consumer := client.NewConsumer("channel-1", "118.25.196.166:8899")
	consumer.StartOnMain(func(message []byte) {
		fmt.Println(string(message))
	})
}
