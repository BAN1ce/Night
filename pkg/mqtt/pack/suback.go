package pack

type subAck struct {
	Identifier []byte
	TopicQos   []byte
}

func NewSubAck(identifier []byte, topicQos []byte) *subAck {
	s := new(subAck)
	s.Identifier = identifier
	s.TopicQos = topicQos
	return s
}
func (s *subAck) GetFixHeadByte() byte {

	return SUBACK << 4
}

func (s *subAck) GetVariableHeader() ([]byte, uint32) {

	return s.Identifier, 2
}

func (s *subAck) GetPayload() ([]byte, uint32) {

	return s.TopicQos, uint32(len(s.TopicQos))
}
