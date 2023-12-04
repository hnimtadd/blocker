package types

import "fmt"

type Stack struct {
	data []any
}

func NewStack() *Stack {
	return &Stack{
		data: []any{},
	}
}

func (l *Stack) Push(v any) {
	l.data = append(l.data, v)
}

func (l *Stack) Pop() (v any) {
	if l.Len() == 0 {
		return
	}
	v = l.data[l.Len()-1]
	l.data = l.data[0 : l.Len()-1]
	return v
}

func (l *Stack) Len() int {
	return len(l.data)
}

func (l Stack) String() string {
	return fmt.Sprintf("stack:\n=>data: %+v\n=>len: %d\n", l.data, l.Len())
}
