package pack

import (
	"live/src/utils"
)

type PubPack struct {
	pack       *Pack
	Dup        bool
	Qos        uint8
	Retain     bool
	TopicName  []byte
	Identifier []byte
	Payload    []byte
}

func NewEmptyPubPack() *PubPack {

	pubPack := new(PubPack)
	return pubPack
}
func NewPubPack(p *Pack) *PubPack {

	pubPack := new(PubPack)
	pubPack.pack = p
	pubPack.Dup = (p.FixedHeader.ByteOne & 0x01) == 1
	pubPack.Qos = p.FixedHeader.ByteOne & 0x06 >> 1
	pubPack.Retain = (p.FixedHeader.ByteOne & 0x08) == 1
	plc := p.FixedHeader.FixLength
	topicLength := utils.UtfLength(p.rawData[plc : plc+2])
	plc += 2
	pubPack.TopicName = p.rawData[plc : plc+topicLength]
	plc += topicLength
	if pubPack.Qos > 0 {
		pubPack.Identifier = p.rawData[plc : plc+2]
	}
	pubPack.Payload = p.rawData[plc:]
	return pubPack
}

func (p *PubPack) GetFixHeadByte() byte {
	c := uint8(PUBLISH << 4)
	if p.Dup {
		c = c | 0x08
	}
	if p.Retain {
		c = c | 0x01
	}

	c = c | (uint8(p.Qos) << 1)

	return c
}

func (p PubPack) GetVariableHeader() ([]byte, uint32) {

	variableHeader := make([]byte, 0, 30)

	topicLength := len(p.TopicName)
	mlsb := utils.Uint16ToBytes(uint16(topicLength))
	variableHeader = append(variableHeader, mlsb...)

	variableHeader = append(variableHeader, p.TopicName...)
	if p.Qos > 0 {
		variableHeader = append(variableHeader, p.Identifier...)
	}

	return variableHeader, uint32(len(variableHeader))
}

func (p PubPack) GetPayload() ([]byte, uint32) {
	return p.Payload, uint32(len(p.Payload))
}
