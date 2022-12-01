package repository

import (
	"fmt"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/harness/ff-golang-server-sdk/storage"
)

// Repository interface for data providers
type Repository interface {
	GetFlag(identifier string) (rest.FeatureConfig, error)
	GetSegment(identifier string) (rest.Segment, error)
	GetFlags() ([]rest.FeatureConfig, error)

	SetFlag(featureConfig rest.FeatureConfig, initialLoad bool)
	SetSegment(segment rest.Segment, initialLoad bool)

	DeleteFlag(identifier string)
	DeleteSegment(identifier string)

	Close()
}

// Callback provides events when repository data being modified
type Callback interface {
	OnFlagStored(identifier string)
	OnFlagDeleted(identifier string)
	OnSegmentStored(identifier string)
	OnSegmentDeleted(identifier string)
}

// FFRepository holds cache and optionally offline data
type FFRepository struct {
	cache    Cache
	storage  storage.Storage
	callback Callback
}

// New repository with only cache capabillity
func New(cache Cache) Repository {
	return FFRepository{
		cache: cache,
	}
}

// NewWithStorage works with offline storage implementation
func NewWithStorage(cache Cache, storage storage.Storage) Repository {
	return FFRepository{
		cache:   cache,
		storage: storage,
	}
}

// NewWithStorageAndCallback factory function with cache, offline storage and
// listener on events
func NewWithStorageAndCallback(cache Cache, storage storage.Storage, callback Callback) Repository {
	return FFRepository{
		cache:    cache,
		storage:  storage,
		callback: callback,
	}
}

func (r FFRepository) getFlagAndCache(identifier string, cacheable bool) (rest.FeatureConfig, error) {
	flagKey := formatFlagKey(identifier)
	flag, ok := r.cache.Get(flagKey)
	if ok {
		return flag.(rest.FeatureConfig), nil
	}

	if r.storage != nil {
		flag, ok := r.storage.Get(flagKey)
		if ok && cacheable {
			r.cache.Set(flagKey, flag)
		}
		return flag.(rest.FeatureConfig), nil
	}
	return rest.FeatureConfig{}, fmt.Errorf("%w with identifier: %s", ErrFeatureConfigNotFound, identifier)
}

// GetFlag returns flag from cache or offline storage
func (r FFRepository) GetFlag(identifier string) (rest.FeatureConfig, error) {
	return r.getFlagAndCache(identifier, true)
}

// GetFlags returns all the flags /* Not implemented */
func (r FFRepository) GetFlags() ([]rest.FeatureConfig, error) {
	return []rest.FeatureConfig{}, nil
}

func (r FFRepository) getSegmentAndCache(identifier string, cacheable bool) (rest.Segment, error) {
	segmentKey := formatSegmentKey(identifier)
	flag, ok := r.cache.Get(segmentKey)
	if ok {
		return flag.(rest.Segment), nil
	}

	if r.storage != nil {
		flag, ok := r.storage.Get(segmentKey)
		if ok && cacheable {
			r.cache.Set(segmentKey, flag)
		}
		return flag.(rest.Segment), nil
	}
	return rest.Segment{}, fmt.Errorf("%w with identifier: %s", ErrSegmentNotFound, identifier)
}

// GetSegment returns flag from cache or offline storage
func (r FFRepository) GetSegment(identifier string) (rest.Segment, error) {
	return r.getSegmentAndCache(identifier, true)
}

// SetFlag places a flag in the repository with the new value
func (r FFRepository) SetFlag(featureConfig rest.FeatureConfig, initialLoad bool) {
	if !initialLoad {
		if r.isFlagOutdated(featureConfig) {
			return
		}
	}
	flagKey := formatFlagKey(featureConfig.Feature)
	if r.storage != nil {
		if err := r.storage.Set(flagKey, featureConfig); err != nil {
			log.Errorf("error while storing the flag %s into repository", featureConfig.Feature)
		}
		r.cache.Remove(flagKey)
	} else {
		r.cache.Set(flagKey, featureConfig)
	}

	if r.callback != nil {
		r.callback.OnFlagStored(featureConfig.Feature)
	}
}

// SetSegment places a segment in the repository with the new value
func (r FFRepository) SetSegment(segment rest.Segment, initialLoad bool) {
	if !initialLoad {
		if r.isSegmentOutdated(segment) {
			return
		}
	}
	segmentKey := formatSegmentKey(segment.Identifier)
	if r.storage != nil {
		if err := r.storage.Set(segmentKey, segment); err != nil {
			log.Errorf("error while storing the segment %s into repository", segment.Identifier)
		}
		r.cache.Remove(segmentKey)
	} else {
		r.cache.Set(segmentKey, segment)
	}

	if r.callback != nil {
		r.callback.OnSegmentStored(segment.Identifier)
	}
}

// DeleteFlag removes a flag from the repository
func (r FFRepository) DeleteFlag(identifier string) {
	flagKey := formatFlagKey(identifier)
	if r.storage != nil {
		// remove from storage
		if err := r.storage.Remove(flagKey); err != nil {
			log.Errorf("error while removing flag %s from repository", identifier)
		}
	}
	// remove from cache
	r.cache.Remove(flagKey)
	if r.callback != nil {
		r.callback.OnFlagDeleted(identifier)
	}
}

// DeleteSegment removes a segment from the repository
func (r FFRepository) DeleteSegment(identifier string) {
	segmentKey := formatSegmentKey(identifier)
	if r.storage != nil {
		// remove from storage
		if err := r.storage.Remove(segmentKey); err != nil {
			log.Errorf("error while removing segment %s from repository", identifier)
		}
	}
	// remove from cache
	r.cache.Remove(segmentKey)
	if r.callback != nil {
		r.callback.OnSegmentDeleted(identifier)
	}
}

func (r FFRepository) isFlagOutdated(featureConfig rest.FeatureConfig) bool {
	oldFlag, err := r.getFlagAndCache(featureConfig.Feature, false)
	if err != nil || oldFlag.Version == nil {
		return false
	}

	return *oldFlag.Version >= *featureConfig.Version
}

func (r FFRepository) isSegmentOutdated(segment rest.Segment) bool {
	oldSegment, err := r.getSegmentAndCache(segment.Identifier, false)
	if err != nil || oldSegment.Version == nil {
		return false
	}

	return *oldSegment.Version >= *segment.Version
}

// Close all resources
func (r FFRepository) Close() {

}

func formatFlagKey(identifier string) string {
	return "flag/" + identifier
}

func formatSegmentKey(identifier string) string {
	return "target-segment/" + identifier
}
