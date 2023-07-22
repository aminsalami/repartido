package ring

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeNode struct {
	name string
}

func (f *fakeNode) Hash() string {
	return f.name
}

func TestRing_GetUniques(t *testing.T) {
	ring := NewRing[*fakeNode]()

	u := ring.GetUniques()
	assert.Empty(t, u)

	node := &fakeNode{name: ""}
	ring.Add(1, node)
	ring.Add(2, node)
	u = ring.GetUniques()
	assert.Len(t, u, 1)
	assert.Equal(t, u[node], []uint32{1, 2})
}

func TestRing_GetN(t *testing.T) {
	ring := NewRing[*fakeNode]()

	node := &fakeNode{name: "f1"}
	for i := 0; i < Size; i++ {
		ring.Add(i, node)
	}

	nNodes := ring.GetNextN(0, 2)
	assert.Len(t, nNodes, 1)

	// Add two new node (3 in total)
	node2 := &fakeNode{name: "f2"}
	node3 := &fakeNode{name: "f3"}
	ring.Add(10, node2)
	ring.Add(20, node2)
	ring.Add(100, node3)
	ring.Add(200, node3)

	nNodes = ring.GetNextN(0, 3)
	assert.Len(t, nNodes, 3)
	nNodes = ring.GetNextN(100, 3)
	assert.Len(t, nNodes, 3)
	nNodes = ring.GetNextN(199, 3)
	assert.Len(t, nNodes, 3)
	nNodes = ring.GetNextN(255, 10)
	assert.Len(t, nNodes, 3)
}
