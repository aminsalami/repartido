package ports

import "github.com/aminsalami/repartido/internal/discovery"

type CacheStorage interface {
	Save(*discovery.CacheNode) error
	Get() (discovery.CacheNode, error)
	GetById(string2 string) (*discovery.CacheNode, error)
	List() ([]*discovery.CacheNode, error)
	Delete(node *discovery.CacheNode) error
	Close() error
}
