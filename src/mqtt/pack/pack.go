package pack

const (
	RESERVED = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

type Pack struct {
	FixedHeader FixHeader
	rawData     []byte
}

type FixHeader struct {
	ByteOne      uint8
	RemainLength int
	FixLength    int
}

func (f FixHeader) GetCMD() int {
	return int(f.ByteOne >> 4)
}

type WritePack interface {
	GetFixHeadByte() byte
	GetVariableHeader() ([]byte, uint32)
	GetPayload() ([]byte, uint32)
}

func NewPack(data []byte) *Pack {
	pack := new(Pack)
	remainLength, fixHeadLength := getRemainLength(data)

	pack.FixedHeader = FixHeader{
		ByteOne:      data[0],
		RemainLength: remainLength,
		FixLength:    fixHeadLength,
	}
	pack.rawData = data
	return pack
}

func getRemainLength(data []byte) (remainLength int, fixHeadLength int) {
	multiplier := 1
	fixHeadLength = 1
	remainLength = 0
	for i := 1; i < len(data) && i <= 4; i++ {
		remainLength += int(data[i]&127) * multiplier
		fixHeadLength++
		if data[i]&128 == 128 {
			multiplier *= 128
		} else {
			break
		}

	}
	return
}

/**

 */
func Encode(p WritePack) []byte {

	payload, pl := p.GetPayload()
	variableHeader, vl := p.GetVariableHeader()

	totalLength := pl + vl
	remain := makeRemainParameter(totalLength)

	responseData := make([]byte, 0, totalLength+1+uint32(len(remain)))
	responseData = append(responseData, p.GetFixHeadByte())
	responseData = append(responseData, remain...)
	if vl != 0 {
		responseData = append(responseData, variableHeader...)
	}
	if pl != 0 {
		responseData = append(responseData, payload...)
	}
	return responseData

}

func makeRemainParameter(remainLength uint32) []byte {
	remain := make([]byte, 0, 4)
	if remainLength == 0 {
		remain = append(remain, 0x00)
	} else {

		for i := 0; i < 4; i++ {
			digit := remainLength % 128
			if digit > 0 {
				if remainLength/128 > 0 {
					digit = digit | 0x80
				}
				remainLength /= 128
				remain = append(remain, uint8(digit))
			} else {
				break
			}
		}
	}
	return remain

}
