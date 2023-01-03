package agent

import (
	"context"
	"errors"
	"github.com/aminsalami/repartido/internal/agent/adaptors"
	"github.com/aminsalami/repartido/internal/agent/entities"
	"github.com/aminsalami/repartido/internal/agent/ports"
	"github.com/aminsalami/repartido/internal/ring"
	"github.com/aminsalami/repartido/proto/discovery"
	nodeGrpc "github.com/aminsalami/repartido/proto/node"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"sync"
	"time"
)

var logger = zap.NewExample().Sugar()

type NodeInfo struct {
	// Node UUID
	Id string

	Name string
	Addr string

	// A gRPC connection to node
	conn *grpc.ClientConn
	grpc nodeGrpc.CommandApiClient
}

func (n *NodeInfo) Hash() string {
	return n.Id + n.Addr
}

// Agent is standing between clients and nodes. Provides an interface for clients
// to get(or set) data from(or to) the right node.
// Implements ports.IAgent
type Agent struct {
	ring *ring.Ring[*NodeInfo]

	discoveryClient discovery.DiscoveryClient
	// HashManager creates a key(aka hash-string) to be used to find out which connector is holding the data
	HashManager ports.HashManager
	sync.RWMutex
}

// -----------------------------------------------------------------

func NewDefaultAgent() Agent {
	return NewAgent(adaptors.NewMd5HashManager())
}

func NewAgent(hm ports.HashManager) Agent {
	// TODO: Start listening on the port
	return Agent{
		ring:        ring.NewRing[*NodeInfo](),
		HashManager: hm,
	}
}

// -----------------------------------------------------------------

func (agent *Agent) LocateNodeByHash(key []byte) (*NodeInfo, error) {
	// position is a number from 0 to 128. Every number indicates a virtual server.
	position := agent.HashManager.IntFromHash(key)
	// TODO: Check if there is exactly one node from 0 to 127
	return agent.ring.Get(position), nil
}

func (agent *Agent) RetrieveData(request entities.ParsedRequest) (string, error) {
	hash := agent.HashManager.Hash(request.Data)
	nodeInfo, err := agent.LocateNodeByHash(hash)
	if err != nil {
		return "", err
	}

	// Create a gRPC request and send it to node
	c := nodeGrpc.Command{
		Cmd: nodeGrpc.Cmd_GET,
		Key: request.Key,
	}
	res, err := nodeInfo.grpc.Get(context.Background(), &c)

	// TODO: wrap response using errors, this is a 500 error code, provide details
	if err != nil {
		return "", err
	}
	if !res.Success {
		return "", errors.New(res.Data)
	}

	return res.Data, err
}

func (agent *Agent) StoreData(request entities.ParsedRequest) error {
	hash := agent.HashManager.Hash(request.Data)
	nodeInfo, err := agent.LocateNodeByHash(hash)
	if err != nil {
		return err
	}

	c := nodeGrpc.Command{
		Cmd:  nodeGrpc.Cmd_SET,
		Key:  request.Key,
		Data: request.Data,
	}
	response, err := nodeInfo.grpc.Set(context.Background(), &c)
	// TODO: wrap response using errors, this is a 500 error code, provide details
	if err != nil {
		return err
	}
	if !response.Success {
		return errors.New(response.Data)
	}
	return nil
}

func (agent *Agent) DeleteData(request entities.ParsedRequest) error {
	hash := agent.HashManager.Hash(request.Data)
	node, err := agent.LocateNodeByHash(hash)
	if err != nil {
		return err
	}
	c := nodeGrpc.Command{
		Cmd:  nodeGrpc.Cmd_DEL,
		Key:  request.Key,
		Data: request.Data,
	}
	response, err := node.grpc.Del(context.Background(), &c)
	// TODO: wrap response using errors, this is a 500 error code, provide details
	if err != nil {
		return err
	}
	if !response.Success {
		return errors.New(response.Data)
	}
	return nil
}

func (agent *Agent) GetRing(request entities.ParsedRequest) error {
	return nil
}

// -----------------------------------------------------------------

// Handle the pulling mechanism: Every x milliseconds pull the latest info from the controller
//func (agent Agent) pull() {
//}

