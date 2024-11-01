package researd

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Ripc"
)

type Client struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
	ctx         context.Context
}

const registerPrefix = "stormi:register:"

func NewClient(redisClient *redis.Client, namespace string) *Client {
	return &Client{
		ripcClient:  ripc.NewClient(redisClient, namespace),
		redisClient: redisClient,
		namespace:   namespace + ":" + registerPrefix,
		ctx:         context.Background(),
	}
}

const alive = "alive"
const ask = "ask"

func (c *Client) Register(name string, weight int, addr string) {
	key := c.namespace + name + ":" + strconv.Itoa(weight) + ":" + addr
	go func() {
		for {
			c.redisClient.Set(c.ctx, key, "", 30*time.Second)
			time.Sleep(15 * time.Second)
		}
	}()
	c.ripcClient.NewListener(name + addr).Listen(func(msg string) {
		if msg == ask {
			for i := 0; i < 10; i++ {
				c.ripcClient.Notify(name+addr, alive)
				time.Sleep(100 * time.Millisecond)
			}
		}
	})
}

func (client *Client) getHighestWeightAddr(name string) string {
	addrs := client.getSortedAddrs(name)
	var validAddr string
	for _, val := range addrs {
		addr := splitAddress(val)
		client.ripcClient.Notify(name+addr, ask)
		res := client.ripcClient.Wait(name+addr, 1*time.Second)
		if res == alive {
			validAddr = addr
			break
		}
	}
	return validAddr
}

func (c *Client) Discover(name string, handler func(addr string)) {
	addr := ""
	newAddr := ""
	for {
		newAddr = c.getHighestWeightAddr(name)
		if newAddr != "" && newAddr != addr {
			addr = newAddr
			handler(addr)
		}
		time.Sleep(2 * time.Second)
	}
}

func (c *Client) getSortedAddrs(name string) []string {
	addrs := getKeysByNamespace(c.redisClient, c.namespace+name)
	sort.Slice(addrs, func(a, b int) bool {
		return addrs[a] > addrs[b]
	})
	return addrs
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
