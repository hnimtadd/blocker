package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack[int]()
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
