package storage

import (
	"sync"

	"github.com/drone/ff-golang-server-sdk/utils"
)

type MemoryStore struct {
	observer sync.Map
	sync.RWMutex
	db map[string]interface{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		db: make(map[string]interface{}),
	}
}

func (m *MemoryStore) Save(key string, value interface{}) error {
	m.RLock()
	defer m.RUnlock()
	m.db[key] = value
	m.Notify(&utils.Event{
		EventType: utils.SAVE,
		Key:       key,
		Value:     value,
	})
	return nil
}

func (m *MemoryStore) Load(key string) (interface{}, error) {
	m.RLock()
	defer m.RUnlock()

	return m.db[key], nil
}

func (m *MemoryStore) AddObserver(observer utils.Observer) {
	m.observer.Store(observer, struct{}{})
}

func (m *MemoryStore) Notify(event *utils.Event) {
	m.observer.Range(func(key, value interface{}) bool {
		if key == nil {
			return false
		}

		key.(utils.Observer).NotifyCallback(event)
		return true
	})
}

func (m *MemoryStore) RemoveObserver(observer interface{}) {
	m.observer.Delete(observer)
}

// List all items
func (m *MemoryStore) List() []interface{} {
	list := make([]interface{}, 0)
	for _, val := range m.db {
		list = append(list, val)
	}
	return list
}
