package sub

import (
	"errors"
	"strings"
	"sync"
)

var localRetainTree *retainNode

type message struct {
	topic   string
	payload []byte
	qos     uint8
}
type retainMessage struct {
	messages map[string]*message
	mutex    sync.RWMutex
}

/**
订阅树节点的子节点map
*/
type childRetainNodes struct {
	m  map[string]*retainNode
	mu sync.RWMutex
}

/**
订阅树中的节点
*/
type retainNode struct {
	topicSection string
	ChildNodes   *childRetainNodes
	m            *message
	topic        string
}

func init() {

	localRetainTree = newRetainNode("/")
}
func newRetainNode(topicSection string) *retainNode {
	n := retainNode{
		topicSection: topicSection,
		ChildNodes:   newChildRetainNodes(),
	}

	return &n
}

/**

 */
func newChildRetainNodes() *childRetainNodes {

	n := new(childRetainNodes)
	n.m = make(map[string]*retainNode)
	return n
}

func newRetainMessage() *retainMessage {

	m := new(retainMessage)
	m.messages = make(map[string]*message)
	return m
}

func NewMessage(payload []byte, qos uint8) *message {

	return &message{
		payload: payload,
		qos:     qos,
	}
}

func SetMessage(topic string, m *message) {

	topicSlice := strings.Split(topic, "/")
	queue := make([]*retainNode, 0)
	queue = append(queue, localRetainTree)
	i := 0
	first := queue[0]
	for len(queue) != 0 && i < len(topicSlice) {
		first = queue[0]
		queue = queue[1:]

		first.ChildNodes.mu.Lock()
		if childNode, ok := first.ChildNodes.m[topicSlice[i]]; ok {
			queue = append(queue, childNode)
			i++
			if i == len(topicSlice) {
				childNode.topicSection = topicSlice[i]
				childNode.m = m
			}
		} else {
			if childTree, err := topicSliceBeRetainTree(topicSlice[i:], m); err == nil {
				first.ChildNodes.m[topicSlice[i]] = childTree
			}
		}
		first.ChildNodes.mu.Unlock()
	}

}

/**
Topic 切片转话成树，树的叶子节点存入clientID
*/
func topicSliceBeRetainTree(topicsSlice []string, m *message) (*retainNode, error) {

	if len(topicsSlice) == 0 {
		return nil, errors.New("topicSlice length can not be 0")
	}

	var first *retainNode
	var last *retainNode
	for i, t := range topicsSlice {
		n := newRetainNode(t)
		if i == 0 {
			first = n
		} else {
			last.ChildNodes.m[t] = n
		}
		last = n
	}
	last.m = m
	return first, nil

}

/**
获取topic在订阅树中保存的节点，未找到返回false
*/
func GetMessages(topicSlice []string) map[string]*message {

	//fixme 订阅树搜索不需每次从根节点开始搜索，topicSlice长度不同，从不同的节点开始搜索，提升查询效率
	queue := make([]*retainNode, 0)
	queue = append(queue, localRetainTree)
	messages := make(map[string]*message)
	i := 0
	first := queue[0]
	flag := false
	class := false
	for len(queue) != 0 {
		first = queue[0]
		queue = queue[1:]

		first.ChildNodes.mu.RLock()
		if topicSlice[i] == "#" {
			flag = true
		}
		if topicSlice[i] == "+" {
			class = true
		}
		if flag == true {
			for _, v := range first.ChildNodes.m {
				queue = append(queue, v)
				messages[v.topic] = v.m
			}
		} else if class == true {
			for _, v := range first.ChildNodes.m {
				queue = append(queue, v)
				messages[v.topic] = v.m
			}
			class = false
		} else {

			if childNode, ok := first.ChildNodes.m[topicSlice[i]]; ok || flag {
				queue = append(queue, childNode)
				i++
			}
		}

		first.ChildNodes.mu.RUnlock()
		if i == len(topicSlice) && flag == false && class == false {
			messages[first.topic] = first.m
		}
	}
	return messages
}
