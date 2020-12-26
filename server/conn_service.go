package server

import (
	"context"
	"learn/chatroom/protocol"
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
	rcvCh   chan *protocol.Message
	sendCh  chan *protocol.Message
	done    chan error
	hbTimer *time.Timer
}

// create a new conn service
func NewConn(conn net.Conn) Conn {
	c := Conn{
		name:   conn.RemoteAddr().String(),
		conn:   conn,
		rcvCh:  make(chan *protocol.Message, 128),
		sendCh: make(chan *protocol.Message, 128),
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
			case protocol.Login, protocol.Logout:
			case protocol.Generate, protocol.BeatHeat:
				data, err := protocol.Encode(msg)
				_, err = c.conn.Write(data)
				if err != nil {
					log.Printf("send message to client: %s error, reason: %v.\n", c.name, err)
					c.Close()
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

			header := make([]byte, protocol.MsgHeaderLen)
			n, err := c.conn.Read(header)
			if err != nil {
				log.Printf("received client %s msg error, reason: %v\n", c.name, err)
				c.done <- err
				return
			}
			if n != protocol.MsgHeaderLen ||
				header[0] != protocol.StartSymbol ||
				header[len(header)-1] != protocol.EndSymbol {

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
				log.Printf("received client %s msg error, reason: %v\n", c.name, err)
				c.done <- err
			}
			if n != int(msgLen) {
				log.Printf("error msg format.\n")
			}
			msg, err := protocol.Decode(data)
			if err != nil {
				log.Printf("protobuf decode error: %v \n", err)
			}

			switch msg.MsgType {
			case protocol.Login, protocol.Logout:
			case protocol.BeatHeat:
			case protocol.Generate:
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
			hbMsg := protocol.NewHeatBeatMessage(c.name, c.conn.RemoteAddr().String())
			c.sendCh <- &hbMsg
			c.hbTimer.Reset(hbInterval)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Conn) Close() {
	c.conn.Close()
}
