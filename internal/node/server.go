package node

import (
	"context"
	"errors"
	"github.com/aminsalami/repartido/internal/node/adaptors"
	"github.com/aminsalami/repartido/internal/ring"
	nodegrpc "github.com/aminsalami/repartido/proto/node"
	"go.uber.org/zap"
	googleGrpc "google.golang.org/grpc"
	"net"
)

var logger = zap.NewExample().Sugar()

type CommandServer struct {
	conf           *Config
	storageService *storageService
	nodegrpc.UnimplementedCommandApiServer
}

func (s *CommandServer) Get(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	res, err := s.storageService.getKey(command.Key)
	if err != nil {
		logger.Error(err.Error())
		return &nodegrpc.CommandResponse{Success: false, Data: err.Error()}, err
	}
	response := nodegrpc.CommandResponse{
		Success: true,
		Data:    res,
	}
	return &response, nil
}

func (s *CommandServer) Set(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	if command.Data == "" || command.Key == "" {
		return &nodegrpc.CommandResponse{Success: false, Data: ""}, errors.New("empty key/vaue is not allowed")
	}
	if e := s.storageService.set(command.Key, command.Data); e != nil {
		return &nodegrpc.CommandResponse{}, e
	}
	return &nodegrpc.CommandResponse{Success: true, Data: "Success"}, nil
}

func (s *CommandServer) Del(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	if e := s.storageService.delete(command.Key); e != nil {
		return &nodegrpc.CommandResponse{}, e
	}
	return &nodegrpc.CommandResponse{Success: true, Data: ""}, nil
}

func (s *CommandServer) Start() {
	l, err := net.Listen("tcp", s.conf.GetNodeAddr())
	if err != nil {
		logger.Fatal(err.Error())
	}

	grpcServer := googleGrpc.NewServer()
	nodegrpc.RegisterCommandApiServer(grpcServer, s)
	logger.Info("-> node started serving request on " + s.conf.GetNodeAddr())
	if err := grpcServer.Serve(l); err != nil {
		logger.Fatal(err.Error())
	}
}

// -----------------------------------------------------------------
// -----------------------------------------------------------------

// CoordinatorServer handles the communication between client and cluster. It receives GET/PUT/... requests
// from clients and coordinates the request to the cluster.
// TODO: implement REST server/controller
type CoordinatorServer struct {
	coordinatorService *Coordinator
	conf               *Config

	nodegrpc.UnimplementedCoordinatorApiServer
}

func (cs *CoordinatorServer) Get(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CoordinatorResponse, error) {
	data, err := cs.coordinatorService.cGet(command)
	if err != nil {
		logger.Error(err.Error())
		return &nodegrpc.CoordinatorResponse{Success: false, Data: err.Error()}, err
	}
	response := nodegrpc.CoordinatorResponse{
		Success: true,
		Data:    data,
	}
	return &response, nil
}

func (cs *CoordinatorServer) Set(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CoordinatorResponse, error) {
	err := cs.coordinatorService.cPut(command)
	if err != nil {
		logger.Error(err.Error())
		return &nodegrpc.CoordinatorResponse{Success: false, Data: err.Error()}, err
	}
	response := nodegrpc.CoordinatorResponse{
		Success: true,
		Data:    "done!",
	}
	return &response, nil
}

func (cs *CoordinatorServer) Del(ctx context.Context, in *nodegrpc.Command) (*nodegrpc.CoordinatorResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (cs *CoordinatorServer) Start() {
	l, err := net.Listen("tcp", cs.conf.GetPublicAddr())
	if err != nil {
		logger.Fatal(err)
	}
	grpcServer := googleGrpc.NewServer()
	nodegrpc.RegisterCoordinatorApiServer(grpcServer, cs)
	logger.Info("-> Coordinator server started on " + cs.conf.GetPublicAddr())
	if err := grpcServer.Serve(l); err != nil {
		logger.Fatal(err)
	}
}

// -----------------------------------------------------------------
// -----------------------------------------------------------------

// Init initializes and starts servers and services including CommandServer, CoordinatorServer, etc.
func Init(conf *Config) {
	r := ring.NewRing[*Node]()
	nodeService := newNodeService(conf, r)
	err, _ := nodeService.JoinCluster()
	if err != nil {
		logger.Fatal(err.Error())
	}

	// New storage service
	storageSrv := &storageService{
		storage: adaptors.NewSimpleCache(),
	}
	// Start the command-server
	commandServer := &CommandServer{
		conf:           conf,
		storageService: storageSrv,
	}
	go commandServer.Start()

	// Create a coordinator service and then start the coordinator-server
	coordinator := NewCoordinator(conf, nodeService, storageSrv, r)
	coordinatorServer := CoordinatorServer{
		coordinatorService: coordinator,
		conf:               conf,
	}
	coordinatorServer.Start() // blocking
}
