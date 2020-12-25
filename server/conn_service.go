package server

import (
	"context"
	"errors"
	"io"
	"learn/chatroom/entity"
	"log"
	"net"
	"time"
)

type ConnServer interface {
	Send(ctx context.Context)
	Received(ctx context.Context)
	HeatBeat(ctx context.Context)
	Close()
}

type Conn struct {
	name       string
	conn       net.Conn
	rcvCh      chan []byte
	sendCh     chan []byte
	done       chan error
	hbTimer    *time.Timer
	hbInterval time.Duration
	hbTimeout  time.Duration
}

// create a new conn service
func NewConn(conn net.Conn) Conn {
	return Conn{
		name:       conn.RemoteAddr().String(),
		conn:       conn,
		rcvCh:      make(chan []byte),
		sendCh:     make(chan []byte),
		done:       make(chan error),
		hbTimer:    &time.Timer{},
		hbInterval: 1 * time.Second,
		hbTimeout:  5 * time.Second,
	}
}

func (c *Conn) SetHeatBeat(hbInterval, hbTimeout time.Duration) {
	c.hbTimeout = hbTimeout
	c.hbInterval = hbInterval
}

func (c *Conn) Send(ctx context.Context) {
	for {
		select {
		case data := <-c.sendCh:
			_, err := c.conn.Write(data)
			if err != nil {
				log.Fatalf("send message to client: %s error, reason: %v.\n", c.name, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Conn) Received(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			header := make([]byte, entity.MsgHeaderLen)
			n, err := c.conn.Read(header)
			if err != nil {
				if err == io.EOF {
					log.Printf("client %s close.\n", c.name)
					return
				}
				log.Fatalf("received client %s msg error, reason: %v\n", c.name, err)
			}
			if n != entity.MsgHeaderLen ||
				header[0] != entity.StartSymbol ||
				header[len(header)-1] != entity.EndSymbol {

				log.Printf("error msg header format: %d.\n", header)
				continue
			}
			var msgLen int32
			msgLen |= int32(header[1]) << 0
			msgLen |= int32(header[2]) << 8
			msgLen |= int32(header[3]) << 16
			msgLen |= int32(header[4]) << 24

			data := make([]byte, msgLen)
			n, err = c.conn.Read(data)
			if err != nil {
				if err == io.EOF {
					log.Printf("client %s close.\n", c.name)
					return
				}
				log.Fatalf("received client %s msg error, reason: %v\n", c.name, err)
			}
			if n != int(msgLen) {
				log.Printf("error msg format.\n")
			}
			data = append(header, data...)
			// 收到客户端数据，验证完毕后发送到广播中
			broadCastCh <- data
		}
	}
}

func (c *Conn) HeatBeat(ctx context.Context) {

}

func (c *Conn) Close() {
	c.done <- errors.New(c.name + "-> close connect")
}

func (c *Conn) sendMsg(message *entity.Message) {

}
