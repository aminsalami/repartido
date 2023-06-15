package agent

//
//import (
//	"context"
//	"github.com/aminsalami/repartido/internal/agent/entities"
//	nodepkg "github.com/aminsalami/repartido/internal/node"
//	"github.com/aminsalami/repartido/internal/ring"
//	discoveryGrpc "github.com/aminsalami/repartido/proto/discovery"
//	nodeGrpc "github.com/aminsalami/repartido/proto/node"
//	"github.com/brianvoe/gofakeit/v6"
//	"github.com/stretchr/testify/assert"
//	googleGrpc "google.golang.org/grpc"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/connectivity"
//	"google.golang.org/grpc/credentials/insecure"
//	"google.golang.org/grpc/status"
//	"net"
//	"strconv"
//	"testing"
//)
////
////type discoveryClient struct{}
////
////func (d *discoveryClient) Get(ctx context.Context, in *discoveryGrpc.NodeId, opts ...googleGrpc.CallOption) (*discoveryGrpc.NodeInfo, error) {
////	//TODO implement me
////	panic("implement me")
////}
////
////func (d *discoveryClient) GetRing(ctx context.Context, in *discoveryGrpc.Empty, opts ...googleGrpc.CallOption) (*discoveryGrpc.RingListResponse, error) {
////	//TODO implement me
////	panic("implement me")
////}
////
////func (d *discoveryClient) Register(ctx context.Context, in *discoveryGrpc.NodeInfo, opts ...googleGrpc.CallOption) (*discoveryGrpc.Response, error) {
////	//TODO implement me
////	panic("implement me")
////}
////
////func (d *discoveryClient) Unregister(ctx context.Context, in *discoveryGrpc.NodeId, opts ...googleGrpc.CallOption) (*discoveryGrpc.Response, error) {
////	//TODO implement me
////	panic("implement me")
////}
//
//// CreateNodeServer on localhost and a random port.
//func CreateNodeServer() *NodeInfo {
//	nc := nodepkg.NodeConfig{
//		Name:    "test?",
//		Host:    "localhost",
//		Port:    gofakeit.IntRange(22000, 23000),
//		RamSize: 2048,
//	}
//	ng := nodepkg.GossipConfig{
//		Port:     2057,
//		Interval: 1000,
//		Peers:    nil,
//	}
//	c := nodepkg.Config{
//		InitCluster: false,
//		Node:        nc,
//		Gossip:      ng,
//	}
//	go func() {
//		nodeServer := nodepkg.NewServer(&c)
//		l, e := net.Listen("tcp", c.GetNodeAddr())
//		if e != nil {
//			logger.Fatal(e.Error())
//		}
//		grpcServer := googleGrpc.NewServer()
//		nodeGrpc.RegisterCommandApiServer(grpcServer, nodeServer)
//		e = grpcServer.Serve(l)
//		if e != nil {
//			logger.Fatal(e.Error())
//		}
//	}()
//	nodeInfo := NodeInfo{}
//	err := gofakeit.Struct(nodeInfo)
//	if err != nil {
//		logger.Fatal(err.Error())
//	}
//	conn, err := googleGrpc.Dial(c.GetNodeAddr(), googleGrpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		logger.Fatal(err.Error())
//	}
//	nodeInfo.grpc = nodeGrpc.NewCommandApiClient(conn)
//	return &nodeInfo
//}
//
//// -----------------------------------------------------------------
//// -----------------------------------------------------------------
//
//func TestAgent_SetAndGet(t *testing.T) {
//	node1 := CreateNodeServer()
//	a := NewDefaultAgent()
//	// manually set a discovery client without grpc.Dial() method
//	a.discoveryClient = &discoveryClient{}
//	for i := 0; i < ring.Size; i++ {
//		a.ring.Add(i, node1)
//	}
//
//	r := entities.ParsedRequest{
//		Command: entities.SET,
//		Key:     "amin",
//		Data:    "123",
//	}
//	err := a.StoreData(r)
//	assert.NoError(t, err)
//
//	r.Command = entities.GET
//	response, err := a.RetrieveData(r)
//	assert.NoError(t, err)
//	assert.Equal(t, r.Data, response)
//
//	r.Key = "amin2"
//	response, err = a.RetrieveData(r)
//	assert.Error(t, err)
//}
//
//func TestAgent_DeleteData(t *testing.T) {
//	node1 := CreateNodeServer()
//
//	// step-1: Delete non-existence key
//	a := NewDefaultAgent()
//	a.discoveryClient = &discoveryClient{}
//
//	for i := 0; i < ring.Size; i++ {
//		a.ring.Add(i, node1)
//	}
//
//	r := entities.ParsedRequest{
//		Command: entities.DEL,
//		Key:     "key-1",
//		Data:    "data-1",
//	}
//	err := a.DeleteData(r)
//	assert.Error(t, err)
//}
//
//// Try to test what happens when the grpcServer disconnected while we call a SET method
//func TestAgent_disconnectWhileGettingData(t *testing.T) {
//	addr := "localhost:" + strconv.Itoa(gofakeit.IntRange(22000, 23000))
//	l, e := net.Listen("tcp", addr)
//	if e != nil {
//		logger.Fatal(e.Error())
//	}
//	go func() {
//		nodeServer := nodepkg.NewServer()
//		nodeServer.SetId("SERVER_UNIQUE_ID__" + addr)
//
//		grpcServer := googleGrpc.NewServer()
//		nodeGrpc.RegisterCommandApiServer(grpcServer, nodeServer)
//		_ = grpcServer.Serve(l)
//	}()
//	nodeInfo := &NodeInfo{}
//	err := gofakeit.Struct(nodeInfo)
//	if err != nil {
//		logger.Fatal(err.Error())
//	}
//	conn, err := googleGrpc.Dial(addr, googleGrpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		logger.Fatal(err.Error())
//	}
//	nodeInfo.conn = conn
//	nodeInfo.grpc = nodeGrpc.NewCommandApiClient(conn)
//	// step-1: Delete non-existence key
//	a := NewDefaultAgent()
//	a.discoveryClient = &discoveryClient{}
//	for i := 0; i < ring.Size; i++ {
//		a.ring.Add(i, nodeInfo)
//	}
//
//	// Manually stop the server
//	_ = l.Close()
//	// Try to call a closed server
//	r := entities.ParsedRequest{Command: entities.SET, Key: "amin", Data: "123"}
//	err = a.StoreData(r)
//	assert.Error(t, err)
//	assert.Equal(t, codes.Unavailable, status.Code(err))
//	assert.NotEqual(t, nodeInfo.conn.GetState(), connectivity.Ready)
//}
