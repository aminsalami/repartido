package node

import (
	"context"
	"errors"
	"github.com/aminsalami/repartido/internal/node/adaptors"
	discoveryGrpc "github.com/aminsalami/repartido/proto/discovery"
	nodegrpc "github.com/aminsalami/repartido/proto/node"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

var logger = zap.NewExample().Sugar()

type GrpcServer struct {
	service *cacheService
	nodegrpc.UnimplementedCommandApiServer
}

func (s *GrpcServer) Get(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	res, err := s.service.getKey(command.Key)
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
	if e := s.service.set(command.Key, command.Data); e != nil {
		return &nodegrpc.CommandResponse{}, e
	}
	return &nodegrpc.CommandResponse{Success: true, Data: "Success"}, nil
}

func (s *GrpcServer) Del(ctx context.Context, command *nodegrpc.Command) (*nodegrpc.CommandResponse, error) {
	if e := s.service.delete(command.Key); e != nil {
		return &nodegrpc.CommandResponse{}, e
	}
	return &nodegrpc.CommandResponse{Success: true, Data: ""}, nil
}

func (s *GrpcServer) SetId(id string) {
	s.service.id = id
}

// -----------------------------------------------------------------

func NewServer() *GrpcServer {
	simpleCache := adaptors.NewSimpleCache()
	srv := &cacheService{
		cache: simpleCache,
	}
	return &GrpcServer{
		service: srv,
	}
}

func StartServer() {
	myServer := NewServer()
	host := viper.GetString("node.host")
	port := viper.GetString("node.port")
	if host == "" || port == "" {
		logger.Fatalf("node.host and node.port is required")
	}
	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		logger.Fatal(err.Error())
	}
	// Register on `discovery server`
	id, err := RegisterMe()
	if err != nil {
		logger.Fatal(err.Error())
	}
	myServer.SetId(id)

	logger.Info("Starting cache server...")
	grpcServer := googleGrpc.NewServer()
	nodegrpc.RegisterCommandApiServer(grpcServer, myServer)
	err = grpcServer.Serve(l)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

// RegisterMe tries to register this node on the `Discovery` server
func RegisterMe() (string, error) {
	discoveryAddr := viper.GetString("discovery.ip") + ":" + viper.GetString("discovery.port")
	conn, err := googleGrpc.Dial(discoveryAddr, googleGrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	client := discoveryGrpc.NewDiscoveryClient(conn)

	// TODO: Handle default values, handle errors when the config is not available
	n := discoveryGrpc.NodeInfo{
		Name:    viper.GetString("node.name"),
		Host:    viper.GetString("node.host"),
		Port:    viper.GetInt32("node.port"),
		RamSize: viper.GetInt32("node.ram_size"),
	}
	response, err := client.Register(context.Background(), &n)
	if err != nil {
		return "", err
	}
	if !response.Ok {
		return "", errors.New("Discovery server refused to register. msg:" + response.Message)
	}
	logger.Info("Successfully registered on discovery")
	logger.Info("NODE ID: " + response.Message)
	return response.Message, nil
}
