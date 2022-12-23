package node

import (
	"context"
	"errors"
	grpc2 "github.com/aminsalami/repartido/proto/discovery"
	nodegrpc "github.com/aminsalami/repartido/proto/node"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

type GrpcServer struct {
	service *cacheService
	nodegrpc.UnimplementedCommandApiServer
}

func (s *GrpcServer) Get(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	res, err := s.service.getKey(command.Key)
	if err != nil {
		logger.Error(err.Error())
	}
	response := nodegrpc.CommandResponse{
		Success: true,
		Data:    res,
	}
	return &response, nil
}

func (s *GrpcServer) Set(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	//TODO implement me
	return nil, nil
}

func (s *GrpcServer) Del(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	return nil, nil
}

func StartServer() {
	simpleCache := NewSimpleCache()
	srv := &cacheService{
		cache: simpleCache,
	}
	myServer := GrpcServer{
		service: srv,
	}

	// TODO from config
	l, err := net.Listen("tcp", "localhost:8100")
	if err != nil {
		logger.Fatal(err.Error())
	}
	// Register on `discovery server`
	err = RegisterMe()
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("Starting cache server...")
	grpcServer := googleGrpc.NewServer()
	nodegrpc.RegisterCommandApiServer(grpcServer, &myServer)
	err = grpcServer.Serve(l)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

// RegisterMe tries to register this node on the `Discovery` server
func RegisterMe() error {
	conn, err := googleGrpc.Dial("127.0.0.1:7100", googleGrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	client := grpc2.NewDiscoveryApiClient(conn)

	// TODO get the discovery address from config
	n := grpc2.Node{
		Name:    "node-1",
		Host:    "127.0.0.1",
		Port:    8101,
		RamSize: 1024,
	}
	response, err := client.Register(context.Background(), &n)
	if err != nil {
		return err
	}
	if !response.Ok {
		return errors.New("Discovery server refused to register. msg:" + response.Message)
	}
	logger.Info("Successfully registered on discovery.")
	return nil
}
