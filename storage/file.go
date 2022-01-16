package storage

import (
	"fmt"
	"sync"

	"github.com/harness/ff-golang-server-sdk/logger"

	jsoniter "github.com/json-iterator/go"

	"os"
	"path/filepath"
	"time"
)

// FileStore object is simple JSON file representation
type FileStore struct {
	project       string
	path          string
	mu            sync.Mutex
	data          map[string]interface{}
	lastPersisted time.Time
	logger        logger.Logger
}

// NewFileStore creates a new file store instance
func NewFileStore(project string, path string, logger logger.Logger) *FileStore {
	return &FileStore{
		project: project,
		path:    filepath.Join(path, fmt.Sprintf("harness-ffm-v1-%s.json", project)),
		data:    make(map[string]interface{}),
		logger:  logger,
	}
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Reset data with custom value, if persist is true save it to the store
func (ds *FileStore) Reset(data map[string]interface{}, persist bool) error {
	ds.data = data
	if persist {
		return ds.Persist()
	}
	return nil
}

// Load data from the store
func (ds *FileStore) Load() error {
	file, err := os.Open(ds.path)
	if err != nil {
		return err
	}
	defer func() {
		if file != nil {
			if err := file.Close(); err != nil {
				ds.logger.Errorf("error closing file, err: %v", err)
			}
		}
	}()
	ds.mu.Lock()
	defer ds.mu.Unlock()
	dec := json.NewDecoder(file)
	if err := dec.Decode(&ds.data); err != nil {
		return err
	}
	return nil
}

// Persist data to the store
func (ds *FileStore) Persist() error {
	file, err := os.Create(ds.path)
	if err != nil {
		return err
	}
	defer func() {
		if file != nil {
			if err := file.Close(); err != nil {
				ds.logger.Errorf("error closing file, err: %v", err)
			}
		}
	}()
	enc := json.NewEncoder(file)
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if err := enc.Encode(ds.data); err != nil {
		return err
	}
	ds.lastPersisted = ds.getTime()
	return nil
}

func (ds *FileStore) getTime() time.Time {
	return time.Now()
}

// Get value with the specified key
func (ds *FileStore) Get(key string) (interface{}, bool) {
	ds.mu.Lock()
	val, ok := ds.data[key]
	ds.mu.Unlock()
	return val, ok
}

// List all values
func (ds *FileStore) List() []interface{} {
	var values []interface{}
	ds.mu.Lock()
	for _, val := range ds.data {
		values = append(values, val)
	}
	ds.mu.Unlock()
	return values
}

// Set new key and value
func (ds *FileStore) Set(key string, value interface{}) error {
	ds.mu.Lock()
	ds.data[key] = value
	ds.mu.Unlock()
	return nil
}

// Remove object from data store identified by key parameter
func (ds *FileStore) Remove(key string) error {
	ds.mu.Lock()
	delete(ds.data, key)
	ds.mu.Unlock()
	return nil
}

// PersistedAt returns when it was last recorded
func (ds *FileStore) PersistedAt() time.Time {
	return ds.lastPersisted
}

// SetLogger set logger
func (ds *FileStore) SetLogger(logger logger.Logger) {
	ds.logger = logger
}
