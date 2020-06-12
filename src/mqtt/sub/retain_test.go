package sub

import (
	"fmt"
	"strings"
	"testing"
)

func setMessage() {
	topics := make([]string, 1000, 10000)
	messages := make([]*Message, 1000)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			for k := 0; k < 10; k++ {
				index := i*100 + j*10 + k
				topics[index] = fmt.Sprintf("%d/%d/%d", i, j, k)
				messages[index] = NewMessage([]byte(topics[index]), 0)
			}
		}
	}
	for i, topic := range topics {
		SetMessage(topic, messages[i])
	}
}

/**
1条retain消息
*/
func TestAMessage(t *testing.T) {
	topic := "a/b/c"

	message := NewMessage([]byte("hello"), 1)

	SetMessage(topic, message)

	messages := GetMessages(strings.Split(topic, "/"))
	for tpc, _ := range messages {
		if tpc != topic {
			t.Error("Can not find a retain message")
		}
	}

}

/**
多条retain消息，各自订阅
*/
func TestSetMessage(t *testing.T) {

	topics := make([]string, 1000, 10000)
	messages := make([]*Message, 1000)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			for k := 0; k < 10; k++ {
				index := i*100 + j*10 + k
				topics[index] = fmt.Sprintf("%d/%d/%d", i, j, k)
				messages[index] = NewMessage([]byte(topics[index]), 0)
			}
		}
	}
	for i, topic := range topics {
		SetMessage(topic, messages[i])
	}
	for index, topic := range topics {
		m := GetMessages(strings.Split(topic, "/"))
		flag := false
		for i, v := range m {
			if i == topic {
				if messages[index] == v {
					flag = true
				}
			}
		}
		if !flag {
			t.Error("Retain FAIL", topic)
		}
		flag = false
	}
}

/**
多条retain消息，通配符订阅
*/
func TestSubWildcardMessage(t *testing.T) {

	SetMessage("a/b/c", nil)
	test := make(map[string]int)

	test["1/1/1"] = 1

	test["+/1/1"] = 10
	test["1/+/1"] = 10
	test["1/1/+"] = 10

	test["+/+/1"] = 100
	test["1/+/+"] = 100
	test["+/1/+"] = 100

	test["+/+/+"] = 1000

	test["1/#"] = 100
	test["1/1/#"] = 10
	test["#"] = 1000


	for i, v := range test {
		m := GetMessages(strings.Split(i, "/"))
		if len(m) == v {
			continue
		} else {
			t.Error(i, v, len(m), "+ Retain Fail")
		}

	}

}

func TestSetANilMessage(t *testing.T) {

	topic := "a/b/c"

	SetMessage(topic, nil)

	m := GetMessages(strings.Split("a/b/c", "/"))

	if m[topic] != nil {
		t.Error("Set A Nil Retain Fail", m[topic])
	}

}

/**
批量删除topic retain消息
*/
func TestSetNilMessage(t *testing.T) {
	topics := make([]string, 1000, 10000)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			for k := 0; k < 10; k++ {
				index := i*100 + j*10 + k
				topics[index] = fmt.Sprintf("%d/%d/%d", i, j, k)
			}
		}
	}
	for _, topic := range topics {
		SetMessage(topic, nil)
	}
	m := GetMessages(strings.Split("#", "/"))
	if len(m) != 0 {
		t.Error("Set Retain Nil Fail", len(m))

	}
}
