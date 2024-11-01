package researd

import (
	"context"
	"strconv"
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
	weight      int
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

func (register *Register) Start(weight int) {
	register.weight = weight
	key := register.namespace + register.serverName + ":" + strconv.Itoa(register.weight) + ":" + register.addr
	go func() {
		for {
			register.redisClient.Set(register.ctx, key, "", 30*time.Second)
			time.Sleep(15 * time.Second)
		}
	}()
	channel := register.serverName + ":" + register.addr
	register.ripcClient.NewListener(channel).Listen(func(msg string) {
		if msg == const_ask {
			for i := 0; i < 10; i++ {
				register.ripcClient.Notify(channel, const_alive)
				time.Sleep(100 * time.Millisecond)
			}
		}
		if msg != const_alive {
			if request, newWeight := splitWeight(msg); request == const_updateWeight {
				register.redisClient.Del(register.ctx, key)
				key = register.namespace + register.serverName + ":" + strconv.Itoa(newWeight) + ":" + register.addr
				register.redisClient.Set(register.ctx, key, "", 30*time.Second)
			}
		}
	})
}

func (register *Register) UpdateWeight(weight int) {
	register.weight = weight
	channel := register.serverName + ":" + register.addr
	register.ripcClient.Notify(channel, const_updateWeight+":"+strconv.Itoa(weight))
}
