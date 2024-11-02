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
	data        map[string]string
}

func newSearcher(redisClient *redis.Client, ripcClient *ripc.Client, namespace string) *Searcher {
	return &Searcher{
		redisClient: redisClient,
		ripcClient:  ripcClient,
		namespace:   namespace,
		ctx:         context.Background(),
	}
}

func (searcher *Searcher) SearchAllServers(serverName string) []string {
	addrs := getKeysByNamespace(searcher.redisClient, searcher.namespace+serverName)
	sort.Slice(addrs, func(a, b int) bool {
		return addrs[a] < addrs[b]
	})
	return addrs
}
func (searcher *Searcher) GetHighestPriorityServer(serverName string) (string, map[string]string) {
	addrs := searcher.SearchStartingServers(serverName)
	var validAddr string
	if len(addrs) > 0 {
		validAddr = split(addrs[0])[1]
		data, _ := searcher.redisClient.Get(searcher.ctx, searcher.namespace+serverName+const_separator+addrs[0]).Result()
		json.Unmarshal([]byte(data), &searcher.data)
	}
	return validAddr, searcher.data
}

func (searcher *Searcher) Listen(serverName string, handler func(address string, data map[string]string)) {
	addr := ""
	jsonByte, _ := json.MarshalIndent(searcher.data, " ", "  ")
	dataStr := string(jsonByte)
	for {
		newAddr, data := searcher.GetHighestPriorityServer(serverName)
		jsonByte, _ = json.MarshalIndent(data, " ", "  ")
		newDataStr := string(jsonByte)
		if newAddr != addr || newDataStr != dataStr {
			addr = newAddr
			dataStr = newDataStr
			handler(addr, searcher.data)
		}
		time.Sleep(2 * time.Second)
	}
}

func (searcher *Searcher) SearchStartingServers(serverName string) []string {
	servers := searcher.SearchAllServers(serverName)
	startingservers := []string{}
	for _, val := range servers {
		temp := split(val)
		if temp[0] == state_start {
			startingservers = append(startingservers, temp[1])
		}
	}
	return startingservers
}

func split(address string) []string {
	index := strings.Index(address, const_separator)
	if index == -1 {
		return nil
	}
	return []string{address[:index], address[index+1:]}
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
