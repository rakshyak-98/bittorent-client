package util

import (
	"crypto/rand"
	"sync"
)

var (
	idOnce sync.Once
	id     [20]byte
)

// GenID returns a stable 20-byte peer ID with the BEP-20 style prefix.
func GenID() [20]byte {
	idOnce.Do(func() {
		copy(id[:], []byte("-ID0001-"))
		if _, err := rand.Read(id[8:]); err != nil {
			panic(err)
		}
	})
	return id
}
