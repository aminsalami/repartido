package ports

import (
	"github.com/aminsalami/repartido/internal/agent/entities"
)

// HashManager implements tools for consistent hashing.
type HashManager interface {
	Hash(string) []byte
	IntFromHash([]byte) int
}

// RequestParser is an interface to convert client requests (through REST, grpc, customized-protocol) to ParsedRequest
type RequestParser interface {
	Parse(any, *entities.ParsedRequest) error
}
