package state

type StateStorage interface {
	Save(key string, data []byte) error
	Load(key string) (*Context, error)
}

type InMemoryStateStorage struct {
	storage map[string][]byte
}

func (s *InMemoryStateStorage) Save(key string, data []byte) error {
	s.storage[key] = data
	return nil
}

func (s *InMemoryStateStorage) Load(key string) (*Context, error) {
	data, exists := s.storage[key]
	if !exists {
		return nil, nil
	}
	// 反序列化逻辑根据具体实现而定，这里仅作示例
	_ = data
	return &Context{}, nil
}

func NewInMemoryStateStorage() StateStorage {
	return &InMemoryStateStorage{
		storage: make(map[string][]byte),
	}
}
