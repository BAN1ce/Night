package utils

import "testing"

func TestUint16ToBytes(t *testing.T) {

	var a uint16

	a = 1

	result := Uint16ToBytes(a)

	correct := []byte{0x00,0x01}

	for i, v:=range result{
		if v == correct[i]  {
			continue
		}else {
			t.Error("uint16 to byte fail")
			t.Log(result)
		}
	}
}
