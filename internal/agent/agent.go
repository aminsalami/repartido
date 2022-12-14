package agent

import (
	"context"
	"encoding/hex"
	"github.com/aminsalami/repartido/internal/agent/connector"
	"go.uber.org/zap"
	"net/http"
)

type Command int

const (
	UNKNOWN Command = iota
	GET
	SET
	DEL
	//LISTNODES
)

var logger *zap.Logger

type NodeInfo struct {
	Name string
	Addr string

	// A gRPC connection to node
	Api connector.NodeAPIClient
}

// nodes map virtual nodes into the actual-node info
// There are exactly 128 virtual nodes
var nodes map[int]*NodeInfo

// Agent is standing between clients and nodes. Provides an interface for clients
// to get(or set) data from(or to) the right node.
type Agent struct {
	// Address in the form of "host:port"
	// Default "0.0.0.0:6000"
	Addr string

	// HashManager creates a key(aka hash-string) to be used to find out which connector is holding the data
	HashManager   HashManager
	RequestParser RequestParser
}

// -----------------------------------------------------------------
func init() {
	// TODO: load config file, create logger based on DEBUG flag
	logger, _ = zap.NewDevelopment()
	nodes = make(map[int]*NodeInfo, 128)
}

func NewDefaultAgent() Agent {
	return NewAgent(newMd5HashManager(), "0.0.0.0:6000")
}

func NewAgent(hm HashManager, addr string) Agent {
	// TODO: Start listening on the port
	return Agent{
		Addr:          addr,
		HashManager:   hm,
		RequestParser: NewRestParser(),
	}
}

func (agent Agent) LocateNodeByKey(key []byte) (*NodeInfo, error) {
	// position is a number from 0 to 128. Every number indicates a virtual server.
	position := agent.HashManager.IntFromHash(key)
	// TODO: Check if there is exactly one node from 0 to 127
	return nodes[position], nil
}

func (agent Agent) RetrieveData(data string) (string, error) {
	key := agent.HashManager.Hash(data)
	nodeInfo, err := agent.LocateNodeByKey(key)
	if err != nil {
		return "", err
	}

	// Create a gRPC request and send it to node
	nodeRequest := connector.Request{Key: hex.EncodeToString(key)}
	res, err := nodeInfo.Api.Get(context.Background(), &nodeRequest)

	if err != nil {
		return "", err
	}

	return res.Data, err
}

func (agent Agent) StoreData(data string) error {
	key := agent.HashManager.Hash(data)
	nodeInfo, err := agent.LocateNodeByKey(key)
	if err != nil {
		return err
	}

	nodeRequest := connector.Request{Key: hex.EncodeToString(key)}
	_, err = nodeInfo.Api.Set(context.Background(), &nodeRequest)
	if err != nil {
		return err
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
func (agent Agent) Start() {
	// Step-1 Start communication with the "DisCache Controller"
	go agent.fetchController()

	// Step-2 Start receiving commands from the clients
	logger.Info("Started listening on " + agent.Addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/data", agent.dataHandler)
	mux.HandleFunc("/commands", agent.commandsHandler)

	err := http.ListenAndServe(agent.Addr, mux)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func (agent Agent) dataHandler(rw http.ResponseWriter, req *http.Request) {
	parsedRequest := ParsedRequest{}
	err := agent.RequestParser.Parse(req, &parsedRequest)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// Handle commands based on parsedRequest
	switch parsedRequest.Command {
	case GET:
		res := agent.handleGetCommand(parsedRequest)
		rw.Write([]byte(res))
	case SET:
		res := agent.handleSetCommand(parsedRequest)
		rw.Write([]byte(res))
	default:
		rw.Write([]byte("FUCK YOU EZEKIEL!"))
	}
}

func (agent Agent) commandsHandler(rw http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

// -----------------------------------------------------------------

func (agent Agent) handleGetCommand(request ParsedRequest) string {
	result, err := agent.RetrieveData(request.Key)
	if err != nil {
		logger.Error(err.Error())
		return err.Error()
	}
	//return "GET Key was:" + " -- " + request.Key
	return result
}

func (agent Agent) handleSetCommand(request ParsedRequest) string {
	return "SET Key was:" + " -- " + request.Key
}

func (agent Agent) fetchController() {

}
