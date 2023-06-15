package node

import "github.com/aminsalami/repartido/internal/node/ports"

type storageService struct {
	storage ports.Storage
}

func (s *storageService) getKey(key string) (string, error) {
	return s.storage.Get(key)
}

func (s *storageService) set(key, data string) error {
	return s.storage.Put(key, data)
}

func (s *storageService) delete(key string) error {
	return s.storage.Del(key)
}
