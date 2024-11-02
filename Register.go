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
	address     string
	channel     string
	close       chan struct{}
}

func newRegister(redisClient *redis.Client, ripcClient *ripc.Client, namespace string, serverName string, address string) *Register {
	return &Register{
		redisClient: redisClient,
		ripcClient:  ripcClient,
		namespace:   namespace,
		serverName:  serverName,
		ctx:         context.Background(),
		address:     address,
		channel:     serverName + const_separator + address,
		close:       make(chan struct{}, 1),
	}
}

func (register *Register) StartOnMain(data map[string]string) {
	register.start(node_main, data)
}

func (register *Register) StartOnStandby(data map[string]string) {
	register.start(node_standby, data)
}

func (register *Register) start(nodeType string, data map[string]string) {
	jsonByte, _ := json.MarshalIndent(data, " ", "  ")
	jsonStr := string(jsonByte)
	register.nodeType = nodeType
	nodeState := state_start
	ticker := time.NewTicker(const_expireTime / 2)
	close := make(chan struct{}, 1)
	update := make(chan struct{}, 1)

	updateHandler := func() {
		key := register.namespace + register.serverName + const_separator + nodeState + const_separator + nodeType + const_separator + register.address
		register.redisClient.Set(register.ctx, key, jsonStr, const_expireTime)
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				updateHandler()
			case <-update:
				updateHandler()
			case <-close:
				return
			}
		}
	}()
	channel := register.serverName + const_separator + register.address
	listener := register.ripcClient.NewListener(channel)
	listener.Listen(func(msg string) {
		if msg == command_close {
			close <- struct{}{}
			register.close <- struct{}{}
			listener.Close()
		}
		if msg == command_start {
			nodeState = state_start
			update <- struct{}{}
		}
		if msg == command_stop {
			nodeState = state_stop
			update <- struct{}{}
		}
		if msg == command_main {
			nodeType = node_main
			update <- struct{}{}
		}
		if msg == command_standby {
			nodeType = node_standby
			update <- struct{}{}
		}
		if command, json := splitCommand(msg); command == command_updateNodeData {
			jsonStr = json
			update <- struct{}{}
		}
	})
}

func (register *Register) ToMain() {
	register.ripcClient.Notify(register.channel, command_main)
}

func (register *Register) ToStandby() {
	register.ripcClient.Notify(register.channel, command_standby)
}

func (register *Register) Close() {
	register.ripcClient.Notify(register.channel, command_close)
}

func (register *Register) Start() {
	register.ripcClient.Notify(register.channel, command_start)
}

func (register *Register) Stop() {
	register.ripcClient.Notify(register.channel, command_stop)
}

func (register *Register) UpdateData(data map[string]string) {
	jsonStr, _ := json.MarshalIndent(data, " ", "  ")
	register.ripcClient.Notify(register.channel, command_updateNodeData+":"+string(jsonStr))
}

func splitCommand(address string) (string, string) {
	index := strings.Index(address, const_separator)
	if index == -1 {
		return "", ""
	}
	return address[:index], address[index+1:]
}
