package pkg

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type OnServerStart func(s *Server) error

var LocalServer *Server

type Server struct {
	name          string
	port          int
	isStop        bool
	clients       map[string]*Client
	clientDone    chan string
	listener      net.Listener
	Protocol      bufio.SplitFunc
	mutex         sync.RWMutex
	onServerStart OnServerStart
	hbTimeout     time.Duration
}

func NewServer(addr net.Addr) *Server {

	var err error
	server := new(Server)
	server.isStop = true
	server.listener, err = net.Listen(addr.Network(), addr.String())
	server.clients = make(map[string]*Client)
	server.clientDone = make(chan string, 1000)

	if err != nil {
		panic("Server Listen on " + addr.String() + " FAIL" + err.Error())
	}

	fmt.Println(`

      ___                       ___           ___           ___     
     /\__\          ___        /\  \         /\__\         /\  \    
    /::|  |        /\  \      /::\  \       /:/  /         \:\  \   
   /:|:|  |        \:\  \    /:/\:\  \     /:/__/           \:\  \  
  /:/|:|  |__      /::\__\  /:/  \:\  \   /::\  \ ___       /::\  \ 
 /:/ |:| /\__\  __/:/\/__/ /:/__/_\:\__\ /:/\:\  /\__\     /:/\:\__\
 \/__|:|/:/  / /\/:/  /    \:\  /\ \/__/ \/__\:\/:/  /    /:/  \/__/
     |:/:/  /  \::/__/      \:\ \:\__\        \::/  /    /:/  /     
     |::/  /    \:\__\       \:\/:/  /        /:/  /    /:/  /     
     /:/  /      \/__/        \::/  /        /:/  /    /:/  /             
     \/__/                     \/__/         \/__/     \/__/            

	`)

	fmt.Println("Server Listen on " + addr.String() + " SUCCESS")

	server.Run(context.Background())

	LocalServer = server
	return server
}

func (s *Server) Run(ctx context.Context) {

	s.mutex.Lock()
	if s.isStop == false {
		fmt.Println("Server is Running")
	} else {

		go func() {
			for {
				conn, err := s.listener.Accept()
				if err != nil {
					log.Println(err)
					continue
				} else {
					go s.acceptClient(ctx, conn)
				}

			}
		}()

	}

	s.mutex.Unlock()
}

func (s *Server) acceptClient(ctx context.Context, conn net.Conn) {

	c := NewClient(conn, s.clientDone)
	c.Run(ctx)

}

/**
客户端成连接服务，connect命令相应成功
*/
func clientJoinServer(clientIdentifier string, client *Client) {

	LocalServer.mutex.Lock()
	LocalServer.clients[clientIdentifier] = client
	LocalServer.mutex.Unlock()
}

func getClient(clientIdentifier string) (*Client, bool) {

	LocalServer.mutex.RLock()
	if c, ok := LocalServer.clients[clientIdentifier]; ok {
		LocalServer.mutex.RUnlock()
		return c, true
	} else {
		LocalServer.mutex.RUnlock()
		return nil, false
	}
}

func (s *Server) CheckSessionExpired() {

	cfg := config.GetConfig()
	t := time.NewTicker(cfg.sessionExpireInterval)
	for {
		<-t.C
		cfg := config.GetConfig()
		for _, c := range s.clients {
			e := time.Duration(time.Now().Unix()-c.session.offlineTime.Unix()) * time.Second
			if e > cfg.sessionExpiredTime {
				c.session = nil
				c.session.Expired()
			}

		}

	}

}
