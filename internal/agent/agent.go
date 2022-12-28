package agent

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/aminsalami/repartido/internal/agent/adaptors"
	"github.com/aminsalami/repartido/internal/agent/entities"
	"github.com/aminsalami/repartido/internal/agent/ports"
	"github.com/aminsalami/repartido/internal/ring"
	"github.com/aminsalami/repartido/proto/discovery"
	connector "github.com/aminsalami/repartido/proto/node"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strconv"
)

var logger = zap.NewExample().Sugar()

type NodeInfo struct {
	// Node UUID
	Id string

	Name string
	Addr string

	// A gRPC connection to node
	grpc connector.CommandApiClient
}

func (n *NodeInfo) Hash() string {
	return n.Id + n.Addr
}

// Agent is standing between clients and nodes. Provides an interface for clients
// to get(or set) data from(or to) the right node.
type Agent struct {
	// Address in the form of "host:port"
	// Default "0.0.0.0:6000"
	Addr string

	ring *ring.Ring[*NodeInfo]

	discoveryClient discovery.DiscoveryApiClient
	// HashManager creates a key(aka hash-string) to be used to find out which connector is holding the data
	HashManager   ports.HashManager
	RequestParser ports.RequestParser
}

// -----------------------------------------------------------------

func NewDefaultAgent() Agent {
	return NewAgent(adaptors.NewMd5HashManager(), "0.0.0.0:6000")
}

func NewAgent(hm ports.HashManager, addr string) Agent {
	// TODO: Start listening on the port
	return Agent{
		Addr: addr,

		ring:          ring.NewRing[*NodeInfo](),
		HashManager:   hm,
		RequestParser: adaptors.NewRestParser(),
	}
}

func (agent *Agent) LocateNodeByKey(key []byte) (*NodeInfo, error) {
	// position is a number from 0 to 128. Every number indicates a virtual server.
	position := agent.HashManager.IntFromHash(key)
	// TODO: Check if there is exactly one node from 0 to 127
	return agent.ring.Get(position), nil
}

func (agent *Agent) RetrieveData(data string) (string, error) {
	key := agent.HashManager.Hash(data)
	nodeInfo, err := agent.LocateNodeByKey(key)
	if err != nil {
		return "", err
	}

	// Create a gRPC request and send it to node
	c := connector.Command{
		Cmd: connector.Cmd_GET,
		Key: hex.EncodeToString(key),
	}
	res, err := nodeInfo.grpc.Get(context.Background(), &c)

	if err != nil {
		return "", err
	}

	return res.Data, err
}

func (agent *Agent) StoreData(data string) error {
	key := agent.HashManager.Hash(data)
	nodeInfo, err := agent.LocateNodeByKey(key)
	if err != nil {
		return err
	}

	c := connector.Command{
		Cmd:  connector.Cmd_SET,
		Key:  hex.EncodeToString(key),
		Data: data,
	}
	r, err := nodeInfo.grpc.Set(context.Background(), &c)
	if err != nil {
		return err
	}
	if !r.Success {
		return fmt.Errorf(r.Data)
	}
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
	// Step-1 Start communication with the "Controller (aka discovery)"
	agent.setupDiscovery()
	agent.updateRing()

	// Step-2 Start receiving commands from the clients
	mux := http.NewServeMux()
	mux.HandleFunc("/data", agent.dataHandler)
	mux.HandleFunc("/commands", agent.commandsHandler)
	logger.Info("Started listening on " + agent.Addr)
	err := http.ListenAndServe(agent.Addr, mux)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func (agent *Agent) dataHandler(rw http.ResponseWriter, req *http.Request) {
	parsedRequest := entities.ParsedRequest{}
	err := agent.RequestParser.Parse(req, &parsedRequest)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// Handle commands based on parsedRequest
	switch parsedRequest.Command {
	case entities.GET:
		res := agent.handleGetCommand(parsedRequest)
		rw.Write([]byte(res))
	case entities.SET:
		res := agent.handleSetCommand(parsedRequest)
		rw.Write([]byte(res))
	default:
		rw.Write([]byte("FUCK YOU EZEKIEL!"))
	}
}

func (agent *Agent) commandsHandler(rw http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

// -----------------------------------------------------------------

func (agent *Agent) handleGetCommand(request entities.ParsedRequest) string {
	result, err := agent.RetrieveData(request.Key)
	if err != nil {
		logger.Error(err.Error())
		return err.Error()
	}
	//return "GET Key was:" + " -- " + request.Key
	return result
}

func (agent *Agent) handleSetCommand(request entities.ParsedRequest) string {
	return "SET Key was:" + " -- " + request.Key
}

func (agent *Agent) setupDiscovery() {
	dhost := viper.GetString("discovery.host")
	dport := viper.GetString("discovery.port")
	if dhost == "" || dport == "" {
		logger.Fatal("Discovery host & port is required. Check the config file.")
	}
	conn, err := grpc.Dial(dhost+":"+dport, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal(err.Error())
	}
	cli := discovery.NewDiscoveryApiClient(conn)
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
	if len(resp.Statuses) < 2 && !viper.GetBool("development") {
		logger.Fatal("Cannot initialize the agent. Insufficient number of nodes."+
			"At least 2 distributed node is needed to gain a reliable cache system.",
			len(resp.Statuses))
	}

	// Add nodes received from discovery server
	for _, node := range resp.Statuses {
		n := NodeInfo{
			Name: node.Name,
			Addr: node.Host + ":" + strconv.Itoa(int(node.Port)),
			grpc: nil,
		}
		agent.ring.Lock()
		for _, v := range node.Vnumbers {
			agent.ring.Add(int(v), &n)
		}
		agent.ring.Unlock()
		logger.Debugw("New NodeInfo.", "name", n.Name, "Addr", n.Addr, "virtual node ids", node.Vnumbers)
	}
	logger.Infow("Successfully initialized the ring.", "# of real nodes", len(resp.Statuses))
}
