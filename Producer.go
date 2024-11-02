package researd

import (
	"encoding/binary"
	"net"
	"time"
)

type Producer struct {
	researdClient *Client
	channel       string
	maxRetries    int
	address       string
	conn          net.Conn
}

func newProducer(researdClient *Client, channel string) *Producer {
	producer := Producer{
		researdClient: researdClient,
		maxRetries:    10,
		channel:       channel,
	}
	go producer.listen()
	return &producer
}

func (producer *Producer) connect() error {
	conn, err := net.Dial("tcp", producer.address)
	if err == nil {
		producer.conn = conn
		return nil
	}
	return err
}

func (producer *Producer) listen() {
	searcher := producer.researdClient.NewSearcher(producer.channel)
	searcher.Listen(func(addr string, data map[string]string) {
		producer.address = addr
		producer.connect()
	})
}

func (producer *Producer) SetMaxRetries(maxRetries int) {
	producer.maxRetries = maxRetries
}

func (producer *Producer) Publish(message []byte) error {
	// 设置重试次数限制，避免无限重试
	var err error
	retryCount := 0
	for producer.conn == nil {
		err = producer.connect()
		if err == nil {
			break
		}
		time.Sleep(const_waitTime)
		if retryCount == producer.maxRetries {
			return err
		}
		retryCount++
	}
	retryCount = 0

	byteMessage := []byte(string(message))
	messageLength := uint32(len(byteMessage))

	// 1. 写入消息长度前缀
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, messageLength)

	for {
		// 尝试写入字节流
		_, err = producer.conn.Write(append(lengthBuf, byteMessage...))
		if err != nil {
			err = producer.connect()
			if err != nil {
				time.Sleep(const_waitTime)
			}
		} else {
			return nil
		}
		if retryCount == producer.maxRetries {
			break
		}
		retryCount++
	}
	return err
}
