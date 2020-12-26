package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"learn/chatroom/protocol"
	"log"
	"net"
	"os"
	"time"
)

var (
	done   = make(chan error)
	sendCh = make(chan *protocol.Message, 128)
	rcvCh  = make(chan *protocol.Message, 128)
)

func inputMsg(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("->")
			line, err := reader.ReadBytes('\n')
			if err != nil {
				done <- err
				return
			}
			if bytes.Equal(line, []byte("quit\n")) {
				done <- errors.New("quit")
				return
			}

			sendCh <- &protocol.Message{
				MsgType: protocol.Generate,
				Msg:     line,
			}
		}
	}

}

func send(ctx context.Context, conn net.Conn) {

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-sendCh:
			msg.ClientName = conn.LocalAddr().String()
			msg.ClientAddr = conn.LocalAddr().String()
			encode, err := protocol.Encode(msg)
			if err != nil {
				log.Println("protobuf encode error, reason: ", err)
				continue
			}
			_, err = conn.Write(encode)
			if err != nil {
				log.Println("write msg to server error, reason: ", err)
				done <- err
				return
			}
		}
	}
}

func received(ctx context.Context, conn net.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			header := make([]byte, protocol.MsgHeaderLen)
			n, err := conn.Read(header)
			if err != nil {
				if err == io.EOF {
					log.Println("is disconnect.")
				} else {
					log.Println("unknown read error: ", err)
				}
				done <- err
				return
			}

			if n != protocol.MsgHeaderLen {
				log.Println("msg header format is error.")
				continue
			}
			var msgLen int32
			msgLen |= int32(header[1]) << 0
			msgLen |= int32(header[2]) << 8
			msgLen |= int32(header[3]) << 16
			msgLen |= int32(header[4]) << 24

			data := make([]byte, msgLen)
			n, err = conn.Read(data)
			if err != nil {
				if err == io.EOF {
					log.Println("is disconnect.")
				} else {
					log.Println("unknown read error: ", err)
				}
				done <- err
				return
			}
			if n != int(msgLen) {
				log.Println("message format error.")
				continue
			}

			msg, err := protocol.Decode(data)
			if err != nil {
				log.Println("encode error: ", err)
				continue
			}

			rcvCh <- msg
		}
	}
}

func handlerRcvMsg(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-rcvCh:
			switch msg.MsgType {
			case protocol.Login, protocol.Logout:
			case protocol.BeatHeat:
				sendCh <- msg
			case protocol.Generate:
				fmt.Printf("%s ===> %s", msg.ClientName, msg.Msg)
				fmt.Print("->")
			}
		}
	}
}

func Start() {
	conn, err := net.Dial("tcp", "localhost:10086")
	if err != nil {
		fmt.Println("connect server error: ", err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		conn.Close()
	}()

	go inputMsg(ctx)
	go handlerRcvMsg(ctx)
	go send(ctx, conn)
	go received(ctx, conn)
	go simulateSendMsg()
	<-done
}

// 模拟兵法消息
func simulateSendMsg() {
	time.Sleep(10 * time.Second)

	for i := 0; i < 10000; i++ {
		time.Sleep(10 * time.Millisecond)
		sendCh <- &protocol.Message{
			MsgType: protocol.Generate,
			Msg:     []byte("hello kiki\n"),
		}
	}

	done <- errors.New("close")
}
func main() {
	Start()
}
