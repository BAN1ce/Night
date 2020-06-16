package pkg

import (
	"bufio"
	"context"
	"github.com/google/uuid"
	"live/pkg/mqtt/pack"
	"live/pkg/mqtt/sub"
	"live/pkg/utils"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	conn             net.Conn
	isOnline         bool
	isStop           bool
	clientIdentifier string
	writeChan        chan pack.WritePack
	readChan         chan *pack.Pack
	uuid             uuid.UUID
	clientDone       chan<- string
	hbTimeout        time.Duration
	mutex            sync.RWMutex
	cancel           context.CancelFunc
	protocol         bufio.SplitFunc
	UserName         string
	session          *session
	ctx              context.Context
	hasWill          bool
	willTopic        string
	willPayload      []byte
	willQos          uint8
}

func NewClient(conn net.Conn, clientDone chan<- string) *Client {
	c := new(Client)
	c.conn = conn
	c.writeChan = make(chan pack.WritePack, 10)
	c.readChan = make(chan *pack.Pack, 10)
	c.clientDone = clientDone
	c.uuid = uuid.New()
	c.isOnline = true
	c.isStop = true
	c.session = newSession()
	c.hbTimeout = 90 * time.Second
	return c
}
func (c *Client) SetWill(willTopic string, willPayload []byte, willQos uint8) {
	c.mutex.Lock()
	c.willTopic = willTopic
	c.hasWill = true
	c.willPayload = willPayload
	c.willQos = willQos
	c.mutex.Unlock()
}

func (c *Client) DeleteWill() {

	c.mutex.Lock()
	c.hasWill = false
	c.willQos = 0
	c.willPayload = nil
	c.willTopic = ""
	c.mutex.Unlock()
}

func (c *Client) copySession(s *session) {
	c.session = s
}
func (c *Client) Run(ctx context.Context) {

	c.mutex.Lock()
	if c.isStop == true {
		//on connect event
		c.ctx, c.cancel = context.WithCancel(ctx)
		go c.input(c.ctx)
		go c.handleRead(c.ctx)
		go c.handleWrite(c.ctx)

		c.isStop = false
	}
	c.mutex.Unlock()
}

func (c *Client) Stop() {
	c.mutex.Lock()
	if c.isStop {
		c.mutex.Unlock()
		return
	} else {
		c.isStop = true
		c.cancel()
		c.isOnline = false
		c.conn.Close()
		c.session.offlineTime = time.Now()
		// 如果有遗嘱消息，发布遗嘱消息
		if c.hasWill {
			nodes, clients := sub.GetSub(c.willTopic)
			pubpack := pack.NewEmptyPubPack()
			pubpack.TopicName = []byte(c.willTopic)
			pubpack.Payload = c.willPayload
			pubpack.Qos = c.willQos
			PubPackToSubClients(nodes, clients, pubpack)
		}
		close(c.writeChan)
		close(c.readChan)
		c.mutex.Unlock()
	}
	// 发送遗嘱消息
}

func (c *Client) input(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return

		default:
			scanner := bufio.NewScanner(c.conn)
			scanner.Split(Input())
			for scanner.Scan() {
				c.readChan <- pack.NewPack(scanner.Bytes())
				if c.hbTimeout > 0 {
					err := c.conn.SetReadDeadline(time.Now().Add(c.hbTimeout))
					if err != nil {
						continue
					}
				}
			}
			if err := scanner.Err(); err != nil {
				c.Stop()
			}
		}
	}
}

func (c *Client) handleRead(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return

		case p, ok := <-c.readChan:

			if ok {
				go handle(c, p)
				// client onMessage event
			} else {
				return
			}
		}
	}
}

func (c *Client) handleWrite(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():

			return
		case m, ok := <-c.writeChan:

			if ok {
				c.mutex.RLock()
				if c.isOnline {
					if _, err := c.conn.Write(pack.Encode(m)); err != nil {
						log.Println(err)
					} else {
						// 发送成功
					}
				}
				c.mutex.RUnlock()
			}



		}
	}
}

func (c *Client) write(writePack pack.WritePack) {

	c.mutex.RLock()

	if !c.isStop {
		c.writeChan <- writePack
	}
	c.mutex.RUnlock()
	return

}

func (c *Client) Pub(pubPack *pack.PubPack, subTopic string) {

	emptyPubPack := getEmptyPub()
	emptyPubPack.TopicName = pubPack.TopicName
	emptyPubPack.Payload = pubPack.Payload
	emptyPubPack.Qos = c.session.getTopicQos(subTopic)
	c.mutex.RLock()
	if emptyPubPack.Qos != 0 {
		emptyPubPack.Identifier = utils.Uint16ToBytes(c.session.GetNewIdentifier())
		c.session.PushPubQueue(emptyPubPack, c)
		// 发送消息给客户端，如果客户端在线则延时检查是否ack
		if c.isOnline {
			c.writeChan <- emptyPubPack
			c.session.RunPubTimer(c.ctx, c)
		}
	} else {
		if c.isOnline {
			c.writeChan <- emptyPubPack
		}
	}
	c.mutex.RUnlock()

	putEmptyPub(emptyPubPack)
}

func (c *Client) PubStoreSession() {
	c.session.PubListPack(c)
	c.session.RunPubTimer(c.ctx, c)
}
