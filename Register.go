package researd

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Researd/Ripc"
)

type Register struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
	ctx         context.Context
	serverName  string
	nodeType    string
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

func (register *Register) StartOnMain(data map[string]string) {
	register.start(node_main, data)
}

func (register *Register) StartOnStandby(data map[string]string) {
	register.start(node_standby, data)
}

func (register *Register) start(nodeType string, data map[string]string) {
	jsonStr, _ := json.MarshalIndent(data, " ", "  ")
	register.nodeType = nodeType
	key := register.namespace + register.serverName + const_separator + nodeType + const_separator + register.addr
	go func() {
		for {
			register.redisClient.Set(register.ctx, key, jsonStr, const_expireTime)
			time.Sleep(const_expireTime / 2)
		}
	}()
	channel := register.serverName + const_separator + register.addr
	register.ripcClient.NewListener(channel).Listen(func(msg string) {
		if command, nodeType := splitNodeType(msg); command == const_updateNodeType {
			register.redisClient.Del(register.ctx, key)
			key = register.namespace + register.serverName + const_separator + nodeType + const_separator + register.addr
			register.redisClient.Set(register.ctx, key, jsonStr, const_expireTime)
		}
	})
}

func (register *Register) ToMain() {
	register.updateNodeType(node_main)
}

func (register *Register) ToStandby() {
	register.updateNodeType(node_standby)
}

func (register *Register) updateNodeType(nodeType string) {
	channel := register.serverName + const_separator + register.addr
	register.ripcClient.Notify(channel, const_updateNodeType+const_separator+nodeType)
}

func splitNodeType(address string) (string, string) {
	index := strings.Index(address, const_separator)
	if index == -1 {
		return "", ""
	}
	return address[:index], address[index+1:]
}
