package core

type Storage interface {
	Put(*Block) error
}

type InMemoryStorage struct{}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{}
}

func (s *InMemoryStorage) Put(b *Block) error {
	return nil
}
