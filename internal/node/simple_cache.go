package node

import (
	"fmt"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------

var logger *zap.Logger

// SimpleCache implements ICache
type SimpleCache struct {
	data map[string]string
}

// -----------------------------------------------------------------

func init() {
	logger, _ = zap.NewDevelopment()
}

func NewSimpleCache() ICache {
	return &SimpleCache{
		make(map[string]string),
	}
}

func (c *SimpleCache) Get(key string) (string, error) {
	value, ok := c.data[key]
	if !ok {
		return "", fmt.Errorf("key %s is not available", key)
	}
	return value, nil
}

func (c *SimpleCache) Put(key, value string) error {
	c.data[key] = value
	return nil
}

func (c *SimpleCache) Del(key string) error {
	delete(c.data, key)
	return nil
}
