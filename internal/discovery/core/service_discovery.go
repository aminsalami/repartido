package core

import (
	"errors"
	"github.com/aminsalami/repartido/internal/discovery"
	"github.com/aminsalami/repartido/internal/discovery/ports"
	"github.com/aminsalami/repartido/internal/ring"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"time"
)

type cacheService struct {
	ring    *ring.Ring[*discovery.CacheNode]
	storage ports.CacheStorage
}

// cs is a singleton cacheService
var cs cacheService

func init() {
	rand.Seed(time.Now().Unix())
}

// ------------------------------------------------------

func newCacheService(storage ports.CacheStorage) *cacheService {
	cs = cacheService{
		storage: storage,
		ring:    ring.NewRing[*discovery.CacheNode](),
	}
	return &cs
}

func (service *cacheService) Ping(node *discovery.CacheNode) error {
	return nil
}

func (service *cacheService) registerNode(node *discovery.CacheNode) error {
	// TODO validate node.host is parsable
	// TODO validate port is open (try to ping the cache server)

	// Check duplicate register
	if service.ring.Contains(node) {
		return errors.New("node has already been registered")
	}

	// With every node added or removed, we have to update the ring!
	// Randomly assign vnumber (virtual-number on the ring, 0-256) based on the node weight
	logger.Infow(
		"Registering a new CacheNode",
		zap.String("name", node.Name), zap.String("host", node.Host), zap.Int32("port", node.Port),
	)
	// Assign a new unique ID to this node
	node.Id = node.Host + strconv.Itoa(int(node.Port)) + "--" + uuid.New().String()

	// Save the node
	if err := service.storage.Save(node); err != nil {
		return err
	}
	realNodes := service.ring.GetUniques()

	// TODO: consider node weight later on
	newSize := ring.Size / (len(realNodes) + 1)
	remains := ring.Size % (len(realNodes) + 1)
	logger.Infow("Number of allocated virtual-nodes per real-node", "newSize", newSize)

	service.ring.Lock()
	defer service.ring.Unlock()
	nodesToNotify := make(map[*discovery.CacheNode][]int)

	if newSize == ring.Size {
		// With the very first node, we have to assign all 256 virtual-nodes to this one!
		for i := 0; i < newSize+remains; i++ {
			service.ring.Add(i, node)
		}
	} else {
		// Otherwise, assign/replace random virtual-node with this new node
		for i := 0; i < newSize+remains; i++ {
			n := rand.Intn(ring.Size)
			oldNode := service.ring.Add(n, node)
			info := oldNode.(*discovery.CacheNode)
			nodesToNotify[info] = append(nodesToNotify[info], n)
		}
	}

	if err := service.notifyRingStatus(nodesToNotify); err != nil {
		logger.Error(err.Error())
	}
	return nil
}

func (service *cacheService) unregisterNode(id string) error {
	service.ring.Lock()
	defer service.ring.Unlock()
	node, err := service.storage.GetById(id)
	if err != nil {
		msg := "node not found on the storage"
		logger.Errorw(msg, "id", id)
		return errors.New(msg)
	}
	removed := service.ring.Remove(node)
	if len(removed) == 0 {
		msg := "node not found on the ring"
		logger.Errorw(msg, "node_id", id)
		return errors.New(msg)
	}

	if err := service.storage.Delete(node); err != nil {
		return err
	}
	uniqueNodes := service.ring.GetUniques()
	if len(uniqueNodes) == 0 { // It means the ring became empty because the deleted node was the last one!
		return nil
	}
	for _, pos := range removed {
		n := rand.Intn(len(uniqueNodes))
		service.ring.Add(pos, uniqueNodes[n])
	}
	return nil
}

func (service *cacheService) getVirtualNodes() []*discovery.CacheNode {
	// hmm, better solution?
	vnodes := service.ring.All()
	logger.Debugf("Number of virtual-nodes: %d", len(vnodes))
	return vnodes
}

//--------------------------------------------

func (service *cacheService) notifyRingStatus(changes map[*discovery.CacheNode][]int) error {
	return nil
}
