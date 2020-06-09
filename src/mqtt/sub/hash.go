package sub

import (
	"sync"
)

/**
topic的绝对订阅哈希表
*/

var clusterSub *clusterHashSub

func init() {
	clusterSub = new(clusterHashSub)
	clusterSub.nodesHashSub = make(map[string]*nodeHashSub)
}

/**
一个节点下的订阅关系
*/
type nodeHashSub struct {
	nodeName string
	sub      map[string]map[string]bool
	mutex    sync.RWMutex
}

/**
集群下所有节点的订阅关系
*/
type clusterHashSub struct {
	nodesHashSub map[string]*nodeHashSub
	mutex        sync.Mutex
}

func newNodeHashSub(nodeName string) *nodeHashSub {

	n := new(nodeHashSub)
	n.nodeName = nodeName
	n.sub = make(map[string]map[string]bool)
	return n
}

/**
一个节点下的客户端订阅一个主题
*/
func AddHashSub(topic, nodeName, clientIdentifier string) {

	var nodeHashSub *nodeHashSub
	var ok bool
	if nodeHashSub, ok = clusterSub.nodesHashSub[nodeName]; ok == false {
		clusterSub.mutex.Lock()
		if nodeHashSub, ok = clusterSub.nodesHashSub[nodeName]; ok == false {
			clusterSub.nodesHashSub[nodeName] = newNodeHashSub(nodeName)
			nodeHashSub = clusterSub.nodesHashSub[nodeName]
		}
		clusterSub.mutex.Unlock()
	}

	nodeHashSub.mutex.Lock()
	nodeHashSub.sub[topic][clientIdentifier] = true
	nodeHashSub.mutex.Unlock()

}

/**
删除某一个节点的客户端订阅
*/
func DelHash(topic, nodeName , clientIdentifier string) {
	var nodeHashSub *nodeHashSub
	var ok bool
	if nodeHashSub, ok = clusterSub.nodesHashSub[nodeName]; ok == false {
		clusterSub.mutex.Lock()
		if nodeHashSub, ok = clusterSub.nodesHashSub[nodeName]; ok == true {
			clusterSub.mutex.Unlock()
			nodeHashSub.mutex.Lock()
			delete(nodeHashSub.sub[topic], clientIdentifier)
			nodeHashSub.mutex.Unlock()
		}
	} else {
		nodeHashSub.mutex.Lock()
		delete(nodeHashSub.sub[topic], clientIdentifier)
		nodeHashSub.mutex.Unlock()
	}

}

/**
获取不同节点的订阅客户端uuid
*/
func GetHashSub(topic string) map[string][]string {

	nodeClients := make(map[string][]string)

	for nodeName, hash := range clusterSub.nodesHashSub {
		nodeClients[nodeName] = make([]string, 0, 20)
		hash.mutex.RLock()
		for id, _ := range hash.sub[topic] {
			nodeClients[nodeName] = append(nodeClients[nodeName], id)
		}
		hash.mutex.RUnlock()

	}

	return nodeClients
}

/**
获取一个节点下的topic订阅客户端标识符
*/
func GetANodeHashSub(topic, nodeName string) []string {
	nodeClients := make([]string, 0)
	if nodeSub, ok := clusterSub.nodesHashSub[nodeName]; ok {
		nodeSub.mutex.RLock()
		for clientIdentifier, _ := range nodeSub.sub[topic] {
			nodeClients = append(nodeClients, clientIdentifier)
		}
		nodeSub.mutex.RUnlock()
	}

	return nodeClients
}
