package utils

import (
	"strconv"
	"unsafe"
)

func Uint16ToBytes(n uint16) []byte {
	return []byte{
		byte(n >> 8),
		byte(n),
	}
}
func BytesToUint16(array []byte) uint16 {
	var data uint16 = 0
	for i := 0; i < len(array); i++ {
		data = data + uint16(uint(array[i])<<uint(8*i))
	}

	return data
}

func Int2Byte(data int) (ret []byte) {
	var len uintptr = unsafe.Sizeof(data)
	ret = make([]byte, len)
	var tmp int = 0xff
	var index uint = 0
	for index = 0; index < uint(len); index++ {
		ret[index] = byte((tmp << (index * 8) & data) >> (index * 8))
	}
	return ret
}

func ConvertToBin(num int) string {
	s := ""

	if num == 0 {
		return "0"
	}

	// num /= 2 每次循环的时候 都将num除以2  再把结果赋值给 num
	for ; num > 0; num /= 2 {
		lsb := num % 2
		// strconv.Itoa() 将数字强制性转化为字符串
		s = strconv.Itoa(lsb) + s
	}
	return s

}

/**
utf 计算字节长度
*/
func UtfLength(l []byte) int {

	return int(l[0])*512 + int(l[1])
}

func getUTFSB(length uint16) {

}
