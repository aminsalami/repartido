package ports

// ConsistentHash provides tools for consistent hashing.
type ConsistentHash interface {
	Hash(string) []byte
	IntFromHash([]byte) int
}
