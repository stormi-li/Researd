package researd

const const_ask = "ask"
const const_alive = "alive"
const const_updateNodeType = "updateNodeType"
const const_NodePrefix = "stormi:node:"
const const_mqPrefix = "stormi:mq:"

type ServerType int

const (
	Node ServerType = iota
	MQ
)

func (s ServerType) String() string {
	switch s {
	case Node:
		return "Node"
	case MQ:
		return "MQ"
	default:
		return "Unknown"
	}
}

type NodeType int

const (
	Main NodeType = iota
	Standby
)

func (s NodeType) String() string {
	switch s {
	case Main:
		return "Main"
	case Standby:
		return "Standby"
	default:
		return "Unknown"
	}
}
