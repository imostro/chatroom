package entity

type MsgType int32

const MsgHeaderLen = 6
const (
	StartSymbol = 0x0f
	EndSymbol   = 0xef
)

const (
	Login int32 = iota
	Generate
	BeatHeat
	Logout
)

type MessageProtocol struct {
	msgLen  int32
	msgBody []byte
}

func NewMessage(msgBody []byte) MessageProtocol {
	return MessageProtocol{
		msgBody: msgBody,
	}
}
