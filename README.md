# Researd 框架

## 简介

Researd 是一个基于 Redis 的服务注册与发现框架，能够为分布式系统提供高效、可靠的服务注册与发现机制。通过 Redis 作为核心存储和消息传递的媒介，实现了服务节点的管理和服务状态的监控。

## 功能

- 支持服务注册：服务注册节点能够将服务节点信息注册到 Redis 中，并且定期发送心跳信号。
- 支持服务发现：服务发现进程可以通过 Redis 快速查找和发现这些已注册的服务节点。
- 支持心跳检测：发现进程会持续监控心跳信号，检测到心跳停止会重新进行服务发现操作

## 安装

```shell
go get github.com/stormi-li/Researd
```

## 使用

### 1. Register

**Register**注册服务能够将服务节点信息注册到 Redis 中，并且定期发送心跳信号。

**示例代码**：

```go
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
```

### 2. Discover

**Discover**发现服务可以通过 Redis 快速查找和发现这些已注册的服务节点，并持续监控心跳信号，检测到心跳停止会重新进行服务发现操作。

**示例代码**：

```go
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
```