package adaptors

import (
	"errors"
	"fmt"
	"github.com/aminsalami/repartido/internal/node/ports"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------

var logger = zap.NewExample().Sugar()

// SimpleCache implements node.ICache
type SimpleCache struct {
	data map[string]string
}

// -----------------------------------------------------------------

func NewSimpleCache() ports.ICache {
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
	logger.Debugw("New key/value", "key", key, "value", value)
	return nil
}

func (c *SimpleCache) Del(key string) error {
	if _, exists := c.data[key]; !exists {
		return errors.New("key not found")
	}
	delete(c.data, key)
	return nil
}
