package node

import (
	"context"
	"errors"
	"github.com/aminsalami/repartido/internal/node/adaptors"
	nodegrpc "github.com/aminsalami/repartido/proto/node"
	"go.uber.org/zap"
	googleGrpc "google.golang.org/grpc"
	"net"
)

var logger = zap.NewExample().Sugar()

type GrpcServer struct {
	conf           *Config
	storageService *storageService
	nodegrpc.UnimplementedCommandApiServer
}

func (s *GrpcServer) Get(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
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

func (s *GrpcServer) Set(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	if command.Data == "" || command.Key == "" {
		return &nodegrpc.CommandResponse{Success: false, Data: ""}, errors.New("empty key/vaue is not allowed")
	}
	if e := s.storageService.set(command.Key, command.Data); e != nil {
		return &nodegrpc.CommandResponse{}, e
	}
	return &nodegrpc.CommandResponse{Success: true, Data: "Success"}, nil
}

func (s *GrpcServer) Del(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	if e := s.storageService.delete(command.Key); e != nil {
		return &nodegrpc.CommandResponse{}, e
	}
	return &nodegrpc.CommandResponse{Success: true, Data: ""}, nil
}

// -----------------------------------------------------------------

func NewServer(conf *Config) *GrpcServer {
	simpleCache := adaptors.NewSimpleCache()
	srv := &storageService{simpleCache}
	return &GrpcServer{
		conf:           conf,
		storageService: srv,
	}
}

func StartServer(conf *Config) {
	myServer := NewServer(conf)

	l, err := net.Listen("tcp", conf.GetNodeAddr())
	if err != nil {
		logger.Fatal(err.Error())
	}

	grpcServer := googleGrpc.NewServer()
	nodegrpc.RegisterCommandApiServer(grpcServer, myServer)
	logger.Info("Node started serving request on " + conf.GetNodeAddr())
	err = grpcServer.Serve(l)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
