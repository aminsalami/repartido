package node

import (
	"context"
	"github.com/aminsalami/repartido/internal/node/adaptors"
	"github.com/aminsalami/repartido/internal/ring"
	nodeProto "github.com/aminsalami/repartido/proto/node"
	"github.com/hashicorp/memberlist"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type nodeService struct {
	conf *Config

	ring *ring.Ring[*Node]

	cluster         *memberlist.Memberlist
	eventBroadcasts *memberlist.TransmitLimitedQueue

	ringUpdateCh chan struct{}
	ringLock     sync.Mutex

	clock   *adaptors.LamportClock
	eventId uint64
}

func newNodeService(conf *Config, ring *ring.Ring[*Node]) *nodeService {
	service := &nodeService{
		conf:         conf,
		ring:         ring,
		ringUpdateCh: make(chan struct{}),
		clock:        adaptors.NewLamportClock(),
		eventId:      1000,
	}

	c := memberlist.DefaultLANConfig()
	c.BindAddr = conf.Node.Host
	c.BindPort = conf.Gossip.Port
	c.Name = conf.Node.Name
	c.GossipInterval = conf.Gossip.Interval * time.Millisecond

	c.Delegate = &gossip{
		service: service,
	}

	ml, err := memberlist.Create(c)
	if err != nil {
		logger.Fatal(err.Error())
	}
	service.cluster = ml

	queue := &memberlist.TransmitLimitedQueue{
		NumNodes: service.numNodes,
		// TODO: tune retransmit number. Add some tests to check the behaviour of RetransmitMult=1
		RetransmitMult: 3,
	}
	service.eventBroadcasts = queue

	return service
}

func (s *nodeService) getId() string {
	return s.conf.Node.Name
}

// initializeRing populates the ring with this node
func (s *nodeService) initializeRing() {
	s.ring.Lock()
	defer s.ring.Unlock()
	me := Node{
		Id:      s.conf.Node.Name,
		Name:    s.conf.Node.Name,
		Host:    s.conf.Node.Host,
		Port:    uint32(s.conf.Node.Port),
		RamSize: uint32(s.conf.Node.RamSize),
		Conn:    nil,
		Client:  nil,
	}
	for i := 0; i < ring.Size; i++ {
		s.ring.Add(i, &me)
	}
	// tick the clock since the ring is created as the first event
	s.clock.Tick()
}

// importRing replace the current ring with the ring-state received from the cluster.
func (s *nodeService) importRing(ringState *nodeProto.RingState) {
	s.ring.Lock()
	defer s.ring.Unlock()

	oldNodes := s.ring.GetUniqueHashes()
	for _, nodeState := range ringState.NodeStates {
		node := &Node{
			Id:      "",
			Name:    nodeState.Name,
			Host:    nodeState.Host,
			Port:    nodeState.Port,
			RamSize: nodeState.RamSize,
			Conn:    nil,
			Client:  nil,
		}
		// Check if the node-state received from remote already exists in the ring
		nodeStateHash := HashFromAddr("", nodeState.Host, nodeState.Port)
		if oldNode, ok := oldNodes[nodeStateHash]; ok {
			node.Conn = oldNode.Conn
			node.Client = oldNode.Client
		} else {
			_ = s.connect(node) // TODO: handle error
		}

		for _, vNum := range nodeState.VNumbers {
			s.ring.Add(int(vNum), node)
		}
	}
	//logger.Infow("ring state", "# of real nodes", len(ringState.NodeStates))
	//for i, server := range ringState.NodeStates {
	//	logger.Debugf("Node-%d -> %s - %s:%d - %d vNodes", i, server.Name, server.Host, server.Port, len(server.VNumbers))
	//}
	//
	logger.Infow("ring imported", "# of real nodes", len(ringState.NodeStates))
	for _, server := range ringState.NodeStates {
		logger.Debugf("%s - %s:%d - %d vNodes", server.Name, server.Host, server.Port, len(server.VNumbers))
	}
}

// getRingState converts ring to a byte-slice ready to be broadcast in cluster
func (s *nodeService) getRingState() *nodeProto.RingState {
	s.ringLock.Lock()
	defer s.ringLock.Unlock()

	var ringState nodeProto.RingState
	uniques := s.ring.GetUniques()

	for vnode, vNumbers := range uniques {
		nodeState := &nodeProto.NodeState{
			Name:     vnode.Name,
			Host:     vnode.Host,
			Port:     uint32(vnode.Port),
			RamSize:  uint32(vnode.RamSize),
			VNumbers: vNumbers,
		}
		ringState.NodeStates = append(ringState.NodeStates, nodeState)
	}

	return &ringState
}

// addNode adds a new node to the ring. Nodes are spread uniformly & randomly.
func (s *nodeService) addNode(node *Node) {
	s.ring.Lock()
	defer s.ring.Unlock()
	uniques := s.ring.GetUniques()
	newSize := ring.Size / (len(uniques) + 1)
	remains := ring.Size % (len(uniques) + 1)
	for i := 0; i < newSize+remains; i++ {
		randomPos := rand.Intn(ring.Size)
		s.ring.Add(randomPos, node)
	}
	logger.Infow("node added to the ring", "node.Name", node.Name, "node.Host", node.Host)
}

// connect tries to create a new grpc connection by dialing the node using Host & Port
func (s *nodeService) connect(node *Node) error {
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Millisecond*1500)
	defer cancelFun()
	conn, err := grpc.DialContext(ctx, node.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}
	node.Client = nodeProto.NewCommandApiClient(conn)
	node.Conn = conn
	logger.Debugw("New GRPC connection to node", node)
	return nil
}

