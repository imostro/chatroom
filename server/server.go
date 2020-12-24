package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"learn/chatroom/common"
	"net"
)

func main() {
	startService()
}

func startService() {
	listener, err := net.Listen("tcp", ":10086")
	if err != nil {
		return
	}
	for true {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		fmt.Printf("client %s connect\n", conn.RemoteAddr().String())
		go Handler(conn)
	}
}

func Handler(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		srv := []byte{0: 0}
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("client %s close conneect.", conn.RemoteAddr())
					conn.Close()
				}
				return
			}
			srv = append(srv, buffer[:n]...)
			if n != len(buffer) {
				break
			}
		}
		message := common.Message{}
		err := proto.Unmarshal(srv, &message)
		if err != nil {
			return
		}
		fmt.Printf("client %v send msg: ", message.Msg)
		_, err = conn.Write(srv)
		if err != nil {
			return
		}
	}
}
