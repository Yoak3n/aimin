package state

type StateStorage interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
}

type InMemoryStateStorage struct {
	storage map[string][]byte
}

func (s *InMemoryStateStorage) Save(key string, data []byte) error {
	s.storage[key] = data
	return nil
}

func (s *InMemoryStateStorage) Load(key string) ([]byte, error) {
	data, exists := s.storage[key]
	if !exists {
		return nil, nil
	}
	return data, nil
}

func NewInMemoryStateStorage() StateStorage {
	return &InMemoryStateStorage{
		storage: make(map[string][]byte),
	}
}
