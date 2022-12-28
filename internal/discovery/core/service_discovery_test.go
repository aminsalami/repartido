package core

import (
	"github.com/aminsalami/repartido/internal/discovery"
	"github.com/aminsalami/repartido/internal/ring"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockedStorage struct {
	db map[string]*discovery.CacheNode
}

func (m *mockedStorage) Save(node *discovery.CacheNode) error {
	m.db[node.Id] = node
	return nil
}

func (m *mockedStorage) Get() (discovery.CacheNode, error) {
	return discovery.CacheNode{}, nil
}

func (m *mockedStorage) List() (result []*discovery.CacheNode, e error) {
	for _, v := range m.db {
		result = append(result, v)
	}
	return
}

func (m *mockedStorage) Close() error {
	return nil
}

func (m *mockedStorage) Delete(node *discovery.CacheNode) error {
	delete(m.db, node.Id)
	return nil
}

func (m *mockedStorage) Clear() {
	m.db = make(map[string]*discovery.CacheNode)
}

// -----------------------------------------------------------------

// Test the behaviour of service when the first node is being registered.
// We expect all virtual-nodes (from 0 to 256) will be assigned to this server!
func TestRegister_ForTheFirstTime(t *testing.T) {
	storage := &mockedStorage{db: make(map[string]*discovery.CacheNode)}
	service := NewCacheService(storage)
	node1 := discovery.CacheNode{
		Id:       "node-1",
		Name:     "node-name-1",
		Host:     "1.1.1.1",
		Port:     8002,
		LastPing: time.Now().Format(time.RFC3339),
		RamSize:  1024,
	}
	err := service.registerNode(&node1)
	assert.NoError(t, err)
	assert.Len(t, storage.db, 1)

	assert.Len(t, service.ring.All(), ring.Size)
	for _, v := range service.ring.All() {
		assert.NotNil(t, v)
	}

	vnodes := service.getVirtualNodes()
	assert.Len(t, vnodes, ring.Size)
	for _, vnode := range service.getVirtualNodes() {
		assert.NotNil(t, vnode)
	}
}

// Test duplicate register/unregister calls
func TestRegister_RegisterUnregister(t *testing.T) {
	storage := &mockedStorage{db: make(map[string]*discovery.CacheNode)}
	service := NewCacheService(storage)
	node1 := discovery.CacheNode{
		Id:       "node-1",
		Name:     "node-name-1",
		Host:     "1.1.1.1",
		Port:     8002,
		LastPing: time.Now().Format(time.RFC3339),
		RamSize:  1024,
	}
	err := service.registerNode(&node1)
	assert.NoError(t, err)
	assert.Len(t, storage.db, 1)
	// Check duplicate register
	err = service.registerNode(&node1)
	assert.Error(t, err)
	assert.Len(t, storage.db, 1)

	err = service.unregisterNode(&node1)
	assert.NoError(t, err)
	assert.Len(t, storage.db, 0)
	assert.Len(t, service.getVirtualNodes(), 0)

	// Try to unregister again
	err = service.unregisterNode(&node1)
	assert.Error(t, err)
}
