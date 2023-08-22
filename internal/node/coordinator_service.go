package node

import (
	"context"
	"errors"
	"github.com/aminsalami/repartido/internal/node/adaptors"
	"github.com/aminsalami/repartido/internal/node/ports"
	"github.com/aminsalami/repartido/internal/ring"
	proto "github.com/aminsalami/repartido/proto/node"
	"time"
)

var insufficientReadResponses = errors.New("insufficient read responses")
var insufficientWriteResponses = errors.New("insufficient write responses")

type quorum struct {
	n int
	r int
	w int
}

func newQuorum(n, r, w int) quorum {
	// TODO: check if R + W > N
	return quorum{
		n: n,
		r: r,
		w: w,
	}
}

// Coordinator acts as a proxy to the clients to store or receive key/values from the cluster.
// The coordinator handles replication and consistency. In a write-scenario, coordinator locates the right nodes
// for a KEY in the ring, then sends N concurrent write-requests to corresponding nodes and waits to
// receive W successful responses (Quorum Consensus with N replica).
type Coordinator struct {
	nodeService    *nodeService
	quorum         quorum
	ring           *ring.Ring[*Node]
	hasher         ports.ConsistentHash
	storageService *storageService

	conf *Config
}

func NewCoordinator(conf *Config, nodeSrv *nodeService, storageSrv *storageService, ring *ring.Ring[*Node]) *Coordinator {
	q := newQuorum(3, 2, 2) // TODO: get it from the conf
	return &Coordinator{
		nodeService:    nodeSrv,
		quorum:         q,
		ring:           ring,
		hasher:         adaptors.NewMd5Hash(),
		storageService: storageSrv,

		conf: conf,
	}
}

// cGet coordinates a get request to N nodes.
// send concurrent GET requests to nodes and wait for R nodes to response. A successful
// response has R equal values. Otherwise, return error with the details on which nodes failed.
// TODO: create a real response struct to inject meta data (from node, date, ...)
func (c *Coordinator) cGet(cmd *proto.Command) (string, error) {
	nodes := c.locateNodes(cmd.Key)
	responses := make(chan *proto.CommandResponse, len(nodes))
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*1400)
	defer cancelFunc()
	// query corresponding nodes concurrently. Cancel the request if at least R of them have responded.
	for _, node := range nodes {
		go func(node *Node, ctx context.Context) {
			var err error
			var r *proto.CommandResponse
			// Check if current node is among located nodes
			if node.Hash() == c.conf.GetNodeAddr() {
				r, err = c.localRequest(cmd)
			} else {
				r, err = node.Client.Get(ctx, cmd)
			}

			if err != nil {
				if err == context.Canceled {
					return
				}
				responses <- &proto.CommandResponse{Success: false, Data: err.Error() + " -> " + node.Addr()}
			} else {
				responses <- r
			}
		}(node, ctx)
	}

	var successfulResponses []*proto.CommandResponse
	for i := 0; i < len(nodes); i++ {
		res := <-responses
		if !res.Success {
			continue // TODO: what to do with failed responses?
		}
		successfulResponses = append(successfulResponses, res)
		if len(successfulResponses) >= c.quorum.r {
			cancelFunc()
			break
		}
	}

	if len(successfulResponses) < c.quorum.r {
		// we need at least R responses from nodes to consider the GET a successful operation
		return "", insufficientReadResponses
	}
	// TODO: detect inconsistent data, resolve inconsistency, send latest data to the corresponding nodes
	// TODO: Check if vectors are equal or one is an ancestor of the other one
	return successfulResponses[0].Data, nil
}

// cPut coordinates a PUT request.
// A put request is considered successful whe at least receives W ACKs from corresponding nodes.
// Does not support transactions!
// TODO:
func (c *Coordinator) cPut(cmd *proto.Command) error {
	nodes := c.locateNodes(cmd.Key)
	responses := make(chan *proto.CommandResponse, len(nodes))

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*1400)
	defer cancelFunc()
	// Send a PUT request to all corresponding nodes
	for _, node := range nodes {
		go func(node *Node, ctx context.Context) {
			var err error
			var r *proto.CommandResponse
			if node.Hash() == c.conf.GetNodeAddr() {
				r, err = c.localRequest(cmd)
			} else {
				r, err = node.Client.Set(ctx, cmd)
			}
			if err == context.Canceled {
				return
			}
			if err != nil {
				responses <- &proto.CommandResponse{Success: false, Data: err.Error() + " -> " + node.Addr()}
			} else {
				responses <- r
			}
		}(node, ctx)
	}

	var successWrites []*proto.CommandResponse
	for i := 0; i < len(nodes); i++ {
		res := <-responses
		if !res.Success {
			// TODO: handle failed write
			continue
		}
		successWrites = append(successWrites, res)
		if len(successWrites) > c.quorum.w {
			cancelFunc()
			// NOTE: We left responses-channel opened intentionally. It should be garbage collected at the end.
			break
		}
	}
	if len(successWrites) < c.quorum.w {
		//	TODO: handle failed cPut coordination
		return insufficientWriteResponses
	}
	return nil
}

func (c *Coordinator) cDelete(cmd *proto.Command) error {
	nodes := c.locateNodes(cmd.Key)
	responses := make(chan *proto.CommandResponse, len(nodes))

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*1400)
	defer cancelFunc()
	// Send a DEL request to all corresponding nodes
	for _, node := range nodes {
		go func(node *Node) {
			var err error
			var r *proto.CommandResponse
			if node.Hash() == c.conf.GetNodeAddr() {
				r, err = c.localRequest(cmd)
			} else {
				r, err = node.Client.Del(ctx, cmd)
			}
			if err == context.Canceled {
				return
			}
			if err != nil {
				responses <- &proto.CommandResponse{Success: false, Data: err.Error() + " -> " + node.Addr()}
			} else {
				responses <- r
			}
		}(node)
	}

	var successResponses []*proto.CommandResponse
	for i := 0; i < len(nodes); i++ {
		r := <-responses
		if !r.Success {
			// TODO: handle/log a failed request!
			continue
		}
		successResponses = append(successResponses, r)
		if len(successResponses) >= c.quorum.w {
			cancelFunc()
			break
		}
	}

	if len(successResponses) < c.quorum.w {
		return insufficientWriteResponses
	}
	return nil
}

func (c *Coordinator) localRequest(cmd *proto.Command) (*proto.CommandResponse, error) {
	var data string
	var err error
	switch cmd.Cmd {
	case proto.Cmd_GET:
		data, err = c.storageService.getKey(cmd.Key)
	case proto.Cmd_SET:
		err = c.storageService.set(cmd.Key, cmd.Data)
	}
	if err != nil {
		return &proto.CommandResponse{Success: false, Data: err.Error()}, err
	}
	return &proto.CommandResponse{
		Success: true,
		Data:    data,
	}, nil
}

// -----------------------------------------------------------------

// locateNodes returns an ordered list of nodes that potentially stored "key"
func (c *Coordinator) locateNodes(key string) []*Node {
	hashKey := c.hasher.Hash(key)
	pos := c.hasher.IntFromHash(hashKey)
	return c.ring.GetNextN(pos, c.quorum.n)
}
