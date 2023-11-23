package types

import (
	"encoding/hex"
	"fmt"
)

type Address [20]uint8

func AddressFromBytes(b []byte) Address {
	if len(b) != 20 {
		panic(fmt.Sprintf("Given with length %d, should be 20", len(b)))
	}

	var value [20]uint8
	for i := 0; i < 20; i++ {
		value[i] = b[i]
	}
	return Address(value)

}

func (a Address) String() string {
	return hex.EncodeToString(a.Bytes())
}

func (a Address) Bytes() []byte {
	b := make([]byte, 20)
	for i := 0; i < 20; i++ {
		b[i] = a[i]
	}
	return b
}
