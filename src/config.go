package src

import (
	"sync"
	"time"
)

var config *Config

func init() {
	config = newConfig()
	config.maxPubQueue = 200
	config.sessionExpiredTime = 1440 * time.Minute
	config.sessionExpireInterval = 120 * time.Second
	config.qos1WaitTime = 10 * time.Second

}

type Config struct {
	maxPubQueue           int
	qos1WaitTime          time.Duration
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
