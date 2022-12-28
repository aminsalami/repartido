package discovery

import (
	"go.uber.org/zap/zapcore"
	"strconv"
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
	//Conn CacheNodeAPI
}

func (n *CacheNode) Hash() string {
	return n.Id + n.Host + ":" + strconv.Itoa(int(n.Port))
}

// MarshalLogObject helps the `zap` to create a structural log of cacheNode object
func (n *CacheNode) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("Name", n.Name)
	enc.AddString("Host", n.Host)
	enc.AddInt32("Port", n.Port)
	enc.AddInt32("RamSize", n.RamSize)
	return nil
}
