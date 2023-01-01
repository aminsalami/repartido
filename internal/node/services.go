package node

import "github.com/aminsalami/repartido/internal/node/ports"

type cacheService struct {
	cache ports.ICache
	id    string
}

func (s *cacheService) getKey(key string) (string, error) {
	return s.cache.Get(key)
}

func (s *cacheService) set(key, data string) error {
	return s.cache.Put(key, data)
}

func (s *cacheService) delete(key string) error {
	return s.cache.Del(key)
}
