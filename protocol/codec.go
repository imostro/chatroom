package protocol

import (
	"google.golang.org/protobuf/proto"
)

func Encode(msg *Message) ([]byte, error) {
	msgEncode, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	encode := make([]byte, 0, MsgHeaderLen+len(msgEncode))
	// start
	encode = append(encode, 0x0f)
	// msg len
	encode = append(encode, byte(len(msgEncode)>>0))
	encode = append(encode, byte(len(msgEncode)>>8))
	encode = append(encode, byte(len(msgEncode)>>16))
	encode = append(encode, byte(len(msgEncode)>>24))
	// end
	encode = append(encode, 0xef)

	//	msg
	encode = append(encode, msgEncode...)

	return encode, nil
}

func Decode(data []byte) (*Message, error) {
	msg := &Message{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return msg, err
	}
	return msg, err
}
