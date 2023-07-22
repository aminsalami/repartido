package node

import (
	nodegrpc "github.com/aminsalami/repartido/proto/node"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"strconv"
)

// Node Implements ring.Comparable in order to be saved on the ring
type Node struct {
	Id      string
	Name    string
	Host    string
	Port    uint32
	RamSize uint32

	Conn   *grpc.ClientConn
	Client nodegrpc.CommandApiClient
}

// Hash generates a unique string from the node data.
func (n *Node) Hash() string {
	return n.Id + n.Addr()
}

// MarshalLogObject helps the `zap` to create a structural log of cacheNode object
func (n *Node) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("Name", n.Name)
	enc.AddString("Host", n.Host)
	enc.AddUint32("Port", n.Port)
	enc.AddUint32("RamSize", n.RamSize)
	return nil
}

func (n *Node) Addr() string {
	return n.Host + ":" + strconv.Itoa(int(n.Port))
}

func HashFromAddr(id, host string, port uint32) string {
	return id + host + ":" + strconv.Itoa(int(port))
}
