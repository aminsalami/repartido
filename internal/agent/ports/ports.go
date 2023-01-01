package ports

import (
	"github.com/aminsalami/repartido/internal/agent/entities"
)

// -----------------------------------------------------------------
// Define API: Application Protocol Interfaces
// -----------------------------------------------------------------

type IAgent interface {
	RetrieveData(entities.ParsedRequest) (string, error)
	StoreData(entities.ParsedRequest) error
	DeleteData(entities.ParsedRequest) error

	GetRing(entities.ParsedRequest) error
}

// -----------------------------------------------------------------
// Define SPI: Service Provider Interfaces
// -----------------------------------------------------------------

// HashManager provides tools for consistent hashing.
type HashManager interface {
	Hash(string) []byte
	IntFromHash([]byte) int
}

// RequestParser is an interface to convert client requests (through REST, grpc, customized-protocol) to ParsedRequest
type RequestParser interface {
	Parse(any, *entities.ParsedRequest) error
}
