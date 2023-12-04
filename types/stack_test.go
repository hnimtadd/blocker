package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntStack(t *testing.T) {
	s := NewStack()
	s.Push(int(1))

	assert.Equal(t, int(1), s.Pop())
}

func TestStringStack(t *testing.T) {
	s := NewStack()
	s.Push("hello")

	assert.Equal(t, "hello", s.Pop())
}

func TestByteStack(t *testing.T) {
	s := NewStack()
	s.Push([]byte("string"))
	assert.Equal(t, 1, s.Len())

	assert.Equal(t, []byte("string"), s.Pop())
}

func TestStack(t *testing.T) {
	s := NewStack()
	assert.Equal(t, 0, s.Len())
	numIns := 1000
	for i := 0; i <= numIns; i++ {
		s.Push(i)
	}
	assert.Equal(t, numIns+1, s.Len())

	for i := numIns; i >= 0; i-- {
		assert.Equal(t, i, s.Pop())
	}
	assert.Equal(t, 0, s.Len())
}
