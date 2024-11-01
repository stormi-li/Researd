package researd

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Ripc"
)

type Discover struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
	ctx         context.Context
	serverName  string
}

func newDiscover(redisClient *redis.Client, ripcClient *ripc.Client, namespace string, serverName string) *Discover {
	return &Discover{
		redisClient: redisClient,
		ripcClient:  ripcClient,
		namespace:   namespace,
		serverName:  serverName,
		ctx:         context.Background(),
	}
}

func (discover *Discover) SearchNodes() []string {
	addrs := getKeysByNamespace(discover.redisClient, discover.namespace+discover.serverName)
	sort.Slice(addrs, func(a, b int) bool {
		return addrs[a] < addrs[b]
	})
	return addrs
}
func (discover *Discover) getMainNodeAddress() string {
	addrs := discover.SearchNodes()
	var validAddr string
	for _, val := range addrs {
		addr := splitAddress(val)
		channel := discover.serverName + ":" + addr
		discover.ripcClient.Notify(channel, const_ask)
		res := discover.ripcClient.Wait(channel, 1*time.Second)
		if res == const_alive {
			validAddr = addr
			break
		}
	}
	return validAddr
}

func (discover *Discover) ListenMainNode(handler func(msg string)) {
	addr := ""
	newAddr := ""
	for {
		newAddr = discover.getMainNodeAddress()
		if newAddr != "" && newAddr != addr {
			addr = newAddr
			handler(addr)
		}
		time.Sleep(2 * time.Second)
	}
}

func splitAddress(address string) string {
	index := strings.Index(address, ":")
	if index == -1 {
		return ""
	}
	hostAndPort := address[index+1:]
	return hostAndPort
}

func getKeysByNamespace(redisClient *redis.Client, namespace string) []string {
	var keys []string
	cursor := uint64(0)
	for {
		// 使用 SCAN 命令获取键名
		res, newCursor, err := redisClient.Scan(context.Background(), cursor, namespace+"*", 0).Result()
		if err != nil {
			return nil
		}
		// 处理键名，去掉命名空间
		for _, key := range res {
			// 去掉命名空间部分
			keyWithoutNamespace := key[len(namespace):]
			keys = append(keys, keyWithoutNamespace[1:])
		}
		cursor = newCursor
		// 如果游标为0，则结束循环
		if cursor == 0 {
			break
		}
	}
	return keys
}
