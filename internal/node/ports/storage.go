package ports

// Storage is only working with strings, for now!
type Storage interface {
	Get(key string) (string, error)
	Put(key, value string) error
	Del(key string) error
}
