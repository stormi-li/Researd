package researd

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	reconfig "github.com/stormi-li/Reconfig"
)

type Client struct {
	reconfig *reconfig.Client
}

func NewClient(addr string) (*Client, error) {
	client := Client{}
	reconfig, err := reconfig.NewClient(addr)
	if err != nil {
		return nil, err
	}
	client.reconfig = reconfig
	return &client, nil
}

const alive = "alive"
const ask = "ask"

func (c *Client) Register(name string, addr string, weight int) {
	cfg := c.reconfig.NewConfig(name+":"+addr, addr)
	cfg.Info.Info = map[string]string{}
	cfg.Info.Info["weight"] = strconv.Itoa(weight)
	go func() {
		for {
			cfg.Upload(30 * time.Second)
			time.Sleep(25 * time.Second)
		}
	}()
	ctx := context.Background()
	go func() {
		c.reconfig.RipcClient.NewListener(ctx, name+addr).Listen(func(msg string) {
			if msg == ask {
				for i := 0; i < 10; i++ {
					c.reconfig.RipcClient.Notify(ctx, name+addr, alive)
					time.Sleep(100 * time.Millisecond)
				}
			}
		})
	}()
	for {
		c.reconfig.RipcClient.Notify(ctx, name+addr, alive)
		time.Sleep(2 * time.Second)
	}
}

const prefix = "stormi:config:"

func (client *Client) Discover(name string) []string {
	names, _ := client.getKeysByNamespace(prefix + name)
	addrs := []string{}
	for _, nn := range names {
		configInfo := client.reconfig.GetConfig(name + nn)
		weight, _ := strconv.Atoi(configInfo.Info["weight"])
		for i := 0; i < weight; i++ {
			addrs = append(addrs, configInfo.Addr)
		}
	}
	shuffleArray(addrs)
	return addrs
}

func (client *Client) getValidAddr(name string) string {
	addrs := client.Discover(name)
	ctx := context.Background()
	var validAddr string
	for _, addr := range addrs {
		client.reconfig.RipcClient.Notify(ctx, name+addr, ask)
		res := client.reconfig.RipcClient.Wait(ctx, name+addr, 1*time.Second)
		if res == alive {
			validAddr = addr
			break
		} else {
			addrs = removeValue(addrs, addr)
		}
	}
	return validAddr
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

func (client *Client) Connect(name string, handler func(addr string)) {
	addr := ""
	ctx := context.Background()
	for {
		if addr == "" {
			addr = client.getValidAddr(name)
			if addr != "" {
				handler(addr)
			}
			time.Sleep(2 * time.Second)
		} else {
			res := client.reconfig.RipcClient.Wait(ctx, name+addr, 5*time.Second)
			if res != alive {
				addr = ""
			}
		}
	}
}

func shuffleArray(arr []string) {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i] // 交换元素
	})
}

func (client *Client) getKeysByNamespace(namespace string) ([]string, error) {
	var keys []string
	cursor := uint64(0)

	for {
		// 使用 SCAN 命令获取键名
		res, newCursor, err := client.reconfig.RipcClient.RedisClient.Scan(context.Background(), cursor, namespace+"*", 0).Result()
		if err != nil {
			return nil, err
		}

		// 处理键名，去掉命名空间
		for _, key := range res {
			// 去掉命名空间部分
			keyWithoutNamespace := key[len(namespace):]
			keys = append(keys, keyWithoutNamespace)
		}

		cursor = newCursor

		// 如果游标为0，则结束循环
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}
