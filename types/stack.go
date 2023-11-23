package types

type Stack[T any] struct {
	data []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		data: []T{},
	}
}

func (l *Stack[T]) Push(v T) {
	l.data = append(l.data, v)
}

func (l *Stack[T]) Pop() (v T) {
	if l.Len() == 0 {
		return
	}
	v = l.data[l.Len()-1]
	l.data = l.data[0 : l.Len()-1]
	return v
}

func (l *Stack[T]) Len() int {
	return len(l.data)
}
