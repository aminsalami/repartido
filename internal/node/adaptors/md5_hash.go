package adaptors

import (
	"crypto/md5"
	"github.com/aminsalami/repartido/internal/node/ports"
	"math/big"
)

type md5Hash struct {
}

func (hm md5Hash) Hash(data string) []byte {
	h := md5.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func (hm md5Hash) IntFromHash(h []byte) int {
	b := big.NewInt(0)
	b.SetBytes(h)

	tmp := big.NewInt(0)
	result := tmp.Mod(b, big.NewInt(128))
	return int(result.Int64())
}

func NewMd5Hash() ports.ConsistentHash {
	return md5Hash{}
}
