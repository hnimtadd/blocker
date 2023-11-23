package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	list := NewList[int]()
	assert.NotNil(t, list)
	insLen := 10
	for i := 0; i < insLen; i++ {
		list.Insert(i)
		assert.Equal(t, list.Len(), i+1)
		assert.Equal(t, i, list.Last())
	}

	for i := 0; i < insLen; i++ {
		assert.True(t, list.Contains(i))
		index := list.GetIndex(i)
		li := list.Get(i)
		assert.Equal(t, i, index)
		assert.Equal(t, i, li)
	}
	list.Clear()
	assert.Equal(t, 0, list.Len())

	list.Insert(1)
	assert.True(t, list.Contains(1))
	list.Pop(list.GetIndex(1))
	assert.False(t, list.Contains(1))

	assert.False(t, list.Contains(10))
	list.Insert(1)
	assert.True(t, list.Contains(1))
	list.Remove(1)
	assert.False(t, list.Contains(1))

	assert.Panics(t, func() {
		list.Get(100)
	})

	// get index for no-exists value
	assert.Equal(t, -1, list.GetIndex(999))
}
