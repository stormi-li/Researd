package researd

import (
	"github.com/go-redis/redis/v8"
	ripc "github.com/stormi-li/Ripc"
)

type Client struct {
	redisClient *redis.Client
	ripcClient  *ripc.Client
	namespace   string
}

func NewClient(redisClient *redis.Client, namespace string, model ...string) *Client {
	prefix := const_registerPrefix
	if len(model) != 0 && model[0] == Model_MQ {
		prefix = const_mqPrefix
	}
	return &Client{
		ripcClient:  ripc.NewClient(redisClient, namespace),
		redisClient: redisClient,
		namespace:   namespace + ":" + prefix,
	}
}

func (c *Client) NewRegister(serverName string, address string) *Register {
	return newRegister(c.redisClient, c.ripcClient, c.namespace, serverName, address)
}

func (c *Client) NewDiscover(serverName string) *Discover {
	return newDiscover(c.redisClient, c.ripcClient, c.namespace, serverName)
}
