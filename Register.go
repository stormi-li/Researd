package researd

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Ripc"
)

type Register struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
	ctx         context.Context
	serverName  string
	nodeType    NodeType
	addr        string
}

func newRegister(redisClient *redis.Client, ripcClient *ripc.Client, namespace string, serverName string, addr string) *Register {
	return &Register{
		redisClient: redisClient,
		ripcClient:  ripcClient,
		namespace:   namespace,
		serverName:  serverName,
		ctx:         context.Background(),
		addr:        addr,
	}
}

func (register *Register) Start(nodeType NodeType) {
	register.nodeType = nodeType
	key := register.namespace + register.serverName + const_splitChar + nodeType.String() + const_splitChar + register.addr
	go func() {
		for {
			register.redisClient.Set(register.ctx, key, "", const_expireTime)
			time.Sleep(const_expireTime / 2)
		}
	}()
	channel := register.serverName + const_splitChar + register.addr
	register.ripcClient.NewListener(channel).Listen(func(msg string) {
		if msg == const_ask {
			for i := 0; i < 10; i++ {
				register.ripcClient.Notify(channel, const_alive)
				time.Sleep(100 * time.Millisecond)
			}
		}
		if msg != const_alive {
			if command, nodeType := splitNodeType(msg); command == const_updateNodeType {
				register.redisClient.Del(register.ctx, key)
				key = register.namespace + register.serverName + const_splitChar + nodeType + const_splitChar + register.addr
				register.redisClient.Set(register.ctx, key, "", const_expireTime)
			}
		}
	})
}

func (register *Register) UpdateNodeType(nodeType NodeType) {
	if nodeType != Main && nodeType != Standby {
		return
	}
	channel := register.serverName + const_splitChar + register.addr
	register.ripcClient.Notify(channel, const_updateNodeType+const_splitChar+nodeType.String())
}

func splitNodeType(address string) (string, string) {
	index := strings.Index(address, const_splitChar)
	if index == -1 {
		return "", ""
	}
	return address[:index], address[index+1:]
}
