package src

import (
	"container/list"
	"context"
	"fmt"
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

	pubWaitMutex sync.RWMutex
	pubWaitQueue *list.List

	mutex      sync.RWMutex
	identifier uint16

	SubTopicsQos map[string]uint8

	hasPubTimer bool
	pubTimer    *time.Timer
	pubWaitTime time.Duration
}

func newSession() *session {
	s := new(session)
	s.SubTopicsQos = make(map[string]uint8)
	s.recIdentifies = make(map[string]bool)
	s.recWaitList = list.New()
	s.pubWaitQueue = list.New()

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
		s.pubWaitTime = 5 * time.Second
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
					fmt.Println("timer exec")
					if s.pubWaitQueue.Len() > 0 && s.hasPubTimer == true {
						s.pubWaitTime *= 2
						s.pubTimer.Reset(s.pubWaitTime)
						s.PubListPack(c)
						s.pubWaitMutex.RUnlock()
					} else {
						s.hasPubTimer = false
						s.pubWaitMutex.RUnlock()
						fmt.Println("timer close")
						return
					}
				}
			}
		}()
	}
	s.pubWaitMutex.Unlock()
}

func (s *session) PubListPack(c *Client) {

	fmt.Println("pub list", time.Now().Format("2006-01-02 15:04:05"), c.clientIdentifier)
	s.pubWaitMutex.RLock()
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
