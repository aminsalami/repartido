package node

// ICache is only working with strings, for now!
type ICache interface {
	Get(key string) (string, error)
	Put(key, value string) error
	Del(key string) error
}
