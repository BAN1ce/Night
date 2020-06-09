package pack

import (
	"testing"
)

func TestMakeRemain(t *testing.T) {

	result := make([][]byte, 8)
	result[0] = []byte{0x00}
	result[1] = []byte{0x7f}
	result[2] = []byte{0x80, 0x01}
	result[3] = []byte{0xff, 0x7f}
	result[4] = []byte{0x80, 0x80, 0x01}
	result[5] = []byte{0xff, 0xff, 0x7f}
	result[6] = []byte{0x80, 0x80, 0x80, 0x01}
	result[7] = []byte{0xff, 0xff, 0xff, 0x7f}

	text := []uint32{0, 127, 128, 16383, 16384, 2097151, 2097152, 268435455}

	for i, v := range text {
		tmp := makeRemainParameter(v)
		for j := 0; j < len(tmp); j++ {
			if tmp[j] == result[i][j] {
				continue
			} else {
				t.Error("make remain bytes wrong", v,tmp,result[i])
				t.FailNow()
			}
		}
	}

}
