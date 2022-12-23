package adaptors

import (
	"crypto/md5"
	"github.com/aminsalami/repartido/internal/agent/ports"
	"math/big"
)

type md5HashManager struct {
}

func (hm md5HashManager) Hash(data string) []byte {
	h := md5.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func (hm md5HashManager) IntFromHash(h []byte) int {
	b := big.NewInt(0)
	b.SetBytes(h)

	tmp := big.NewInt(0)
	result := tmp.Mod(b, big.NewInt(128))
	return int(result.Int64())
}

func NewMd5HashManager() ports.HashManager {
	return md5HashManager{}
}
