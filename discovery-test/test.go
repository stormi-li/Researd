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
	client := researd.NewClient(redisClient, "researd-namespace", researd.Server)
	discover := client.NewSearcher("server")
	fmt.Println(discover.SearchServer())
	discover.Listen(func(address string, data map[string]string) {
		fmt.Println(address)
		fmt.Println(data)
	})
}
