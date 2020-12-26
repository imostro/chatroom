package protocol

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

func NewHeatBeatMessage(clientName, clientAddr string) Message {
	return Message{
		MsgType:    BeatHeat,
		ClientName: clientName,
		ClientAddr: clientAddr,
	}
}
