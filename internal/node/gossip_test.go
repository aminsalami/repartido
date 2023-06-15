package node

import (
	nodeProto "github.com/aminsalami/repartido/proto/node"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

type mockedService struct {
}

func (m *mockedService) syncRing(ringState *nodeProto.RingState) bool {
	return false
}

func TestGossip_NotifyMsg_validRingUpdate(t *testing.T) {
	//srv := nodeService{}
	//g := gossip{
	//	service: &srv,
	//}
	//
	//sampleState := &nodeProto.NodeState{
	//	Name:     "shitty-node",
	//	Host:     "127.0.0.1",
	//	Port:     10,
	//	RamSize:  3000,
	//	VNumbers: []uint32{1, 3, 9, 255},
	//}
	//
	//state := &nodeProto.RingState{
	//	NodeStates: []*nodeProto.NodeState{
	//		sampleState,
	//	},
	//}
	//
	//msg, err := proto.Marshal(&nodeProto.Broadcast{Payload: &nodeProto.Broadcast_RingState{RingState: state}})
	//assert.NoError(t, err)
	//
	//_ = msg
	//g.NotifyMsg(msg)
}

func createService(nodeName string, startNum int, initCluster bool, peers []string) *nodeService {

	conf := &Config{
		InitCluster: initCluster,
		Node: NodeConfig{
			Name:    nodeName,
			Host:    "127.0.0.1",
			Port:    17000 + startNum,
			RamSize: 1023 + startNum,
		},
		Gossip: GossipConfig{
			Port: 17945 + startNum,
			// gossip interval set to 100 millisecond, it is fine for local testing
			Interval: 100,
			Peers:    peers,
		},
	}
	srv := newNodeService(conf)
	return srv
}

func terminateService(stop chan os.Signal) {
	stop <- syscall.SIGKILL
}

// TestNodeService_Join3Nodes tests if ring state is consistent when 2 nodes having the first
// node as their only "peers" in the config
func TestNodeService_Join3Nodes(t *testing.T) {
	srv1 := createService("node-1", 1, true, nil)
	node1Address := srv1.conf.Node.Host + ":" + strconv.Itoa(srv1.conf.Gossip.Port)
	err, stop1 := srv1.JoinCluster()
	defer terminateService(stop1)
	assert.NoError(t, err)

	ring1 := srv1.getRingState()
	// assert when we are creating the cluster, there is exactly 1 node on the ring
	assert.Len(t, ring1.NodeStates, 1)
	assert.Equal(t, srv1.conf.Node.Name, ring1.NodeStates[0].Name)

	srv2 := createService(
		"node-2", 2, false,
		[]string{node1Address},
	)
	err, stop2 := srv2.JoinCluster()
	defer terminateService(stop2)
	assert.NoError(t, err)
	// wait for the broadcast ... (Is there any better solution?!)
	time.Sleep(200 * time.Millisecond)

	// --- Both nodes operating on the same cluster --- //
	// Test if both nodes have the same lamport time
	assert.Equal(t, srv1.clock.Time(), srv2.clock.Time())
	// Test if they have the same ring
	ring1 = srv1.getRingState()
	ring2 := srv2.getRingState()
	assert.Equal(t, len(ring1.NodeStates), len(ring2.NodeStates))

	// --- node-3 joins and sets "node-1" as its only peer --- //
	srv3 := createService(
		"node-3", 3, false,
		[]string{node1Address},
	)
	err, stop3 := srv3.JoinCluster()
	defer terminateService(stop3)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, srv1.clock.Time(), srv2.clock.Time())
	assert.Equal(t, srv1.clock.Time(), srv3.clock.Time())
	assert.Equal(t, srv2.clock.Time(), srv3.clock.Time())

	ring1 = srv1.getRingState()
	ring2 = srv2.getRingState()
	ring3 := srv3.getRingState()
	assert.Equal(t, len(ring1.NodeStates), len(ring2.NodeStates))
	assert.Equal(t, len(ring1.NodeStates), len(ring3.NodeStates))
	assert.Equal(t, len(ring2.NodeStates), len(ring3.NodeStates))
	time.Sleep(200 * time.Millisecond)
}

func TestNodeService_Join5NodesWithComplexPeers(t *testing.T) {
	srv1 := createService("node-1", 1, true, nil)
	srv2 := createService("node-2", 2, false, []string{srv1.getGossipAddr()})
	srv3 := createService("node-3", 3, false, []string{srv2.getGossipAddr()})
	srv4 := createService("node-4", 4, false, []string{srv1.getGossipAddr(), srv2.getGossipAddr()})
	srv5 := createService("node-5", 5, false, []string{srv3.getGossipAddr()})

	err, stop1 := srv1.JoinCluster()
	defer terminateService(stop1)
	assert.NoError(t, err)
	err, stop2 := srv2.JoinCluster()
	defer terminateService(stop2)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	err, stop3 := srv3.JoinCluster()
	defer terminateService(stop3)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	err, stop4 := srv4.JoinCluster()
	defer terminateService(stop4)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	err, stop5 := srv5.JoinCluster()
	defer terminateService(stop5)
	assert.NoError(t, err)
}
