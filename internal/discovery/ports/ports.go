package ports

import "github.com/aminsalami/repartido/internal/discovery"

type CacheStorage interface {
	Save(*discovery.CacheNode) error
	Get() (discovery.CacheNode, error)
	List() ([]*discovery.CacheNode, error)
	Delete(node *discovery.CacheNode) error
	Close() error
}
