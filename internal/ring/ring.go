package ring

import (
	"sync"
)

// Size indicates the exact number of virtual-nodes on the ring.
const Size = 256

type Comparable interface {
	comparable
	Hash() string
}

type Ring[K Comparable] struct {
	sync.Mutex
	vnodes map[int]K
}

func NewRing[K Comparable]() *Ring[K] {
	return &Ring[K]{
		Mutex:  sync.Mutex{},
		vnodes: make(map[int]K),
	}
}

func (r *Ring[K]) Get(position int) K {
	if position > Size {
		panic("node position cannot be greater than ring.Size")
	}
	return r.vnodes[position]
}

// GetNextN returns n clockwise nodes starting from position on the ring.
// does not return error if it cannot find n nodes
func (r *Ring[K]) GetNextN(position int, n int) []K {
	startingNode := r.Get(position)

	seen := make(map[string]K)
	seen[startingNode.Hash()] = startingNode
	// locate real nodes after startNode
	for i := 0; i < Size; i++ {
		pos := (i + position) % Size
		_, ok := seen[r.vnodes[pos].Hash()]
		if !ok {
			seen[r.vnodes[pos].Hash()] = r.vnodes[pos]
		}
		if len(seen) >= n {
			break
		}
	}

	var result []K
	for _, node := range seen {
		result = append(result, node)
	}
	return result
}

func (r *Ring[K]) Contains(node K) bool {
	for _, v := range r.vnodes {
		if v.Hash() == node.Hash() {
			return true
		}
	}
	return false
}

// Add a new node to the ring on the specified position.
// If a node with the same position is already exists, replace it then return the old node.
func (r *Ring[K]) Add(position int, node K) any {
	old, ok := r.vnodes[position]
	r.vnodes[position] = node
	if ok {
		return old
	}
	return nil
}

// Remove all positions that are pointing to `node`
func (r *Ring[K]) Remove(node K) []int {
	var removed []int
	for key, v := range r.vnodes {
		if node.Hash() == v.Hash() {
			removed = append(removed, key)
			delete(r.vnodes, key)
		}
	}
	return removed
}

// All returns virtual nodes on the ring sorted by their index (from 0 to 255)
func (r *Ring[K]) All() []K {
	// The ring size is ALWAYS `0` or `256`
	if len(r.vnodes) < Size {
		return []K{}
	}
	var tmp [Size]K
	for k, v := range r.vnodes {
		tmp[k] = v
	}
	return tmp[:]
}

// GetUniques returns all the unique values stored on the ring
func (r *Ring[K]) GetUniques() map[K][]uint32 {
	tmp := make(map[K][]uint32)

	for i, v := range r.vnodes {
		tmp[v] = append(tmp[v], uint32(i))
	}
	return tmp
}

func (r *Ring[K]) GetUniqueHashes() map[string]K {
	// TODO: performance? compare Hash() vs builtin tmp[K] access performance
	tmp := make(map[string]K)
	for _, node := range r.vnodes {
		if _, ok := tmp[node.Hash()]; !ok {
			tmp[node.Hash()] = node
		}
	}
	return tmp
}
