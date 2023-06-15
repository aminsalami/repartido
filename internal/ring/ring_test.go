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
