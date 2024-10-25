package researd

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	reconfig "github.com/stormi-li/Reconfig"
	ripc "github.com/stormi-li/Ripc"
)

type Client struct {
	reconfigClient *reconfig.Client
	redisClient    *redis.Client
	ripcClient     *ripc.Client
	Namespace      string
	Context        context.Context
}

func NewClient(redisClient *redis.Client) *Client {
	ripcClient := ripc.NewClient(redisClient)
	reconfigClient := reconfig.NewClient(redisClient)
	return &Client{reconfigClient: reconfigClient, ripcClient: ripcClient, redisClient: redisClient, Namespace: "", Context: reconfigClient.Context}
}

func (c *Client) SetNamespace(namespace string) {
	c.reconfigClient.SetNamespace(namespace)
	c.ripcClient.SetNamespace(namespace)
	c.Namespace = namespace + ":"
}

const alive = "alive"
const ask = "ask"

func (c *Client) Register(name string, addr string, weight int) {
	cfg := c.reconfigClient.NewConfig(name+":"+addr, addr)
	cfg.Info.Data = map[string]string{}
	cfg.Info.Data["weight"] = strconv.Itoa(weight)
	go func() {
		for {
			cfg.Upload(30 * time.Second)
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

const prefix = "stormi:config:"

func (c *Client) getAddrs(name string) []string {
	names := reconfig.GetKeysByNamespace(c.redisClient, c.Namespace+prefix+name)
	addrs := []string{}
	for _, nn := range names {
		configInfo := c.reconfigClient.GetConfig(name + nn)
		weight, _ := strconv.Atoi(configInfo.Data["weight"])
		for i := 0; i < weight; i++ {
			addrs = append(addrs, configInfo.Addr)
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
