package serialize

import (
	"encoding/binary"
)

func SerializeUint64(val uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, val)
	return buf
}

func DeSerializeUint64(buf []byte) uint64 {
	return binary.LittleEndian.Uint64(buf)
}
