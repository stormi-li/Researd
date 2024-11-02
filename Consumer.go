package researd

import (
	"encoding/binary"
	"net"
	"strings"
	"sync"
)

type Consumer struct {
	researdClient *Client
	channel       string
	address       string
	listener      net.Listener
	messageChan   chan []byte
	buffer        [][]byte
	bufferLock    sync.Mutex
}

func newConsumer(researdCLient *Client, channel string, address string) *Consumer {
	listener, err := net.Listen("tcp", ":"+strings.Split(address, ":")[1])
	if err != nil {
		panic(err)
	}
	consumer := Consumer{
		researdClient: researdCLient,
		channel:       channel,
		address:       address,
		listener:      listener,
		messageChan:   make(chan []byte, 1000),
		buffer:        [][]byte{},
		bufferLock:    sync.Mutex{},
	}
	go consumer.startListen()
	return &consumer
}

func (consumer *Consumer) SetCapacity(capacity int) {
	consumer.messageChan = make(chan []byte, capacity)
}

func (consumer *Consumer) startListen() {
	for {
		conn, err := consumer.listener.Accept()
		if err != nil {
			continue
		}
		go consumer.handleConnection(conn)
	}
}

func (consumer *Consumer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 用于存放拼接的分块数据，支持连续读取消息
	tempBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}

		// 将读取的数据追加到缓存
		tempBuffer = append(tempBuffer, buffer[:n]...)

		// 循环解析缓存中的消息
		for {
			// 1. 检查是否有足够的字节来读取消息长度前缀（4字节）
			if len(tempBuffer) < 4 {
				break // 不足以读取长度前缀，等待更多数据
			}

			// 2. 读取消息长度前缀
			messageLength := binary.BigEndian.Uint32(tempBuffer[:4])
			totalLength := 4 + int(messageLength) // 总消息长度=长度前缀+消息体

			// 3. 检查缓存中是否有完整的消息
			if len(tempBuffer) < totalLength {
				break // 不足以读取完整消息体，等待更多数据
			}

			// 4. 提取完整的消息体
			messageBuf := tempBuffer[4:totalLength]
			// 发送完整消息到消息通道，或放入缓冲区
			consumer.bufferLock.Lock()
			for len(consumer.buffer) > 0 {
				flag := false
				select {
				case consumer.messageChan <- consumer.buffer[0]: // 非阻塞写入
					// 发送成功后删除缓冲区中的消息
					consumer.buffer = consumer.buffer[1:]
				default:
					flag = true
				}
				if flag {
					break
				}
			}
			select {
			case consumer.messageChan <- messageBuf:
			default:
				consumer.buffer = append(consumer.buffer, messageBuf)
			}
			consumer.bufferLock.Unlock()

			// 5. 从缓存中移除已处理的消息
			tempBuffer = tempBuffer[totalLength:]
		}
	}
}

func (consumer *Consumer) StartOnMain(handler func(data []byte)) {
	register := consumer.researdClient.NewRegister(consumer.channel, consumer.address)
	go register.StartOnMain(map[string]string{})
	consumer.start(handler)
}

func (consumer *Consumer) StartOnStandby(handler func(data []byte)) {
	register := consumer.researdClient.NewRegister(consumer.channel, consumer.address)
	go register.StartOnStandby(map[string]string{})
	consumer.start(handler)
}

func (consumer *Consumer) start(handler func(message []byte)) {
	for {
		msg := <-consumer.messageChan
		consumer.bufferLock.Lock()
		if len(consumer.buffer) > 0 {
			select {
			case consumer.messageChan <- consumer.buffer[0]: // 非阻塞写入
				// 发送成功后删除缓冲区中的消息
				consumer.buffer = consumer.buffer[1:]
			default:
			}
		}
		consumer.bufferLock.Unlock()
		handler(msg)
	}
}
