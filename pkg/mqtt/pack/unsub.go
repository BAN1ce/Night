package pack

import "live/pkg/utils"

type UnSub struct {
	pack       *Pack
	Identifier []byte
	Payload    []byte
	Topics     []string
}

func NewUnSubPack(p *Pack) *UnSub {
	s := new(UnSub)
	s.Topics = make([]string, 0)
	s.pack = p
	fixLength := p.FixedHeader.FixLength
	s.Identifier = p.rawData[fixLength : fixLength+2]
	plc := fixLength + 2
	s.Payload = p.rawData[plc:]
	payloadLength := len(s.Payload)

	for i := 0; i+2 < payloadLength; {
		l := utils.UtfLength(s.Payload[i : i+2])
		if (i + 2 + l) <= payloadLength {
			s.Topics = append(s.Topics, string(s.Payload[i+2:i+2+l]))
		} else {
			break
		}
		i = i + 2 + l
	}
	return s

}
