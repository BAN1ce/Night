package pack

import "live/pkg/utils"

type ConnectPack struct {
	pack             *Pack
	BodyStart        int
	ProtocolName     string
	ProtocolLevel    int
	ConnectFlags     byte
	Reserved         int
	CleanSession     bool
	WillFlag         bool
	WillQos          uint8
	WillRetain       bool
	PasswordFlag     bool
	UserNameFlag     bool
	ClientIdentifier string
	WillTopic        string
	WillPayload      []byte
	UserName         string
	Password         string
	KeepAlive        int
}

func NewConnectPack(p *Pack) *ConnectPack {
	connectPack := new(ConnectPack)
	connectPack.pack = p

	fixLength := p.FixedHeader.FixLength
	plc := utils.UtfLength(p.rawData[fixLength:fixLength+2]) + 2

	connectPack.ProtocolName = string(p.rawData[fixLength+2 : fixLength+plc])
	connectPack.ProtocolLevel = int(p.rawData[fixLength+plc])
	connectPack.ConnectFlags = p.rawData[fixLength+plc+1]
	connectPack.Reserved = int(connectPack.ConnectFlags & 1)
	connectPack.CleanSession = int((connectPack.ConnectFlags&2)>>1) == 1
	connectPack.WillFlag = int((connectPack.ConnectFlags&4)>>2) == 1
	if (connectPack.ConnectFlags&24)>>3 == 0 {
		connectPack.WillQos = 0
	} else {
		connectPack.WillQos = 1
	}
	connectPack.WillRetain = int((connectPack.ConnectFlags&32)>>5) == 1
	connectPack.PasswordFlag = int((connectPack.ConnectFlags&64)>>6) == 1
	connectPack.UserNameFlag = int((connectPack.ConnectFlags&128)>>7) == 1
	connectPack.KeepAlive = utils.UtfLength(p.rawData[fixLength+plc+2 : fixLength+plc+4])
	connectPack.BodyStart = plc + fixLength + 4
	plc = connectPack.BodyStart
	clientIDLength := utils.UtfLength(p.rawData[plc : plc+2])
	plc += 2
	connectPack.ClientIdentifier = string(p.rawData[plc : plc+clientIDLength])
	plc += clientIDLength
	if connectPack.WillFlag == true {
		willTopicLength := utils.UtfLength(p.rawData[plc : plc+2])
		plc += 2
		connectPack.WillTopic = string(p.rawData[plc : willTopicLength+plc])
		plc += willTopicLength
		willMessageLength := utils.UtfLength(p.rawData[plc : plc+2])
		plc += 2
		connectPack.WillPayload = p.rawData[plc : willMessageLength+plc]
		plc += willMessageLength
	}
	if connectPack.UserNameFlag == true {
		userNameLength := utils.UtfLength(p.rawData[plc : plc+2])
		plc += 2
		connectPack.UserName = string(p.rawData[plc : plc+userNameLength])
		plc += userNameLength
	}
	if connectPack.PasswordFlag == true {
		passwordLength := utils.UtfLength(p.rawData[plc : plc+2])
		plc += 2
		connectPack.Password = string(p.rawData[plc : plc+passwordLength])
		plc += passwordLength
	}

	return connectPack

}
