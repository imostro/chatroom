package entity

import (
	"fmt"
	"testing"
)

func TestEncode(t *testing.T) {
	message := Message{
		MsgType:    Login,
		ClientName: "127.0.0.1:1212",
		ClientAddr: "gray",
		Msg:        []byte("hehhhhhhh"),
	}
	encode, err := Encode(&message)
	if err != nil {
		fmt.Println("error encode!")
	}
	decode, err := Decode(encode)
	if err != nil {
		fmt.Println("error decode")
	}
	fmt.Println(decode.Msg)
}
