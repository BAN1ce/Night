package src

import (
	"sync"
	"time"
)

var config *Config

func init() {
	config = newConfig()
	config.maxPubQueue = 200
	config.sessionExpiredTime = time.Second * 10
	config.sessionExpireInterval = time.Second * 5

}

type Config struct {
	maxPubQueue           int
	sessionExpiredTime    time.Duration
	sessionExpireInterval time.Duration
	mutex                 sync.RWMutex
}

func newConfig() *Config {
	return new(Config)
}

func (c *Config) GetConfig() *Config {
	c.mutex.RLock()
	tmp := c
	c.mutex.RUnlock()
	return tmp
}
