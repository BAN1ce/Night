package pack

type PingResp struct {
}

func NewPingResp() *PingResp {

	return new(PingResp)
}
func (p *PingResp) GetFixHeadByte() byte {
	return 0xd0
}

func (p *PingResp) GetVariableHeader() ([]byte, uint32) {
	return nil, 0
}

func (p *PingResp) GetPayload() ([]byte, uint32) {
	return nil, 0
}
