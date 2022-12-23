package ring

import "sync"

// Size indicates the exact number of virtual-nodes on the ring.
const Size = 256

type Ring struct {
	sync.Mutex
	vnodes map[int]any
}

func NewRing() *Ring {
	return &Ring{
		Mutex:  sync.Mutex{},
		vnodes: make(map[int]any),
	}
}

func (r *Ring) Get(position int) any {
	return r.vnodes[position]
}

// Add a new node to the ring on the specified position.
// If a node with the same position is already exists, replace it then return the old node.
func (r *Ring) Add(position int, a any) any {
	old, ok := r.vnodes[position]
	r.vnodes[position] = a
	if ok {
		return old
	}
	return nil
}

// All returns virtual nodes on the ring sorted by their index (from 0 to 255)
func (r *Ring) All() [Size]any {
	var tmp [Size]any
	for k, v := range r.vnodes {
		tmp[k] = v
	}
	return tmp
}

// GetUniques returns all the unique values stored on the ring
func (r *Ring) GetUniques() []any {
	tmp := make(map[any]struct{})
	for _, v := range r.vnodes {
		if _, ok := tmp[v]; !ok {
			tmp[v] = struct{}{}
		}
	}
	var res []any
	for _, v := range tmp {
		res = append(res, v)
	}
	return res
}
