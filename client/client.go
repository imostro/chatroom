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

var done = make(chan error)

func Send(ctx context.Context, conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Print(conn.LocalAddr(), " ->")
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			if bytes.Equal(line, []byte("quit\n")) {
				done <- errors.New("quit")
			}
			message := entity.Message{
				MsgType:    entity.Generate,
				ClientName: conn.LocalAddr().String(),
				ClientAddr: conn.LocalAddr().String(),
				Msg:        line,
			}
			encode, err := entity.Encode(&message)
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

func Received(ctx context.Context, conn net.Conn) {
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
			fmt.Printf("%s -> %s", message.ClientName, message.Msg)
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

	go Send(ctx, conn)
	go Received(ctx, conn)

	<-done
}

func main() {
	Start()
}
