package types

import (
	"crypto/rand"
	"encoding/hex"
)

type Hash [32]byte

func (h Hash) IsZero() bool {
	for i := 0; i < 32; i++ {
		if h[i] != 0 {
			return false
		}
	}
	return true
}

func (h Hash) ToSlice() []byte {
	b := make([]byte, 32)
	for i := 0; i < 32; i++ {
		b[i] = h[i]
	}
	return b
}

func (h Hash) String() string {
	return hex.EncodeToString(h.ToSlice())
}

func HashFromBytes(b []byte) Hash {
	// if len(b) != 32 {
	// 	msg := fmt.Sprintf("given %d bytes should be 32 bytes", len(b))
	// 	panic(msg)
	// }

	var value [32]uint8
	for i := 0; i < 32; i++ {
		value[i] = b[i]
	}

	hash := Hash(value)
	////fmt.Println("Hash(Value) =>", hash)
	return hash
}

func RandomBytes(size int) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func RandomHash() Hash {
	return HashFromBytes(RandomBytes(32))
}
