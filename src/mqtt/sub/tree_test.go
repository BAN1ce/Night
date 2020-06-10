package sub

import (
	"fmt"
	"strings"
	"testing"
)

/**
空订阅树添加一个订阅和搜索
*/
func TestTreeAddASub(t *testing.T) {

	topic := "product1/device1/get"
	clientId := "123456"
	topics := strings.Split(topic, "/")

	AddTreeSub(topics, clientId)

	node, ok := GetTreeSub(topics)
	if ok {
		flag := true
		for k, _ := range node.Clients.M {
			if k == clientId {
				flag = false
			}
		}
		if flag {
			t.Error("Get Node But Can Find ClientIdentifier")
			t.FailNow()
		}
	} else {
		t.Error("Can Not Find Node")
		t.FailNow()
	}

	DeleteTreeSub(topics, clientId)

	node, ok = GetTreeSub(topics)
	if !ok {
		t.Log("Delete Sub Success")
	} else {
		t.Error("Delete Sub Fail")
	}
}

/**
通配订阅
*/
func TestTreeWildcardSub(t *testing.T) {

	topic := "product/device1/get"
	subTopic := "product/#"
	clientId := "123456"
	topics := strings.Split(topic, "/")
	subTopics := strings.Split(subTopic, "/")
	AddTreeSub(subTopics, clientId)

	nodes := GetWildCards(topics)
	flag := false
	for _, v := range nodes {
		v.Clients.Mu.RLocker()
		for c, _ := range v.Clients.M {
			if c != clientId {
				t.Error("Get Wildcard Sub Fail")
			} else {
				flag = true
				t.Log("Get Sub client", c)
			}
		}
	}
	if !flag {
		t.Error("Get Wildcard Sub Fail")
	}

}

/**
多个客户端订阅不同的topic
*/
func TestTreeAddPluralSub(t *testing.T) {

	clientIds := make([]string, 0)
	topics := make([]string, 0)
	for i := 0; i < 10; i++ {
		topic := fmt.Sprintf("product%d/device%d/get", i, i)
		clientId := fmt.Sprintf("0000%d", i)
		clientIds = append(clientIds, clientId)
		topics = append(topics, topic)
		topicSplit := strings.Split(topic, "/")
		AddTreeSub(topicSplit, clientId)
	}

	for i := 0; i < 10; i++ {
		node, ok := GetTreeSub(strings.Split(topics[i], "/"))
		if ok {

			flag := true
			for k, _ := range node.Clients.M {
				if k == clientIds[i] {
					flag = false
				}
			}
			if flag {
				t.Error("Get Node But Can Find ClientIdentifier")
			}

		} else {
			t.Error("Can Not Find Node")
		}

	}
}

/**
  不同客户端订阅不同的topic，每个topic订阅的客户端数量不同
*/
func TestTreeDiffClientSub(t *testing.T) {
	clientIds := make([]map[string]bool, 0)
	topics := make([]string, 0)
	for i := 0; i < 10; i++ {
		topic := fmt.Sprintf("product%d/device%d/get", i, i)
		cs := make(map[string]bool)
		for j := 0; j < i+1; j++ {
			clientId := fmt.Sprintf("000%d", j)
			cs[clientId] = true
		}
		clientIds = append(clientIds, cs)
		topics = append(topics, topic)
		topicSplit := strings.Split(topic, "/")
		for k, _ := range cs {
			AddTreeSub(topicSplit, k)
		}
	}

	for i := 0; i < 10; i++ {
		topic := fmt.Sprintf("product%d/device%d/get", i, i)
		node, ok := GetTreeSub(strings.Split(topic, "/"))
		if ok {

			flag := true
			for k, _ := range node.Clients.M {
				if _, ok := clientIds[i][k]; ok {
				} else {
					flag = false
				}
			}
			if flag {
				t.Error("Get Node But Can Find ClientIdentifier")
			}

		} else {
			t.Error("Topic can not find sub ", topic)
		}

	}
}
