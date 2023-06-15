package discovery

import (
	nodegrpc "github.com/aminsalami/repartido/proto/node"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"strconv"
)

type State int

const (
	Healthy State = iota
	Down
)

// CacheNode Implements ring.Comparable in order to be saved on the ring
type CacheNode struct {
	Id   string
	Name string // zap:
	Host string
	Port int32

	// Latest ping datetime in time.RFC3339
	LastPing string
	RamSize  int32

	Conn   *grpc.ClientConn
	Client nodegrpc.CommandApiClient

	State State
}

// Hash generates a unique string from the node data.
func (n *CacheNode) Hash() string {
	return n.Id + n.Addr()
}

// MarshalLogObject helps the `zap` to create a structural log of cacheNode object
func (n *CacheNode) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("Name", n.Name)
	enc.AddString("Host", n.Host)
	enc.AddInt32("Port", n.Port)
	enc.AddInt32("RamSize", n.RamSize)
	return nil
}

func (n *CacheNode) Addr() string {
	return n.Host + ":" + strconv.Itoa(int(n.Port))
}
