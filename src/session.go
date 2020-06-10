package src

import (
	"container/list"
	"context"
	"live/src/mqtt/pack"
	"sync"
	"time"
)

type session struct {
	clientIdentifier string
	recWaitMutex     sync.RWMutex
	recWaitList      *list.List

	/**
	收到消息标识符的可用性
	*/
	recIdentifies      map[string]bool
	recIdentifierMutex sync.RWMutex

	pubWaitMutex    sync.RWMutex
	pubWaitQueue    *list.List
	pubWaitInitTime time.Duration
	pubWaitTime     time.Duration

	mutex      sync.RWMutex
	identifier uint16

	SubTopicsQos map[string]uint8

	hasPubTimer bool
	pubTimer    *time.Timer
}

func newSession() *session {
	s := new(session)
	s.SubTopicsQos = make(map[string]uint8)
	s.recIdentifies = make(map[string]bool)
	s.recWaitList = list.New()
	s.pubWaitQueue = list.New()
	s.pubWaitInitTime = 2 * time.Second

	return s
}

func (s *session) Close() {

}

func (s *session) GetNewIdentifier() uint16 {
	s.mutex.Lock()
	if (s.identifier | 0x00) == (1<<16 - 1) {
		s.identifier = 0
	}
	tmp := s.identifier + 1
	s.mutex.Unlock()
	return tmp
}

func (s *session) PushPubQueue(pubPack *pack.PubPack, c *Client) {

	//todo 检查队列的大小，超出进行修剪
	s.pubWaitMutex.Lock()
	pubPack.Dup = true
	s.pubWaitQueue.PushBack(pubPack)
	s.pubWaitMutex.Unlock()

}

/**
收到发布消息的ack
*/
func (s *session) pubAck(ackPack *pack.PubAckPack) {

	s.pubWaitMutex.Lock()
	p := s.pubWaitQueue.Front()
	if pub, ok := p.Value.(*pack.PubPack); ok {
		if string(pub.Identifier) == string(ackPack.Identifier) {
			s.pubWaitQueue.Remove(p)
		}
	}
	s.pubWaitMutex.Unlock()
}

/**
发布非qos0的消息给客户端时开启延迟任务
*/
func (s *session) RunPubTimer(ctx context.Context, c *Client) {

	s.pubWaitMutex.Lock()

	if s.pubWaitQueue.Len() == 0 || s.hasPubTimer == true {
		s.pubWaitTime = 5 * time.Second
		s.pubTimer.Reset(s.pubWaitTime)
		s.pubWaitMutex.Unlock()
		return
	} else {

		s.hasPubTimer = true
		s.pubWaitTime = s.pubWaitInitTime
		s.pubTimer = time.NewTimer(s.pubWaitTime)
		go func() {
			for {
				select {
				case <-ctx.Done():
					s.pubWaitTime = 5 * time.Second
					s.hasPubTimer = false
					return

				case <-s.pubTimer.C:
					s.pubWaitMutex.RLock()
					if s.pubWaitQueue.Len() > 0 && s.hasPubTimer == true {
						s.pubWaitTime *= 2
						s.pubTimer.Reset(s.pubWaitTime)
						s.PubListPack(c)
						s.pubWaitMutex.RUnlock()
					} else {
						s.hasPubTimer = false
						s.pubWaitMutex.RUnlock()
						return
					}
				}
			}
		}()
	}
	s.pubWaitMutex.Unlock()
}

func (s *session) PubListPack(c *Client) {

	s.pubWaitMutex.RLock()
	// 队列中有效性时重置定时器
	if s.pubWaitQueue.Len() > 0 {
		s.pubTimer.Reset(s.pubWaitInitTime)
	}
	for p := s.pubWaitQueue.Front(); p != nil; p = p.Next() {
		if pub, ok := p.Value.(*pack.PubPack); ok {
			c.writeChan <- pub
		}
	}
	s.pubWaitMutex.RUnlock()
}

func (s *session) getTopicQos(topic string) uint8 {
	if qos, ok := s.SubTopicsQos[topic]; ok {
		return qos
	}
	return 0
}

func (s *session) SubTopic(topic string, qos uint8) {
	s.mutex.Lock()
	s.SubTopicsQos[topic] = qos
	s.mutex.Unlock()
}

func (s *session) UnsubTopic(topic string) {
	s.mutex.Lock()
	delete(s.SubTopicsQos, topic)
	s.mutex.Unlock()

}
