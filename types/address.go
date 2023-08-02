package types

import (
	"encoding/hex"
	"fmt"
)

type Address [20]uint8

func (a Address) ToSlice() []byte {
	b := make([]byte, 20)
	for i := 0; i < 20; i++ {
		b[i] = a[i]
	}
	return b
}

func (a Address) String() string {
	return hex.EncodeToString(a.ToSlice())
}

func MustAddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		msg := fmt.Sprintf("given %d bytes should be 20 bytes", len(b))
		panic(msg)
	}

	var value [20]uint8
	for i := 0; i < 20; i++ {
		value[i] = b[i]
	}
	return Address(value)
}

func IsAddressValid(address string) bool {
	if len(address) != 20 {
		msg := fmt.Sprintf("given %d bytes should be 20 bytes", len(address))
		panic(msg)
	}
	return true
}
