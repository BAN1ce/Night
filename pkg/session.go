package pkg

import (
	"container/list"
	"context"
	"live/pkg/mqtt/pack"
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
	pubWaitCancel   context.CancelFunc

	mutex      sync.RWMutex
	identifier uint16

	SubTopicsQos map[string]uint8

	hasPubTimer bool
	pubTimer    *time.Timer
	offlineTime time.Time
	isExpired   bool
}

func newSession() *session {

	s := new(session)
	s.SubTopicsQos = make(map[string]uint8)
	s.recIdentifies = make(map[string]bool)
	s.isExpired = false
	s.recWaitList = list.New()
	s.pubWaitQueue = list.New()
	s.pubWaitInitTime = config.qos1WaitTime

	return s
}

func (s *session) Expired() {
	s.mutex.Lock()
	s.isExpired = true
	s.mutex.Unlock()
}

func (s *session) ToValid() {
	s.mutex.Lock()
	s.isExpired = false
	s.mutex.Unlock()
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

	// 检查队列的大小，超出进行修剪
	s.pubWaitMutex.Lock()
	if !s.isExpired {
		cfg := config.GetConfig()
		if s.pubWaitQueue.Len() > cfg.maxPubQueue && cfg.maxPubQueue > 0 {
			flag := true
			for p := s.pubWaitQueue.Front(); p != nil; p = p.Next() {
				if pub, ok := p.Value.(*pack.PubPack); ok {
					// 优先删除Qos=0的消息
					if pub.Qos == 0 {
						s.pubWaitQueue.Remove(p)
						flag = false
						break
					}
				}
			}
			if flag {
				s.pubWaitQueue.Remove(s.pubWaitQueue.Front())
			}
		}
		pubPack.Dup = true
		s.pubWaitQueue.PushBack(pubPack)
	}
	s.pubWaitMutex.Unlock()

}

/**
收到发布消息的ack
*/
func (s *session) pubAck(ackPack *pack.PubAckPack) {

	s.pubWaitMutex.Lock()
	if !s.isExpired {
		p := s.pubWaitQueue.Front()
		if p != nil {
			if pub, ok := p.Value.(*pack.PubPack); ok {
				if string(pub.Identifier) == string(ackPack.Identifier) {
					s.pubWaitQueue.Remove(p)
				}
			}
		}
		// 队列为空时结束回收延迟任务的资源
		if s.pubWaitQueue.Len() == 0 {
			s.pubTimer.Stop()
			s.hasPubTimer = false
			s.pubWaitCancel() //退出延时任务的G
		}
	}
	s.pubWaitMutex.Unlock()
}

/**
发布非qos0的消息给客户端时开启延迟任务
*/
func (s *session) RunPubTimer(ctx context.Context, c *Client) {

	ctx, s.pubWaitCancel = context.WithCancel(ctx)
	s.pubWaitMutex.Lock()

	if s.pubWaitQueue.Len() == 0 || s.hasPubTimer == true {
		s.pubWaitTime = s.pubWaitInitTime
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
					s.pubWaitTime = s.pubWaitInitTime
					s.hasPubTimer = false
					return

				case <-s.pubTimer.C:
					s.pubWaitMutex.Lock()
					if s.isExpired {
						return
					}
					if s.pubWaitQueue.Len() > 0 && s.hasPubTimer == true {
						s.pubWaitTime *= 2
						s.pubTimer.Reset(s.pubWaitTime)
						s.PubListPack(c)
						s.pubWaitMutex.Unlock()
					} else {
						s.hasPubTimer = false
						s.pubWaitMutex.Unlock()
						return
					}
				}
			}
		}()
	}
	s.pubWaitMutex.Unlock()
}

/**
发布list中等待ack的消息
*/
func (s *session) PubListPack(c *Client) {

	s.pubWaitMutex.RLock()
	for p := s.pubWaitQueue.Front(); p != nil; p = p.Next() {
		if pub, ok := p.Value.(*pack.PubPack); ok {
			c.write(pub)
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
