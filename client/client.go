package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"learn/chatroom/common"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:10086")
	if err != nil {
		return
	}

	buffer := make([]byte, 1024)
	reader := bufio.NewReader(os.Stdin)
	for {
		br, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		br = br[:len(br)-1]
		if len(br) == 0 {
			continue
		}

		if bytes.Equal(br, []byte("quit")) {
			conn.Close()
			break
		}

		// Marshal
		send := &common.Message{
			Msg:     br,
			MsgID:   time.Now().UnixNano(),
			MsgSize: int32(len(br)),
		}
		encoded, err := proto.Marshal(send)
		if err != nil {
			return
		}
		_, err = conn.Write(encoded)

		// 读取服务端数据
		if err != nil {
			return
		}
		msgByte := make([]byte, 0, 1024)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				return
			}
			msgByte = append(msgByte, buffer[:n]...)
			if n < len(buffer) {
				break
			}
		}
		message := common.Message{}
		err = proto.Unmarshal(msgByte, &message)
		if err != nil {
			return
		}
		fmt.Println(string(message.Msg))
	}
}
