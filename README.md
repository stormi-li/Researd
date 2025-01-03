# RESEARD Guides

Simple and stable service registration and discovery library.

# Overview

- Support service registration
- Support service discovery
- Support heartbeat detection
- Every feature comes with tests
- Developer Friendly

# Install

```shell
go get -u github.com/stormi-li/Researd
```

# Quick Start

```go
package main

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

var redisAddr = “localhost:6379”
var password = “your password”

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
	})
	client := researd.NewClient(redisClient, "researd-namespace")
	go func() {
		client.Register("server", 3, "localhost:8080")
	}()
	client.Discover("server", func(addr string) {
		fmt.Println(addr)
	})
}
```

# Interface - researd

## NewClient

### Create researd client
```go
package main

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	researd "github.com/stormi-li/Researd"
)

var redisAddr = “localhost:6379”
var password = “your password”

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
	})
	client := researd.NewClient(redisClient, "researd-namespace")
}
```
The first parameter is a redis client of successful connection, the second parameter is a unique namespace.

# Interface - researd.Client

## Register

### Register server node information
```go
client.Register("server", 2, “localhost:8080”)
```
The first parameter is server name, the second parameter is node weightaddress, the third parameter is node address. The process will block and continue to send a heartbeat.

## Discover

### Discover a registered service with a heartbeat
```go
client.Discover("server", func(addr string) {
	fmt.Println(addr)
})
```
The first parameter is server name,  the second parameter is a handler for discovered server address.  The process will block and continue to listen heartbeat.

#  Community

## Ask

### How do I ask a good question?
- Email - 2785782829@qq.com
- Github Issues - https://github.com/stormi-li/Researd/issues