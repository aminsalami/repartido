package agent

import (
	"context"
	"errors"
	"github.com/aminsalami/repartido/internal/agent/entities"
	"github.com/aminsalami/repartido/internal/ring"
	discoveryGrpc "github.com/aminsalami/repartido/proto/discovery"
	nodeGrpc "github.com/aminsalami/repartido/proto/node"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"testing"
)

type discoveryClient struct{}

func (d *discoveryClient) Get(ctx context.Context, in *discoveryGrpc.NodeId, opts ...grpc.CallOption) (*discoveryGrpc.NodeInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (d *discoveryClient) GetRing(ctx context.Context, in *discoveryGrpc.Empty, opts ...grpc.CallOption) (*discoveryGrpc.RingListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (d *discoveryClient) Register(ctx context.Context, in *discoveryGrpc.NodeInfo, opts ...grpc.CallOption) (*discoveryGrpc.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (d *discoveryClient) Unregister(ctx context.Context, in *discoveryGrpc.NodeId, opts ...grpc.CallOption) (*discoveryGrpc.Response, error) {
	//TODO implement me
	panic("implement me")
}

type mockedNodeClient struct {
	cache map[string]string
}

func (n *mockedNodeClient) Get(ctx context.Context, in *nodeGrpc.Command, opts ...grpc.CallOption) (*nodeGrpc.CommandResponse, error) {
	d, ok := n.cache[in.Key]
	if !ok {
		return &nodeGrpc.CommandResponse{
			Success: false,
			Data:    "Not Found!",
		}, errors.New("not found")
	}
	return &nodeGrpc.CommandResponse{
		Success: true,
		Data:    d,
	}, nil

}

func (n *mockedNodeClient) Set(ctx context.Context, in *nodeGrpc.Command, opts ...grpc.CallOption) (*nodeGrpc.CommandResponse, error) {
	if in.Key == "" {
		return &nodeGrpc.CommandResponse{}, errors.New("wrong key/data")
	}
	n.cache[in.Key] = in.Data
	return &nodeGrpc.CommandResponse{
		Success: true,
		Data:    "DONE?",
	}, nil
}

func (n *mockedNodeClient) Del(ctx context.Context, in *nodeGrpc.Command, opts ...grpc.CallOption) (*nodeGrpc.CommandResponse, error) {
	//TODO implement me
	panic("implement me")
}

func generateRandomNodeInfo() *NodeInfo {
	return &NodeInfo{
		Id:   "1.1.1.1:8101--" + uuid.New().String(),
		Name: "node-1",
		Addr: "1.1.1.1:8101",
		grpc: nil,
	}
}

// -----------------------------------------------------------------
// -----------------------------------------------------------------

func TestSetAndGet(t *testing.T) {
	a := NewDefaultAgent()
	// manually set a discovery client without grpc.Dial() method
	a.discoveryClient = &discoveryClient{}
	node1 := &NodeInfo{
		Id:   "1.1.1.1:8101--" + uuid.New().String(),
		Name: "node-1",
		Addr: "1.1.1.1:8101",
		grpc: &mockedNodeClient{cache: make(map[string]string)},
	}
	for i := 0; i < ring.Size; i++ {
		a.ring.Add(i, node1)
	}

	r := entities.ParsedRequest{
		Command: entities.SET,
		Key:     "amin",
		Data:    "123",
	}
	err := a.StoreData(r)
	assert.NoError(t, err)

	r.Command = entities.GET
	response, err := a.RetrieveData(r)
	assert.NoError(t, err)
	assert.Equal(t, r.Data, response)

	r.Key = "amin2"
	response, err = a.RetrieveData(r)
	assert.Error(t, err)
	assert.NotEqual(t, r.Data, response)
}
