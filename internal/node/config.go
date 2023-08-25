package node

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	intervalDefault = 300
	intervalMax     = 1000
	intervalMin     = 300
)

type NodeConfig struct {
	// A unique name in the cluster/memberlist
	Name string
	// The public hostname or IP of the node. Default is os.Hostname()
	Host string

	// Internal/Command Port
	Port int
	// Coordinator/Public Port used by clients
	CoordinatorPort int

	RamSize int `mapstructure:ram_size`
}

type GossipConfig struct {
	Port     int
	Interval time.Duration
	Peers    []string
}

type Config struct {
	Node   NodeConfig
	Gossip GossipConfig

	// Dev or Production mode
	DevMode bool
}

// Validate check the conf object and returns a modified/validated version of the config object
// TODO: Improve the validation for every available config
func (c *Config) Validate() (error, *Config) {
	if c.Node.Host == "" {
		host, err := os.Hostname()
		if err != nil {
			logger.Fatalw("invalid hostname", "err", err)
		}
		c.Node.Host = host
	}
	if c.Node.Port == 0 {
		c.Node.Port = 8100
	}
	if c.Node.CoordinatorPort == 0 {
		c.Node.CoordinatorPort = c.Node.Port + 100
	}

	if c.Node.RamSize == 0 {
		c.Node.RamSize = 1024
	}

	if c.Gossip.Interval == 0 {
		c.Gossip.Interval = intervalDefault
	} else if c.Gossip.Interval > intervalMax {
		logger.Warnw("value must be less than `1000`ms. it's been set to default max-value", "key", "gossip.Interval", "value", c.Gossip.Interval)
		c.Gossip.Interval = intervalMax
	} else if c.Gossip.Interval < intervalMin {
		logger.Warnw("value must be more than `100`ms. it's been set to default min-value", "key", "gossip.Interval", "value", c.Gossip.Interval)
		c.Gossip.Interval = intervalMin
	}
	if c.Gossip.Port == 0 {
		c.Gossip.Port = 7946
	}

	var newPeers []string
	// add default port to every peer
	for _, peer := range c.Gossip.Peers {
		newPeers = append(newPeers, c.validatePeer(peer))
	}

	return nil, c
}

func (c *Config) validatePeer(peer string) string {
	// TODO: do we need to validate the peer format? such as URLs, IPs, etc
	split := strings.Split(peer, ":")
	if len(split) == 1 {
		return split[0] + ":7946"
	}
	return peer
}

func (c *Config) GetNodeAddr() string {
	return c.Node.Host + ":" + strconv.Itoa(c.Node.Port)
}

func (c *Config) GetCoordinatorAddr() string {
	return c.Node.Host + ":" + strconv.Itoa(c.Node.CoordinatorPort)
}

func (c *Config) GetGossipAddr() string {
	return c.Node.Host + ":" + strconv.Itoa(c.Gossip.Port)
}
