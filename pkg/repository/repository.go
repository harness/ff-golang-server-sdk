package repository

import (
	"fmt"

	"github.com/harness/ff-golang-server-sdk/log"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/harness/ff-golang-server-sdk/storage"
)

// Callback provides events when repository data being modified
type Callback interface {
	OnFlagStored(identifier string)
	OnFlagDeleted(identifier string)
	OnSegmentStored(identifier string)
	OnSegmentDeleted(identifier string)
}

// Repository holds cache and optionally offline data
type Repository struct {
	cache    Cache
	storage  storage.Storage
	callback Callback
}

// New repository with only cache capabillity
func New(cache Cache) Repository {
	return Repository{
		cache: cache,
	}
}

// NewWithStorage works with offline storage implementation
func NewWithStorage(cache Cache, storage storage.Storage) Repository {
	return Repository{
		cache:   cache,
		storage: storage,
	}
}

// NewWithStorageAndCallback factory function with cache, offline storage and
// listener on events
func NewWithStorageAndCallback(cache Cache, storage storage.Storage, callback Callback) Repository {
	return Repository{
		cache:    cache,
		storage:  storage,
		callback: callback,
	}
}

func (r Repository) getFlagAndCache(identifier string, cacheable bool) (rest.FeatureConfig, error) {
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
func (r Repository) GetFlag(identifier string) (rest.FeatureConfig, error) {
	return r.getFlagAndCache(identifier, true)
}

func (r Repository) getSegmentAndCache(identifier string, cacheable bool) (rest.Segment, error) {
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
func (r Repository) GetSegment(identifier string) (rest.Segment, error) {
	return r.getSegmentAndCache(identifier, true)
}

// SetFlag places a flag in the repository with the new value
func (r Repository) SetFlag(featureConfig rest.FeatureConfig) {
	if r.isFlagOutdated(featureConfig) {
		return
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
func (r Repository) SetSegment(segment rest.Segment) {
	if r.isSegmentOutdated(segment) {
		return
	}
	segmentKey := formatFlagKey(segment.Identifier)
	if r.storage != nil {
		if err := r.storage.Set(segmentKey, segment); err != nil {
			log.Errorf("error while storing the segment %s into repository", segment.Identifier)
		}
		r.cache.Remove(segmentKey)
	} else {
		r.cache.Set(segmentKey, segment)
	}

	if r.callback != nil {
		r.callback.OnFlagStored(segment.Identifier)
	}
}

// DeleteFlag removes a flag from the repository
func (r Repository) DeleteFlag(identifier string) {
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
func (r Repository) DeleteSegment(identifier string) {
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

func (r Repository) isFlagOutdated(featureConfig rest.FeatureConfig) bool {
	oldFlag, err := r.getFlagAndCache(featureConfig.Feature, false)
	if err != nil || oldFlag.Version == nil {
		return false
	}

	return *oldFlag.Version >= *featureConfig.Version
}

func (r Repository) isSegmentOutdated(segment rest.Segment) bool {
	oldSegment, err := r.getSegmentAndCache(segment.Identifier, false)
	if err != nil || oldSegment.Version == nil {
		return false
	}

	return *oldSegment.Version >= *segment.Version
}

// Close all resources
func (r Repository) Close() {

}

func formatFlagKey(identifier string) string {
	return "flags/" + identifier
}

func formatSegmentKey(identifier string) string {
	return "segments/" + identifier
}
