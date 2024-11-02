package researd

import "time"

const const_updateNodeType = "updateNodeType"
const const_configPrefix = "stormi:config:"
const const_serverPrefix = "stormi:server:"
const const_mqPrefix = "stormi:mq:"
const const_separator = ":"
const node_standby = "standby"
const node_main = "main"

const const_waitTime = 500 * time.Millisecond
const const_expireTime = 2 * time.Second

type ServerType int

const (
	Server ServerType = iota
	MQ
	Config
)

func (s ServerType) String() string {
	switch s {
	case Server:
		return "Server"
	case MQ:
		return "MQ"
	case Config:
		return "Config"
	default:
		return "Unknown"
	}
}
