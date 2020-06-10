package src

import (
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

	case pack.UNSUBSCRIBE:
		unSubHandle(c, p)
	case pack.PINGREQ:
		pingHandle(c, p)
	}

}

func connectHandle(c *Client, p *pack.Pack) {

	//todo 重名客户端

	connectPack := pack.NewConnectPack(p)
	connack := pack.NewConnAckPack(0)

	c.clientIdentifier = connectPack.ClientIdentifier

	// 客户端连接成功
	c.writeChan <- connack

	if connectPack.CleanSession != true {

		if oldClient, ok := getClient(c.clientIdentifier); ok {
			//todo 恢复session中的消息
			// will message handle
			c.session = oldClient.session
			c.PubStoreSession()
		}
	}
	clientJoinServer(connectPack.ClientIdentifier, c)

}
func pubHandle(c *Client, p *pack.Pack) {
	pubPack := pack.NewPubPack(p)

	topicSlice := strings.Split(string(pubPack.TopicName), "/")
	nodes := sub.GetWildCards(topicSlice)

	clients := sub.GetHashSub(string(pubPack.TopicName))
	if pubPack.Qos == 1 {
		pubAckPack := pack.NewEmptyPubAckPack(pubPack.Identifier)
		// 返回pub ack
		c.writeChan <- pubAckPack
	}
	for _, node := range nodes {
		node.Clients.Mu.RLock()
		for clientIdentifier, _ := range node.Clients.M {
			if c, ok := getClient(clientIdentifier); ok {
				c.Pub(pubPack, node.Topic)
			}
		}
		node.Clients.Mu.RUnlock()
	}

	for _, clientIdentifier := range clients {
		if c, ok := getClient(clientIdentifier); ok {
			c.Pub(pubPack, string(pubPack.TopicName))
		}
	}

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
		qoss = append(qoss, qos)
		c.session.SubTopic(topic, qos)
		topicSlice := strings.Split(topic, "/")
		// 客户端模糊订阅和绝对订阅分开记录
		if utils.HasWildcard(topicSlice) {
			sub.AddTreeSub(topicSlice, c.clientIdentifier)
		} else {
			sub.AddHashSub(topic, c.clientIdentifier)
		}
	}
	subAck := pack.NewSubAck(subPack.Identifier, qoss)
	c.writeChan <- subAck
}

/**
从订阅中删除订阅节点，session中删除topic
*/
func unSubHandle(c *Client, p *pack.Pack) {
	unsubPack := pack.NewUnSubPack(p)

	for _, topic := range unsubPack.Topics {

		topicSlice := strings.Split(topic, "/")

		if utils.HasWildcard(topicSlice) {
			sub.DeleteTreeSub(topicSlice, c.clientIdentifier)
		} else {
			sub.DeleteHashSub(topic, c.clientIdentifier)
		}
		c.session.UnsubTopic(topic)
	}
	unsubAckPack := pack.NewUnSubAck(unsubPack.Identifier)

	c.writeChan <- unsubAckPack

}

/**
心跳请求
*/
func pingHandle(c *Client, p *pack.Pack) {

	pingResp := pack.NewPingResp()

	c.writeChan <- pingResp
}
