package ripc

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Ripc结构体
type Client struct {
	redisClient *redis.Client
	namespace   string
	ctx         context.Context
}

// 创建Ripc客户端
func NewClient(redisClient *redis.Client, namespace string) *Client {
	return &Client{
		redisClient: redisClient,
		namespace:   namespace + const_separator,
		ctx:         context.Background(),
	}
}

// 向所有监听频道的进程发送通知
func (c *Client) Notify(channel, msg string) {
	//使用redis的Publish功能发送通知--------------------------redis代码
	c.redisClient.Publish(c.ctx, c.namespace+channel, msg)
}

// 监听一个消息，返回收到的信息，如果超时返回""
func (c *Client) Wait(channel string, timeout time.Duration) string {
	//使用redis的Subscribe功能订阅频道--------------------------redis代码
	sub := c.redisClient.Subscribe(c.ctx, c.namespace+channel)

	msgChan := sub.Channel()
	defer sub.Close()

	if timeout == 0 {
		msg := <-msgChan
		return msg.Payload
	}

	timer := time.NewTicker(timeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		return ""
	case msg := <-msgChan:
		return msg.Payload
	}
}

func (c *Client) NewListener(channel string) *Listener {
	return newListener(c.redisClient.Subscribe(c.ctx, c.namespace+channel))
}

func (c *Client) NewLock(lockName string) *Lock {
	return newLock(c.redisClient, c, lockName, c.namespace)
}
