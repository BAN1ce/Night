package src

import (
	"bufio"
	"fmt"
)





/**
mqtt包拆包
*/
func Input() bufio.SplitFunc {

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {

		if atEOF {
			return
		}
		fmt.Println(data,"data")
		if len(data) >= 2 {
			headLength := 1
			multiplier := 1
			bodyLength := 0
			for i := 1; i < len(data) && i <= 4; i++ {
				bodyLength += int(data[i]&127) * multiplier
				headLength++
				if data[i]&128 == 128 {
					multiplier *= 128
				} else {
					break
				}
			}
			sum := headLength + bodyLength
			if sum > len(data) {
				return 0, nil, nil
			}
			token = make([]byte, sum)
			for i := 0; i < len(token); i++ {
				token[i] = data[i]
			}
			return sum, token, nil
		}

		return
	}
}
