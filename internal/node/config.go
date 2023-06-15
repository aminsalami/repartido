package node

import (
	"strconv"
	"time"
)

const (
	intervalDefault = 0
	intervalMax     = 1000
	intervalMin     = 100
)

type NodeConfig struct {
	Name    string
	Host    string
	Port    int
	RamSize int `mapstructure:ram_size`
}

type GossipConfig struct {
	Port     int
	Interval time.Duration
	Peers    []string
}

type Config struct {
	InitCluster bool
	Node        NodeConfig
	Gossip      GossipConfig
}

// Validate ...
func (c *Config) Validate() (error, *Config) {
	if c.Gossip.Interval == 0 {
		c.Gossip.Interval = intervalDefault
	} else if c.Gossip.Interval > intervalMax {
		logger.Warnw("value must be less than `1000`ms. it's been set to default max-value", "key", "gossip.Interval", "value", c.Gossip.Interval)
		c.Gossip.Interval = intervalMax
	} else if c.Gossip.Interval < intervalMin {
		logger.Warnw("value must be more than `100`ms. it's been set to default min-value", "key", "gossip.Interval", "value", c.Gossip.Interval)
		c.Gossip.Interval = intervalMin
	}

	if !c.InitCluster && len(c.Gossip.Peers) == 0 {
		logger.Fatalw("invalid config. `gossip.peers` is required to join the cluster")
	}
	return nil, c
}

func (c *Config) GetNodeAddr() string {
	return c.Node.Host + ":" + strconv.Itoa(c.Node.Port)
}

func (c *Config) GetGossipAddr() string {
	return c.Node.Host + ":" + strconv.Itoa(c.Gossip.Port)
}
