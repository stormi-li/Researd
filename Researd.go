package researd

import (
	"context"
	"math/rand/v2"
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

func NewClient(redisClient *redis.Client, namespace string) *Client {
	return &Client{
		ripcClient:  ripc.NewClient(redisClient, namespace),
		redisClient: redisClient,
		namespace:   namespace + ":",
		ctx:         context.Background(),
	}
}

const alive = "alive"
const ask = "ask"

const registerPrefix = "stormi:register:"

func (c *Client) Register(name string, addr string, weight int) {
	key := c.namespace + registerPrefix + name + ":" + addr + ":" + strconv.Itoa(weight)
	go func() {
		for {
			c.redisClient.Set(c.ctx, key, "", 30*time.Second)
			time.Sleep(25 * time.Second)
		}
	}()
	go func() {
		c.ripcClient.NewListener(name + addr).Listen(func(msg string) {
			if msg == ask {
				for i := 0; i < 10; i++ {
					c.ripcClient.Notify(name+addr, alive)
					time.Sleep(100 * time.Millisecond)
				}
			}
		})
	}()
	for {
		c.ripcClient.Notify(name+addr, alive)
		time.Sleep(2 * time.Second)
	}
}

func (c *Client) getAddrs(name string) []string {
	names := getKeysByNamespace(c.redisClient, c.namespace+registerPrefix+name)
	addrs := []string{}
	for _, name := range names {
		addr, weight := splitAddress(name)
		for i := 0; i < weight; i++ {
			addrs = append(addrs, addr)
		}
	}
	shuffleArray(addrs)
	return addrs
}

func (client *Client) getValidAddr(name string) string {
	addrs := client.getAddrs(name)
	var validAddr string
	for _, addr := range addrs {
		client.ripcClient.Notify(name+addr, ask)
		res := client.ripcClient.Wait(name+addr, 1*time.Second)
		if res == alive {
			validAddr = addr
			break
		} else {
			addrs = removeValue(addrs, addr)
		}
	}
	return validAddr
}

func (client *Client) Discover(name string, handler func(addr string)) {
	addr := ""
	for {
		if addr == "" {
			addr = client.getValidAddr(name)
			if addr != "" {
				handler(addr)
			}
			time.Sleep(2 * time.Second)
		} else {
			res := client.ripcClient.Wait(name+addr, 5*time.Second)
			if res != alive {
				addr = ""
			}
		}
	}
}

func splitAddress(address string) (string, int) {
	index := strings.LastIndex(address, ":")

	// 如果没有找到冒号，返回错误
	if index == -1 {
		return "", 0
	}

	// 分割成前部分和后部分
	hostAndPort := address[:index] // 冒号前的部分
	numberStr := address[index+1:] // 冒号后的部分

	// 将冒号后的部分转换为 int
	num, err := strconv.Atoi(numberStr)
	if err != nil {
		return "", 0
	}

	return hostAndPort, num
}

func removeValue(arr []string, value string) []string {
	result := []string{}
	for _, val := range arr {
		if val != value {
			result = append(result, value)
		}
	}
	return result
}

func shuffleArray(arr []string) {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i] // 交换元素
	})
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
