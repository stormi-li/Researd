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

func NewClient(redisClient *redis.Client, namespace string, serverType ...ServerType) *Client {
	prefix := const_NodePrefix
	if len(serverType) != 0 && serverType[0] == MQ {
		prefix = const_mqPrefix
	}
	return &Client{
		ripcClient:  ripc.NewClient(redisClient, namespace),
		redisClient: redisClient,
		namespace:   namespace + const_splitChar + prefix,
	}
}

func (c *Client) NewRegister(serverName string, address string) *Register {
	return newRegister(c.redisClient, c.ripcClient, c.namespace, serverName, address)
}

func (c *Client) NewDiscovery(serverName string) *Discovery {
	return newDiscovery(c.redisClient, c.ripcClient, c.namespace, serverName)
}
