package main

import (
	"fmt"

	researd "github.com/stormi-li/Researd"
)

func main() {
	client, _ := researd.NewClient("localhost:6379")
	client.Connect("server", func(addr string) {
		fmt.Println(addr)
	})
}
