package storage

import (
	"fmt"

	"github.com/drone/ff-golang-server-sdk/logger"

	jsoniter "github.com/json-iterator/go"

	"os"
	"path/filepath"
	"time"
)

// FileStore object is simple JSON file representation
type FileStore struct {
	name          string
	path          string
	data          map[string]interface{}
	lastPersisted time.Time
	logger        logger.Logger
}

// NewFileStore creates a new file store instance
func NewFileStore(name string, path string, logger logger.Logger) *FileStore {
	return &FileStore{
		name:   name,
		path:   filepath.Join(path, fmt.Sprintf("ffm-v1-%s.json", name)),
		data:   make(map[string]interface{}),
		logger: logger,
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
		if err := file.Close(); err != nil {
			ds.logger.Errorf("error closing file, err: %v", err)
		}
	}()
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
		if err := file.Close(); err != nil {
			ds.logger.Errorf("error closing file, err: %v", err)
		}
	}()
	enc := json.NewEncoder(file)
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
func (ds FileStore) Get(key string) (interface{}, bool) {
	val, ok := ds.data[key]
	return val, ok
}

// List all values
func (ds *FileStore) List() []interface{} {
	var values []interface{}
	for _, val := range ds.data {
		values = append(values, val)
	}
	return values
}

// Set new key and value
func (ds *FileStore) Set(key string, value interface{}) error {
	ds.data[key] = value
	return nil
}

// PersistedAt returns when it was last recorded
func (ds *FileStore) PersistedAt() time.Time {
	return ds.lastPersisted
}

// SetLogger set logger
func (ds FileStore) SetLogger(logger logger.Logger) {
	ds.logger = logger
}
