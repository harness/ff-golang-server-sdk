package storage

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/drone/ff-golang-server-sdk.v0/logger"
	"github.com/mitchellh/go-homedir"
)

// Storage is an interface that can be implemented in order to have control over how
// the repository of feature toggles is persisted.
type Storage interface {

	// Reset is called after the repository has fetched the feature toggles from the server.
	// If persist is true the implementation of this function should call Persist(). The data
	// passed in here should be owned by the implementer of this interface.
	Reset(data map[string]interface{}, persist bool) error

	// Load is called to load the data from persistent storage and hold it in memory for fast
	// querying.
	Load() error

	// Persist is called when the data in the storage implementation should be persisted to disk.
	Persist() error

	// Get returns the data for the specified feature toggle.
	Get(string) (interface{}, bool)

	Set(string, interface{}) error

	// List returns a list of all feature toggles.
	List() []interface{}

	PersistedAt() time.Time

	SetLogger(logger logger.Logger)
}

// GetHarnessDir returns home folder for harness ff server files
func GetHarnessDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Printf("error while getting home dir: %v", err)
		return ""
	}
	harnessDir := path.Join(home, "harness")
	if _, err := os.Stat(harnessDir); os.IsNotExist(err) {
		err := os.Mkdir(harnessDir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	return harnessDir
}
