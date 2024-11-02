package researd

import (
	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Researd/Ripc"
)

type Client struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
	serverType  ServerType
}

func NewClient(redisClient *redis.Client, namespace string, serverType ServerType) *Client {
	prefix := ""
	if serverType == Server {
		prefix = const_serverPrefix
	}
	if serverType == MQ {
		prefix = const_mqPrefix
	}
	if serverType == Config {
		prefix = const_configPrefix
	}
	return &Client{
		ripcClient:  ripc.NewClient(redisClient, namespace),
		redisClient: redisClient,
		namespace:   namespace + const_separator + prefix,
		serverType:  serverType,
	}
}

func (c *Client) GetRipc() *ripc.Client {
	return c.ripcClient
}

func (c *Client) NewRegister(serverName string, address string) *Register {
	return newRegister(c.redisClient, c.ripcClient, c.namespace, serverName, address)
}

func (c *Client) NewSearcher() *Searcher {
	return newSearcher(c.redisClient, c.ripcClient, c.namespace)
}

func (c *Client) NewConsumer(channel string, address string) *Consumer {
	if c.serverType != MQ {
		panic("server type must be mq")
	}
	return newConsumer(c, channel, address)
}

func (c *Client) NewProducer(channel string) *Producer {
	if c.serverType != MQ {
		panic("server type must be mq")
	}
	return newProducer(c, channel)
}

func (c *Client) NewRouter() *Router {
	return newRouter(c.NewSearcher())
}
