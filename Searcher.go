package researd

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Researd/Ripc"
)

type Searcher struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
	ctx         context.Context
	serverName  string
	data        map[string]string
}

func newSearcher(redisClient *redis.Client, ripcClient *ripc.Client, namespace string, serverName string) *Searcher {
	return &Searcher{
		redisClient: redisClient,
		ripcClient:  ripcClient,
		namespace:   namespace,
		serverName:  serverName,
		ctx:         context.Background(),
	}
}

func (discover *Searcher) SearchServer() []string {
	addrs := getKeysByNamespace(discover.redisClient, discover.namespace+discover.serverName)
	sort.Slice(addrs, func(a, b int) bool {
		return addrs[a] < addrs[b]
	})
	return addrs
}
func (discover *Searcher) getMainNodeAddress() string {
	addrs := discover.SearchServer()
	var validAddr string
	if len(addrs) > 0 {
		validAddr = splitAddress(addrs[0])
		data, _ := discover.redisClient.Get(discover.ctx, discover.namespace+discover.serverName+const_separator+addrs[0]).Result()
		json.Unmarshal([]byte(data), &discover.data)
	}
	return validAddr
}

func (discover *Searcher) Listen(handler func(address string, data map[string]string)) {
	addr := ""
	newAddr := ""
	for {
		newAddr = discover.getMainNodeAddress()
		if newAddr != "" && newAddr != addr {
			addr = newAddr
			handler(addr, discover.data)
		}
		time.Sleep(2 * time.Second)
	}
}

func splitAddress(address string) string {
	index := strings.Index(address, const_separator)
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
