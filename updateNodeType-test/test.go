package main

import (
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
	register1 := client.NewRegister("channel-1", "118.25.196.166:8899")
	register1.Close()
	// register2 := client.NewConsumer("channel-1", "118.25.196.166:8999")
	// register1.ToStandby()
	// register2.ToMain()
}
