package researd

import (
	"encoding/binary"
	"fmt"
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
	go producer.researdClient.NewSearcher().Listen(producer.channel, func(addr string, data map[string]string) {
		producer.address = addr
		producer.connect()
	})
	return &producer
}

func (producer *Producer) connect() error {
	if producer.address == "" {
		producer.conn = nil
		return fmt.Errorf("no message queue service was found")
	}
	conn, err := net.Dial("tcp", producer.address)
	if err == nil {
		producer.conn = conn
		return nil
	}
	return err
}

func (producer *Producer) SetMaxRetries(maxRetries int) {
	producer.maxRetries = maxRetries
}

func (producer *Producer) Publish(message []byte) error {
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

	//长度前缀协议
	byteMessage := []byte(string(message))
	messageLength := uint32(len(byteMessage))

	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, messageLength)

	for {
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
