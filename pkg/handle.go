package pkg

import (
	"fmt"
	"live/pkg/mqtt/pack"
	"live/pkg/mqtt/sub"
	"live/pkg/utils"
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

	case pack.DISCONNECT:
		disconnect(c, p)

	}

}

/**
客户端连接
*/
func connectHandle(c *Client, p *pack.Pack) {

	connectPack := pack.NewConnectPack(p)
	connack := pack.NewConnAckPack(0)

	c.clientIdentifier = connectPack.ClientIdentifier

	// 客户端连接成功
	c.writeChan <- connack

	fmt.Println("Connecting ... ")

	if connectPack.CleanSession != true {
		if oldClient, ok := getClient(c.clientIdentifier); ok {
			oldClient.Stop()
			// 恢复session中的消息
			c.session = oldClient.session
			c.session.ToValid()
			c.PubStoreSession()
		}
	}

	if connectPack.WillFlag {
		c.SetWill(connectPack.WillTopic, connectPack.WillPayload, connectPack.WillQos)
		if connectPack.WillRetain {
			sub.SetMessage(c.willTopic, sub.NewMessage(c.willPayload, c.willQos))
		}
	}
	clientJoinServer(connectPack.ClientIdentifier, c)

	fmt.Println("Connecting ... End ")
}

/**
客户端发布消息
*/
func pubHandle(c *Client, p *pack.Pack) {
	pubPack := pack.NewPubPack(p)

	nodes, clients := sub.GetSub(string(pubPack.TopicName))

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

	// 记录 retain 消息
	if pubPack.Retain {
		if len(pubPack.Payload) == 0 {
			sub.SetMessage(string(pubPack.TopicName), nil)
		} else {
			sub.SetMessage(string(pubPack.TopicName), sub.NewMessage(pubPack.Payload, pubPack.Qos))
		}
	}

}

/**
客户端回复确认发布消息
*/
func pubAckHandle(c *Client, p *pack.Pack) {

	//todo 从list中删除收到的ack发送消息
	pubAckPack := pack.NewPubAckPack(p)

	c.session.pubAck(pubAckPack)

}

/**
客户端订阅
*/
func subHandle(c *Client, p *pack.Pack) {

	fmt.Println("1")

	subPack := pack.NewSubPack(p)
	qoss := make([]byte, 0, len(subPack.TopicQos))
	fmt.Println("1")
	for topic, qos := range subPack.TopicQos {
		if qos >= 2 {
			qos = 1
		}
		qoss = append(qoss, qos)
		fmt.Println("1")
		c.session.SubTopic(topic, qos)
		topicSlice := strings.Split(topic, "/")
		fmt.Println("2")
		// 客户端模糊订阅和绝对订阅分开记录
		sub.Sub(topic, c.clientIdentifier, topicSlice)
		fmt.Println("1")
		retainMessages := sub.GetMessages(topicSlice)

		fmt.Println("1")
		for topic, message := range retainMessages {
			pubpack := pack.NewEmptyPubPack()
			pubpack.Qos = message.Qos
			pubpack.Payload = message.Payload
			pubpack.TopicName = []byte(topic)
			pubpack.Retain = true
			c.Pub(pubpack, topic)
			fmt.Println("1")
		}
	}
	fmt.Println("1")
	subAck := pack.NewSubAck(subPack.Identifier, qoss)
	fmt.Println("Sub -> ", subAck.Identifier, qoss)
	c.writeChan <- subAck
}

/**
取消订阅
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

/**
断开连接
*/
func disconnect(c *Client, p *pack.Pack) {
	c.DeleteWill()
	c.Stop()

}

func PubPackToSubClients(nodes []*sub.Node, clients []string, pubPack *pack.PubPack) {

	go func() {
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

	}()
}
