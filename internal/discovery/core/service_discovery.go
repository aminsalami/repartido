package core

import (
	"github.com/aminsalami/repartido/internal/discovery"
	"github.com/aminsalami/repartido/internal/discovery/ports"
	"sync"
)

var once1 sync.Once

type cacheService struct {
	cacheNodes []*discovery.CacheNode
	storage    ports.CacheStorage
}

// cs is a singleton cacheService
var cs cacheService

// ------------------------------------------------------
func NewCacheService(storage ports.CacheStorage) *cacheService {
	once1.Do(func() {
		cs = cacheService{
			cacheNodes: make([]*discovery.CacheNode, 0, 8),
			storage:    storage,
		}
	})
	return &cs
}

func (service *cacheService) Ping(node *discovery.CacheNode) error {
	return nil
}

func (service *cacheService) registerNode(node discovery.CacheNode) error {
	// TODO validate node.host is parsable
	// TODO validate port is open (try to ping the cache server)
	if err := service.storage.Save(node); err != nil {
		return err
	}
	service.cacheNodes = append(service.cacheNodes, &node)
	return nil
}

func (service *cacheService) unregisterNode(node discovery.CacheNode) error {
	return nil
}

func (service *cacheService) listNodes() []*discovery.CacheNode {
	return service.cacheNodes
}

//--------------------------------------------
