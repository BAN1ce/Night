package sub

import (
	"fmt"
	"strings"
	"testing"
)

func TestSetMessage(t *testing.T) {

	topics := make([]string, 10000, 10000)
	messages := make([]*message, 1000)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			for k := 0; k < 10; k++ {
				topics[i] = fmt.Sprintf("%d/%d/%d", i, j, k)
				messages[i] = NewMessage([]byte(topics[i]), 0)
			}
		}
	}
	for i, topic := range topics {
		SetMessage(topic, messages[i])
	}
	for _, topic := range topics {
		m := GetMessages(strings.Split(topic, "/"))

		for i, v := range m {
			fmt.Println(i, v)
		}

	}

}
