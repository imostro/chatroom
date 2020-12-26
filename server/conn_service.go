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

var (
	hbInterval = 1 * time.Second
	hbTimeout  = 5 * time.Second
)

type ConnServer interface {
	Send(ctx context.Context)
	Received(ctx context.Context)
	HeatBeat(ctx context.Context)
	Close()
}

type Conn struct {
	name    string
	conn    net.Conn
	rcvCh   chan *entity.Message
	sendCh  chan *entity.Message
	done    chan error
	hbTimer *time.Timer
}

// create a new conn service
func NewConn(conn net.Conn) Conn {
	c := Conn{
		name:   conn.RemoteAddr().String(),
		conn:   conn,
		rcvCh:  make(chan *entity.Message, 128),
		sendCh: make(chan *entity.Message, 128),
		done:   make(chan error),
	}
	c.hbTimer = time.NewTimer(hbInterval)

	return c
}

func (c *Conn) Send(ctx context.Context) {
	for {
		select {
		case msg := <-c.sendCh:
			switch msg.MsgType {
			case entity.Login, entity.Logout:
			case entity.Generate, entity.BeatHeat:
				data, err := entity.Encode(msg)
				_, err = c.conn.Write(data)
				if err != nil {
					log.Fatalf("send message to client: %s error, reason: %v.\n", c.name, err)
				}
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
			c.conn.SetDeadline(time.Now().Add(hbTimeout))

			header := make([]byte, entity.MsgHeaderLen)
			n, err := c.conn.Read(header)
			if err != nil {
				log.Printf("received client %s msg error, reason: %v\n", c.name, err)
				c.Close()
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
			msg, err := entity.Decode(data)
			if err != nil {
				log.Printf("protobuf decode error: %v \n", err)
			}

			switch msg.MsgType {
			case entity.Login, entity.Logout:
			case entity.BeatHeat:
			case entity.Generate:
				// 收到客户端数据，验证完毕后发送到广播中
				broadCastCh <- msg
			}
		}
	}
}

func (c *Conn) HeatBeat(ctx context.Context) {
	for true {
		select {
		case <-c.hbTimer.C:
			hbMsg := entity.NewHeatBeatMessage(c.name, c.conn.RemoteAddr().String())
			c.sendCh <- &hbMsg
			c.hbTimer.Reset(hbInterval)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Conn) Close() {
	c.done <- errors.New(c.name + "-> close connect")
}
