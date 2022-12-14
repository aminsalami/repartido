package agent

// Connector connects to a node and retrieve the cached data
// Implementation could be based on REST, grpc, or any customized protocol.
type Connector interface {
	Get(*NodeInfo, string) (string, error)
	Put(*NodeInfo, string) error
	Del(*NodeInfo, string) error
}
