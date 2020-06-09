package pack

type ConnAckPack struct {
	ackSign    uint8
	returnCode uint8
}

func NewConnAckPack(returnCode int) *ConnAckPack {

	c := new(ConnAckPack)
	c.ackSign = 0 & 0
	c.returnCode = 0 | c.returnCode
	return c
}

func (c *ConnAckPack) GetFixHeadByte() byte {
	return 0x20
}

func (c *ConnAckPack) GetVariableHeader() ([]byte, uint32) {
	tmp := make([]byte, 2)
	tmp[0] = c.ackSign
	tmp[1] = c.returnCode
	return tmp, 2
}

func (c *ConnAckPack) GetPayload() ([]byte, uint32) {
	return nil, 0
}
