package pack

type unsubAck struct {
	Identifier []byte
}

func NewUnSubAck(identifier []byte) *unsubAck {
	s := new(unsubAck)
	s.Identifier = identifier
	return s
}
func (s *unsubAck) GetFixHeadByte() byte {

	return UNSUBACK << 4
}

func (s *unsubAck) GetVariableHeader() ([]byte, uint32) {

	return s.Identifier, 2
}

func (s *unsubAck) GetPayload() ([]byte, uint32) {

	return nil, 0
}