// Start listening on port to receive commands locally
// The first idea was to implement on top of tcp connector with a simple customized protocol.
// However, It is unnecessary for now. We can rely on http to pass our customized messages.
func (agent *Agent) Start() {
	// Step-1 Start communicating with the "Controller (aka discovery)"
	agent.setupDiscovery()
	agent.updateRing()

	// Try to initialize a grpc connection for every node received from the discovery
	agent.setupConnections()
}

func (agent *Agent) setupDiscovery() {
	dhost := viper.GetString("discovery.host")
	dport := viper.GetString("discovery.port")
	if dhost == "" || dport == "" {
		logger.Fatal("Discovery host & port is required. Check the config file.")
	}
	conn, err := grpc.Dial(dhost+":"+dport, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		logger.Fatal(err.Error())
	}
	cli := discovery.NewDiscoveryClient(conn)
	agent.discoveryClient = cli
	logger.Info("Successfully connected to the discovery server.")
}

func (agent *Agent) updateRing() {
	logger.Info("Updating the ring.")
	resp, err := agent.discoveryClient.GetRing(context.Background(), &discovery.Empty{})
	if err != nil {
		logger.Fatal(err.Error())
	}

	// TODO: Allow working with 1 node in experimental mode (from config or --experimental)
	if len(resp.Nodes) < 2 && !viper.GetBool("development") {
		logger.Fatal("Cannot initialize the agent. Insufficient number of nodes."+
			"At least 2 distributed node is needed to gain a reliable cache system.",
			len(resp.Nodes))
	}

	agent.ring.Lock()
	// Add nodes received from discovery server
	for _, node := range resp.Nodes {
		n := NodeInfo{
			Id:   node.Info.Id,
			Name: node.Info.Name,
			Addr: node.Info.Host + ":" + strconv.Itoa(int(node.Info.Port)),
			grpc: nil,
		}
		for _, v := range node.Vnumbers {
			agent.ring.Add(int(v), &n)
		}
		logger.Debugw("New NodeInfo.", "name", n.Name, "Addr", n.Addr, "virtual node ids", node.Vnumbers)
	}
	agent.ring.Unlock()
	logger.Infow("Successfully initialized the ring.", "# of real nodes", len(resp.Nodes))
}

// setupConnections create a grpc connection for every NodeInfo
func (agent *Agent) setupConnections() {
	logger.Infow("Trying to setup a grpc connection to real nodes...")
	nodes := agent.ring.GetUniques()
	logger.Infow("Trying to setup a grpc connection to real nodes...", "# nodes", len(nodes))
	wg := sync.WaitGroup{}

	for _, node := range nodes {
		go func(n *NodeInfo) {
			wg.Add(1)
			// Connect for the first time
			err := agent.connect(n)
			if err != nil {
				logger.Errorw("Failed to setup grpc connection to node", "node_id", n.Id, "err_msg", err.Error())
				go agent.recoverNode(n)
				wg.Done()
				return
			}
			go agent.watchNode(n)
			logger.Info("Successfully connected to node " + n.Id)
			wg.Done()
		}(node)
	}
	wg.Wait()
}

func (agent *Agent) recoverNode(node *NodeInfo) {
	logger.Infow("Trying to recover node connection", "node_id", node.Id)
	return
}

// connect to a node and set the `node.grpc` value
func (agent *Agent) connect(node *NodeInfo) error {
	logger.Infow("connecting to node", "node_id", node.Id, "addr", node.Addr)
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFun()
	conn, err := grpc.DialContext(ctx, node.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return err
	}
	node.grpc = nodeGrpc.NewCommandApiClient(conn)
	node.conn = conn
	return nil
}

// watchNode watches grpc-connection state
// Tries to reconnect whenever a connection changes the status to anything other than READY
// ? Does the grpc handles reconnection
func (agent *Agent) watchNode(node *NodeInfo) {
	if node.grpc == nil || node.conn.GetState() != connectivity.Ready {
		panic("agent: a connection must be set up already")
	}

	for {
		node.conn.WaitForStateChange(context.Background(), connectivity.Ready)
		// state has changed, Try to reconnect
		err := agent.connect(node)
		if err != nil {
			agent.recoverNode(node)
			logger.Infow("permanently lost connection to node", "node_id", node.Id, "addr", node.Addr)
			break
		}
	}
}
