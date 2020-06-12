package sub

import (
	"errors"
	"strings"
	"sync"
)

var localRetainTree *retainNode

/**
retain message
*/
type Message struct {
	Payload []byte
	Qos     uint8
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
	topicSection string            // 子topic
	ChildNodes   *childRetainNodes // 子节点的集合
	m            *Message          // retain message
	topic        string            // 节点的完整topic
	class        int               //表示节点的高度，根节点高度为0
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
创建一个子节点集合
*/
func newChildRetainNodes() *childRetainNodes {

	n := new(childRetainNodes)
	n.m = make(map[string]*retainNode)
	return n
}

/**
创建一个新message
*/
func NewMessage(payload []byte, qos uint8) *Message {

	return &Message{
		Payload: payload,
		Qos:     qos,
	}
}

/**
Topic 设置 Message
*/
func SetMessage(topic string, m *Message) {

	topicSlice := strings.Split(topic, "/")
	topicHigh := len(topicSlice)
	queue := make([]*retainNode, 0)
	queue = append(queue, localRetainTree)
	i := 0
	first := queue[0]
	for len(queue) != 0 && i < topicHigh {
		first = queue[0]
		queue = queue[1:]

		first.ChildNodes.mu.Lock()
		if childNode, ok := first.ChildNodes.m[topicSlice[i]]; ok {
			queue = append(queue, childNode)
			i++
			// 如果子节点已经存在
			if i == topicHigh {
				childNode.topicSection = topicSlice[i-1]
				childNode.topic = topic
				childNode.m = m
				childNode.class = i
			}
		} else {
			// 如果子树不存在则创建一个子树连接到父结点上
			if childTree, err := topicSliceBeRetainTree(topicSlice[i:], m, topic, i); err == nil {
				first.ChildNodes.m[topicSlice[i]] = childTree
			}
		}
		first.ChildNodes.mu.Unlock()
	}

}

/**
创建一个完整或不完整的topic的子树
*/
func topicSliceBeRetainTree(topicsSlice []string, m *Message, topic string, class int) (*retainNode, error) {

	if len(topicsSlice) == 0 {
		return nil, errors.New("topicSlice length can not be 0")
	}

	var first *retainNode
	var last *retainNode
	for i, t := range topicsSlice {
		n := newRetainNode(t)
		class++
		n.class = class
		if i == 0 {
			first = n
		} else {
			last.ChildNodes.m[t] = n
		}
		last = n
	}
	last.m = m
	last.topic = topic
	return first, nil

}

/**
获取订阅topic下的所有retain message
*/
func GetMessages(topicSlice []string) map[string]*Message {

	//fixme 订阅树搜索不需每次从根节点开始搜索，topicSlice长度不同，从不同的节点开始搜索，提升查询效率
	queue := make([]*retainNode, 0)
	queue = append(queue, localRetainTree)
	messages := make(map[string]*Message)
	subTopicHigh := len(topicSlice)
	first := queue[0]
	flag := false
	class := false
	for len(queue) != 0 {
		first = queue[0]
		queue = queue[1:]

		first.ChildNodes.mu.RLock()
		if first.class < subTopicHigh && topicSlice[first.class] == "#" {
			flag = true
		}
		if first.class < subTopicHigh && topicSlice[first.class] == "+" {
			class = true
		}
		if flag == true {
			for _, v := range first.ChildNodes.m {
				queue = append(queue, v)
				if v.m != nil {
					messages[v.topic] = v.m
				}
			}
		} else if class == true {
			for _, v := range first.ChildNodes.m {
				if v.class != subTopicHigh {
					queue = append(queue, v)
				}
				if v.m != nil && v.class == subTopicHigh {
					messages[v.topic] = v.m
				}
			}
			class = false
		} else {

			if childNode, ok := first.ChildNodes.m[topicSlice[first.class]]; ok {
				if childNode.class == subTopicHigh {
					if childNode.m != nil {
						messages[childNode.topic] = childNode.m
					}
				} else {
					queue = append(queue, childNode)
				}
			}
		}

		first.ChildNodes.mu.RUnlock()

	}
	return messages
}
