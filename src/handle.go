package src

import (
	"fmt"
	"live/src/cluster"
	"live/src/mqtt/pack"
	"live/src/mqtt/sub"
	"live/src/utils"
	"strings"
)

func handle(c *Client, p *pack.Pack) {

	switch p.FixedHeader.GetCMD() {
	case pack.CONNECT:
		connectHandle(c, p)

	case pack.PUBLISH:
		pubHandle(c, p)

	case pack.PUBACK:
		pubAckHandle(c, p)

	case pack.SUBSCRIBE:
		subHandle(c, p)

	case pack.PINGREQ:
		pingHandle(c, p)

	}

}

func connectHandle(c *Client, p *pack.Pack) {

	//todo 重名客户端

	fmt.Println("Connect Handle")
	connectPack := pack.NewConnectPack(p)
	connack := pack.NewConnAckPack(0)

	c.clientIdentifier = connectPack.ClientIdentifier
	if connectPack.CleanSession != true {

		if oldClient, ok := getClient(c.clientIdentifier); ok {
			//todo 恢复session中的消息
			c.session = oldClient.session
			fmt.Println("Get old session")
			c.PubStoreSession()
		}
	}
	clientJoinServer(connectPack.ClientIdentifier, c)
	// 客户端连接成功
	c.writeChan <- connack
}
func pubHandle(c *Client, p *pack.Pack) {
	pubPack := pack.NewPubPack(p)

	fmt.Println("pubpack", pubPack, string(pubPack.TopicName))

	topicSlice := strings.Split(string(pubPack.TopicName), "/")
	nodes := sub.GetWildCards(topicSlice)

	nodeClients := sub.GetHashSub(string(pubPack.TopicName))
	if pubPack.Qos == 1 {
		pubAckPack := pack.NewEmptyPubAckPack(pubPack.Identifier)
		// 返回pub ack
		c.writeChan <- pubAckPack
		for _, node := range nodes {
			node.Clients.Mu.RLock()

			for clientIdentifier, _ := range node.Clients.M {
				fmt.Println("Wildcard pub", clientIdentifier)
				if c, ok := getClient(clientIdentifier); ok {
					c.Pub(pubPack)
				}
			}
			node.Clients.Mu.RUnlock()
		}

		for nodeName, clients := range nodeClients {
			if nodeName == LocalServer.name {
				for _, clientIdentifier := range clients {
					if c, ok := getClient(clientIdentifier); ok {
						c.Pub(pubPack)
					}
				}
			}
		}
	}

	//todo 转发给用户
}

func pubAckHandle(c *Client, p *pack.Pack) {

	//todo 从list中删除收到的ack发送消息
	pubAckPack := pack.NewPubAckPack(p)

	c.session.pubAck(pubAckPack)

}
func subHandle(c *Client, p *pack.Pack) {

	subPack := pack.NewSubPack(p)
	qoss := make([]byte, 0, len(subPack.TopicQos))
	for topic, qos := range subPack.TopicQos {
		if qos >= 2 {
			qos = 1
		}
		// 添加订阅记录
		// todo 广播给集群订阅关系
		qoss = append(qoss, qos)
		c.session.SubTopicsQos[topic] = qos
		topicSlice := strings.Split(topic, "/")
		// 客户端模糊订阅和绝对订阅分开记录
		if utils.HasWildcard(topicSlice) {
			fmt.Println("sub", c.clientIdentifier)
			sub.AddTreeSub(topicSlice, c.clientIdentifier)
		} else {
			sub.AddHashSub(topic, cluster.LocalNodeName, c.clientIdentifier)
		}
	}
	subAck := pack.NewSubAck(subPack.Identifier, qoss)
	c.writeChan <- subAck

}

/**
心跳请求
*/
func pingHandle(c *Client, p *pack.Pack) {

	pingResp := pack.NewPingResp()

	c.writeChan <- pingResp
}
