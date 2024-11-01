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
	client := researd.NewClient(redisClient, "researd-namespace")
	register1 := client.NewRegister("server", "1223213:1111")
	register2 := client.NewRegister("server", "1223213:2222")
	register1.UpdateNodeType(researd.Standby)
	register2.UpdateNodeType(researd.Main)
}
