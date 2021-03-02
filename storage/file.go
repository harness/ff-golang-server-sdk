package storage

import (
	"fmt"
	"github.com/drone/ff-golang-server-sdk/logger"
	jsoniter "github.com/json-iterator/go"
	"os"
	"path/filepath"
	"time"
)

type FileStore struct {
	project       string
	path          string
	data          map[string]interface{}
	lastPersisted time.Time
	logger        logger.Logger
}

func NewFileStore(project string, path string, logger logger.Logger) *FileStore {
	return &FileStore{
		project: project,
		path:    filepath.Join(path, fmt.Sprintf("harness-ffm-v1-%s.json", project)),
		data:    make(map[string]interface{}),
		logger:  logger,
	}
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (ds *FileStore) Reset(data map[string]interface{}, persist bool) error {
	ds.data = data
	if persist {
		return ds.Persist()
	}
	return nil
}

func (ds *FileStore) Load() error {
	if file, err := os.Open(ds.path); err != nil {
		return err
	} else {
		dec := json.NewDecoder(file)
		if err := dec.Decode(&ds.data); err != nil {
			return err
		}
	}
	return nil
}

func (ds *FileStore) Persist() error {
	if file, err := os.Create(ds.path); err != nil {
		return err
	} else {
		defer file.Close()
		enc := json.NewEncoder(file)
		if err := enc.Encode(ds.data); err != nil {
			return err
		}
		ds.lastPersisted = ds.getTime()
	}
	return nil
}

func (ds *FileStore) getTime() time.Time {
	return time.Now()
}

func (ds FileStore) Get(key string) (interface{}, bool) {
	val, ok := ds.data[key]
	return val, ok
}

func (ds *FileStore) List() []interface{} {
	var features []interface{}
	for _, val := range ds.data {
		features = append(features, val)
	}
	return features
}

func (ds *FileStore) Set(key string, value interface{}) error {
	ds.data[key] = value
	return nil
}

func (ds *FileStore) PersistedAt() time.Time {
	return ds.lastPersisted
}

func (ds FileStore) SetLogger(logger logger.Logger) {
	ds.logger = logger
}
