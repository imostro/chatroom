package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"learn/chatroom/entity"
	"log"
	"net"
	"os"
)

var (
	done   = make(chan error)
	sendCh = make(chan *entity.Message, 128)
	rcvCh  = make(chan *entity.Message, 128)
)

func inputMsg(ctx context.Context) {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print("client ->")
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			if bytes.Equal(line, []byte("quit\n")) {
				done <- errors.New("quit")
			}
			message := entity.Message{
				MsgType: entity.Generate,
				Msg:     line,
			}
			sendCh <- &message
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
			encode, err := entity.Encode(msg)
			if err != nil {
				log.Println("encode error:", err)
			}
			_, err = conn.Write(encode)
			if err != nil {
				log.Println("write error: ", err)
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
			header := make([]byte, entity.MsgHeaderLen)
			n, err := conn.Read(header)
			if err != nil {
				done <- err
				return
			}

			if n < entity.MsgHeaderLen {
				log.Println("msg header format error.")
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
				done <- err
				return
			}
			if n != int(msgLen) {
				log.Println("message format error.")
				continue
			}

			message, err := entity.Decode(data)
			if err != nil {
				log.Fatalln("encode error: ", err)
			}
			switch message.MsgType {
			case entity.Login, entity.Logout:
			case entity.BeatHeat:
				if err != nil {
					log.Fatalf("encode error: %v \n", err)
				}
				sendCh <- message
			case entity.Generate:
				rcvCh <- message
			}
		}
	}
}

func outputMsg(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-rcvCh:
			fmt.Printf("%s -> %s", msg.ClientName, msg.Msg)
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
	go outputMsg(ctx)
	go send(ctx, conn)
	go received(ctx, conn)

	<-done
}

func main() {
	Start()
}
