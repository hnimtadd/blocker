package core

import "errors"

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

var (
	ErrStateExsited    error = errors.New("state exists")
	ErrStateNotExsited error = errors.New("state exists")
)

func (s *State) Put(key string, value []byte) error {
	_, exists := s.data[key]
	if exists {
		return ErrStateExsited
	}

	s.data[key] = value
	return nil
}

func (s *State) Get(key string) ([]byte, error) {
	val, exists := s.data[key]
	if !exists {
		return nil, ErrStateNotExsited
	}
	return val, nil
}

func (s *State) Delete(key string) error {
	delete(s.data, key)
	return nil
}
