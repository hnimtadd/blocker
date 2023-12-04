package core

import (
	"blocker/serialize"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPack(t *testing.T) {
	state := NewState()
	data := []byte{0x44, 0x0b, 0x61, 0x0b, 0x74, 0x0b, 0x03, 0x0a, 0x0d}
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	bb := vm.stack.Pop().([]byte)

	fmt.Println(string(bb))
}

func TestPackWithState(t *testing.T) {
	state := NewState()
	data := []byte{
		0x44, 0x0b, // t
		0x61, 0x0b, // a
		0x74, 0x0b, // d
		0x03, 0x0a, 0x0d, // pack => tad
		0x01, 0x0a, // 1
		0x02, 0x0a, // 2
		0x0c,       // add => 3
		0x0f,       // store
		0x44, 0x0b, // t
		0x61, 0x0b, // a
		0x74, 0x0b, // d
		0x03, 0x0a, 0x0d, // pack => tad
		0x0e,
	}
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())

	buf, err := state.Get("taD")
	assert.Nil(t, err)
	val := serialize.DeSerializeUint64(buf)
	assert.Equal(t, uint64(3), val)

	fmt.Println(vm.stack)
}

func TestVM(t *testing.T) {
	state := NewState()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0c}
	vm := NewVM(data, state)
	assert.Nil(t, vm.Run())
	assert.Equal(t, 3, vm.stack.Pop())
}
