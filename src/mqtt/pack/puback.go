package pack

type PubAckPack struct {
	Identifier []byte
}

func NewPubAckPack(p *Pack) *PubAckPack {
	pubAckPack := new(PubAckPack)

	pubAckPack.Identifier = p.rawData[p.FixedHeader.FixLength : p.FixedHeader.FixLength+2]

	return pubAckPack
}
func NewEmptyPubAckPack(identifier []byte) *PubAckPack {
	pubAckPack := new(PubAckPack)
	pubAckPack.Identifier = identifier
	return pubAckPack
}

func (p *PubAckPack) GetFixHeadByte() byte {
	return PUBACK << 4
}

func (p *PubAckPack) GetVariableHeader() ([]byte, uint32) {
	return p.Identifier, 2
}

func (p *PubAckPack) GetPayload() ([]byte, uint32) {
	return nil, 0
}
