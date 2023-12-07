package network

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestServerOptions(t *testing.T) {
	opt := ServerOptions{}
	fmt.Printf("ID: %d, padd: %d\n", unsafe.Sizeof(opt.ID), unsafe.Offsetof(opt.ID))

	fmt.Printf("Logger: %d, padd: %d\n", unsafe.Sizeof(opt.Logger), unsafe.Offsetof(opt.Logger))

	fmt.Printf("Priv: %d, padd: %d\n", unsafe.Sizeof(opt.PrivKey), unsafe.Offsetof(opt.PrivKey))

	fmt.Printf("BlockTime: %d, padd: %d\n", unsafe.Sizeof(opt.blockTime), unsafe.Offsetof(opt.blockTime))

	fmt.Printf("MaxBoolLen: %d, padd: %d\n", unsafe.Sizeof(opt.MaxPoolLen), unsafe.Offsetof(opt.MaxPoolLen))

	fmt.Printf("Version: %d, padd: %d\n", unsafe.Sizeof(opt.Version), unsafe.Offsetof(opt.Version))

	fmt.Printf("Localseed: %d, padding: %d\n", unsafe.Sizeof(opt.LocalSeed), unsafe.Offsetof(opt.LocalSeed))
	fmt.Printf("Process: %d, padding: %d\n", unsafe.Sizeof(opt.RPCProcessor), unsafe.Offsetof(opt.RPCProcessor))
	fmt.Printf("decode: %d, padding: %d\n", unsafe.Sizeof(opt.RPCDecodeFunc), unsafe.Offsetof(opt.RPCDecodeFunc))

	fmt.Printf("Full: %d\n", unsafe.Sizeof(opt))
	assert.NotNil(t, opt)
}
