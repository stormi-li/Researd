package main

import researd "github.com/stormi-li/Researd"

func main() {
	client, _ := researd.NewClient("localhost:6379")
	client.Register("server", "lll:111", 3)

}
