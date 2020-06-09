package pack

import "live/src/utils"

type SubPack struct {
	pack       *Pack
	Identifier []byte
	Payload    []byte
	TopicQos   map[string]uint8
}

func NewSubPack(p *Pack) *SubPack {

	s := new(SubPack)
	s.TopicQos = make(map[string]uint8)
	s.pack = p
	fixLength := p.FixedHeader.FixLength
	s.Identifier = p.rawData[fixLength : fixLength+2]
	plc := fixLength + 2
	s.Payload = p.rawData[plc:]
	for len(p.rawData) > plc+2 {
		l := utils.UtfLength(p.rawData[plc : plc+2])
		plc += 2
		if plc+l < len(p.rawData) {
			topic := string(p.rawData[plc : plc+l])
			plc += l
			qos := p.rawData[plc]
			plc += 1
			s.TopicQos[topic] = qos
		}
	}
	return s
}
