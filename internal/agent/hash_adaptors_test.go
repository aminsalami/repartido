package agent

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().Unix())
}

func getRandomString(l int) string {
	word := make([]rune, l)
	for j := 0; j < l; j++ {
		word[j] = letters[rand.Intn(len(letters))]
	}
	return string(word)
}

func TestMd5HashManager(t *testing.T) {
	hm := newMd5HashManager()
	assert.Equal(t, "952d2c56d0485958336747bcdd98590d", hex.EncodeToString(hm.Hash("Hello!")))
}

func TestMd5IntFromHash(t *testing.T) {
	hm := newMd5HashManager()
	hashed := hm.Hash("Hello!")
	assert.Equal(t, 13, hm.IntFromHash(hashed))
}

// Test if the hash mod is always less than 128 for random strings
func TestMd5IntFromHashLessThan128(t *testing.T) {
	hm := newMd5HashManager()
	for i := 1; i < len(letters); i++ {
		word := getRandomString(i)
		hashed := hm.Hash(word)
		assert.Less(t, hm.IntFromHash(hashed), 128)
	}
}
