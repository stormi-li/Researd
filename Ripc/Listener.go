package ripc

import "github.com/go-redis/redis/v8"

// 监听器结构体
type Listener struct {
	shutdown chan struct{}
	sub      *redis.PubSub
}

func newListener(sub *redis.PubSub) *Listener {
	return &Listener{
		shutdown: make(chan struct{}, 1),
		sub:      sub,
	}
}

// 关闭监听器
func (listener Listener) Close() {
	listener.shutdown <- struct{}{}
}

// 接受所有发送过来的消息，并执行handler
func (listener *Listener) Listen(handler func(msg string)) {
	c := listener.sub.Channel()
	defer listener.sub.Close()
	for {
		select {
		case msg := <-c:
			handler(msg.Payload)
		case <-listener.shutdown:
			return
		}
	}
}
