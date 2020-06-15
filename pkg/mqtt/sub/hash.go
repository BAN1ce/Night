package sub

import (
	"sync"
)

/**
topic的绝对订阅哈希表
*/

var localHashSub *hashSub

func init() {
	localHashSub = newHashSub()
}

/**
一个节点下的订阅关系
*/
type hashSub struct {
	sub   map[string]map[string]bool
	mutex sync.RWMutex
}

func newHashSub() *hashSub {

	n := new(hashSub)
	n.sub = make(map[string]map[string]bool)
	return n
}

/**
一个节点下的客户端订阅一个主题
*/
func AddHashSub(topic, clientIdentifier string) {

	localHashSub.mutex.Lock()

	if _, ok := localHashSub.sub[topic]; ok == false {
		localHashSub.sub[topic] = make(map[string]bool)
	}
	localHashSub.sub[topic][clientIdentifier] = true
	localHashSub.mutex.Unlock()
}

/**
删除某一个节点的客户端订阅
*/
func DeleteHashSub(topic, clientIdentifier string) {

	localHashSub.mutex.Lock()
	if _, ok := localHashSub.sub[topic]; ok {
		delete(localHashSub.sub[topic], clientIdentifier)
	}
	localHashSub.mutex.Unlock()
}

/**
获取不同节点的订阅客户端uuid
*/
func GetHashSub(topic string) []string {

	nodeClients := make([]string, 0)

	localHashSub.mutex.RLock()

	if v, ok := localHashSub.sub[topic]; ok {
		for c, _ := range v {
			nodeClients = append(nodeClients, c)
		}
	}
	localHashSub.mutex.RUnlock()
	return nodeClients
}