func (s *nodeService) numNodes() int {
	return s.cluster.NumMembers()
}

// -----------------------------------------------------------------

func (s *nodeService) itsMe(node *memberlist.Node) bool {
	return s.conf.Node.Name == node.Name
}

func (s *nodeService) joinRing() {
	logger.Debug("going to join the ring...")
	// injecting me to the ring
	me := Node{
		Id:      "",
		Name:    s.conf.Node.Name,
		Host:    s.conf.Node.Host,
		Port:    uint32(s.conf.Node.Port),
		RamSize: uint32(s.conf.Node.RamSize),
		Conn:    nil,
		Client:  nil,
	}
	s.addNode(&me)

	// Send a new broadcast due to the ring changes
	buff := s.makeRingEvent()
	s.eventBroadcasts.QueueBroadcast(&broadcast{buff})
}

func (s *nodeService) makeRingEvent() []byte {
	ringState := s.getRingState()
	event := s.createEmptyEvent()
	event.Payload = &nodeProto.Event_RingState{RingState: ringState}
	buff, err := proto.Marshal(event)
	if err != nil {
		logger.Error(err)
		return []byte{}
	}
	return buff
}

func (s *nodeService) createEmptyEvent() *nodeProto.Event {
	s.eventId = s.eventId + 1
	e := nodeProto.Event{
		Id:     s.eventId,
		LClock: s.clock.Tick(),
		Origin: s.getId(),
	}
	return &e
}

// -----------------------------------------------------------------

func (s *nodeService) JoinCluster() (error, chan os.Signal) {
	stop := make(chan os.Signal)
	// Handle the termination gracefully. Notify other members.
	go func() {
		signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP)
		<-stop
		logger.Infow("os signal received. Terminating...")
		err := s.cluster.Leave(time.Second * 2)
		if err != nil {
			logger.Error(err)
		}
		os.Exit(1)
	}()

	// Do not proceed if this is the first node creating the cluster
	if s.conf.InitCluster {
		logger.Infow("`InitCluster` flag is found")
		s.initializeRing()
		return nil, stop
	}
	logger.Infow("Trying to join the cluster...")
	// Introduce yourself to the cluster by sending a broadcast join message
	if _, err := s.cluster.Join(s.conf.Gossip.Peers); err != nil {
		return err, stop
	}

	// add me to the ring and let other nodes know about it
	s.joinRing()

	return nil, stop
}

func (s *nodeService) getBroadcastMessages(overhead, limit int) [][]byte {
	return s.eventBroadcasts.GetBroadcasts(overhead, limit)
}

func (s *nodeService) queueBroadcast(b *broadcast) {
	s.eventBroadcasts.QueueBroadcast(b)
}

func (s *nodeService) getNodeAddr() string {
	return s.conf.GetNodeAddr()
}

func (s *nodeService) getGossipAddr() string {
	return s.conf.GetGossipAddr()
}

// witnessClock updates the internal clock and returns true if the internal clock is updated
func (s *nodeService) witnessClock(t uint64) bool {
	return s.clock.Witness(t)
}
