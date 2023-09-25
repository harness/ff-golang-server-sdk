package repository

import (
	"fmt"

	"golang.org/x/exp/slices"

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
	SetFlags(initialLoad bool, envID string, featureConfig ...rest.FeatureConfig)
	SetSegment(segment rest.Segment, initialLoad bool)
	SetSegments(initialLoad bool, envID string, segment ...rest.Segment)

	DeleteFlag(identifier string)
	DeleteFlags(envID string, identifier string)
	DeleteSegment(identifier string)
	DeleteSegments(envID string, identifier string)

	Close()
}

// Callback provides events when repository data being modified
type Callback interface {
	OnFlagStored(identifier string)
	OnFlagsStored(envID string)
	OnFlagsDeleted(envID string, identifier string)
	OnFlagDeleted(identifier string)
	OnSegmentStored(identifier string)
	OnSegmentsStored(envID string)
	OnSegmentDeleted(identifier string)
	OnSegmentsDeleted(envID string, identifier string)
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

func (r FFRepository) getFlags(envID string) ([]rest.FeatureConfig, error) {
	flagsKey := formatFlagsKey(envID)
	flags, ok := r.cache.Get(flagsKey)
	if ok {
		return flags.([]rest.FeatureConfig), nil
	}

	return []rest.FeatureConfig{}, fmt.Errorf("%w with environment: %s", ErrFeatureConfigNotFound, envID)
}

func (r FFRepository) getSegments(envID string) ([]rest.Segment, error) {
	segmentsKey := formatSegmentsKey(envID)
	flags, ok := r.cache.Get(segmentsKey)
	if ok {
		return flags.([]rest.Segment), nil
	}

	return []rest.Segment{}, fmt.Errorf("%w with environment: %s", ErrFeatureConfigNotFound, envID)
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
		// If the flag is up to date then we don't need to bother updating the cache
		if !r.isFlagOutdated(featureConfig) {
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

// SetFlags places all the flags in the repository
func (r FFRepository) SetFlags(initialLoad bool, envID string, featureConfigs ...rest.FeatureConfig) {
	if !initialLoad {
		// If the flags are all up to date then we don't need to bother updating the cache and can exit
		if !r.areFlagsOutdated(envID, featureConfigs...) {
			return
		}
	}

	key := formatFlagsKey(envID)

	if r.storage != nil {
		if err := r.storage.Set(key, featureConfigs); err != nil {
			log.Errorf("error while storing flags for env=%s into repository", envID)
		}
		r.cache.Remove(key)
	} else {
		r.cache.Set(key, featureConfigs)
	}

	if r.callback != nil {
		r.callback.OnFlagsStored(envID)
	}
}

// SetSegment places a segment in the repository with the new value
func (r FFRepository) SetSegment(segment rest.Segment, initialLoad bool) {
	if !initialLoad {
		// If the segment isn't outdated then we can exit as we don't need to refresh the cache
		if !r.isSegmentOutdated(segment) {
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

// SetSegments places all the segments in the repository
func (r FFRepository) SetSegments(initialLoad bool, envID string, segments ...rest.Segment) {
	if !initialLoad {
		// If segments aren't outdated then we can exit as we don't need to refresh the cache
		if !r.areSegmentsOutdated(envID, segments...) {
			return
		}
	}

	key := formatSegmentsKey(envID)

	if r.storage != nil {
		if err := r.storage.Set(key, segments); err != nil {
			log.Errorf("error while storing flags for env=%s into repository", envID)
		}
		r.cache.Remove(key)
	} else {
		r.cache.Set(key, segments)
	}

	if r.callback != nil {
		r.callback.OnSegmentsStored(envID)
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

// DeleteFlags removes a flag from the flags key.
//
// We can't just delete the key here the way we can for a single flag because then we'd be removing flags that
// haven't been deleted. So we have to first fetch value, then remove the specific flag that has been deleted
// and update the key in the cache/storage
func (r FFRepository) DeleteFlags(envID string, identifier string) {
	flagsKey := formatFlagsKey(envID)
	if r.storage != nil {
		// remove from storage
		if err := r.storage.Remove(flagsKey); err != nil {
			log.Errorf("error while removing flags %s from repository", envID)
		}
	}

	value, ok := r.cache.Get(flagsKey)
	if !ok {
		log.Errorf("error fetching flags from cache for env=%s", envID)
		return
	}

	featureConfigs, ok := value.([]rest.FeatureConfig)
	if !ok {
		log.Errorf("failed to delete flags, expected type to be []rest.FeatureConfig but got %T", featureConfigs)
		return
	}

	updatedFeatureConfigs := slices.DeleteFunc(featureConfigs, func(element rest.FeatureConfig) bool {
		return element.Feature == identifier
	})
	r.cache.Set(flagsKey, updatedFeatureConfigs)

	if r.callback != nil {
		r.callback.OnFlagsDeleted(envID, identifier)
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

// DeleteSegments removes a Segment from the segments key.
//
// We can't just delete the key here the way we can for a single flag because then we'd be removing segments that
// haven't been deleted. So we have to first fetch value, then remove the specific segment that has been deleted
// and update the key in the cache/storage
func (r FFRepository) DeleteSegments(envID string, identifier string) {
	segmentsKey := formatSegmentsKey(envID)
	if r.storage != nil {
		// remove from storage
		if err := r.storage.Remove(segmentsKey); err != nil {
			log.Errorf("error while removing segments %s from repository", envID)
		}
	}

	value, ok := r.cache.Get(segmentsKey)
	if !ok {
		log.Errorf("error fetching segments from cache for env=%s", envID)
		return
	}

	segments, ok := value.([]rest.Segment)
	if !ok {
		log.Errorf("failed to delete flags, expected type to be []rest.Segment but got %T", segments)
		return
	}

	updatedSegments := slices.DeleteFunc(segments, func(element rest.Segment) bool {
		return element.Identifier == identifier
	})
	r.cache.Set(segmentsKey, updatedSegments)

	if r.callback != nil {
		r.callback.OnSegmentsDeleted(envID, identifier)
	}
}

func (r FFRepository) isFlagOutdated(featureConfig rest.FeatureConfig) bool {
	oldFlag, err := r.getFlagAndCache(featureConfig.Feature, false)
	if err != nil || oldFlag.Version == nil {
		// If we get an error here return true to force a cache update
		return true
	}

	return *oldFlag.Version < *featureConfig.Version
}

func (r FFRepository) getFlagsAndCache(envID string, cacheable bool) ([]rest.FeatureConfig, error) {
	flagKey := formatFlagsKey(envID)
	flag, ok := r.cache.Get(flagKey)
	if ok {
		return flag.([]rest.FeatureConfig), nil
	}

	if r.storage != nil {
		flag, ok := r.storage.Get(flagKey)
		if ok && cacheable {
			r.cache.Set(flagKey, flag)
			return flag.([]rest.FeatureConfig), nil
		}
	}
	return []rest.FeatureConfig{}, fmt.Errorf("%w with identifier: %s", ErrFeatureConfigNotFound, envID)
}

func (r FFRepository) areFlagsOutdated(envID string, flags ...rest.FeatureConfig) bool {

	oldFlags, err := r.getFlags(envID)
	if err != nil {
		// If we get an error return true to force a cache refresh
		return true
	}

	oldFlagMap := map[string]rest.FeatureConfig{}
	for _, v := range oldFlags {
		oldFlagMap[v.Feature] = v
	}

	for _, flag := range flags {
		of, ok := oldFlagMap[flag.Feature]
		if !ok {
			// If a new flag isn't in the oldFlagMap then the list of old flags are outdated and we'll
			// want to refresh the cache
			return true
		}

		if *of.Version < *flag.Version {
			return true
		}
	}
	return false
}

func (r FFRepository) isSegmentOutdated(segment rest.Segment) bool {
	oldSegment, err := r.getSegmentAndCache(segment.Identifier, false)
	if err != nil || oldSegment.Version == nil {
		// If we get an error here return true to force a cache update
		return true
	}

	return *oldSegment.Version < *segment.Version
}

func (r FFRepository) areSegmentsOutdated(envID string, segments ...rest.Segment) bool {
	oldSegments, err := r.getSegments(envID)
	if err != nil {
		// If we get an error return true to force a cache refresh
		return true
	}

	oldSegmentsMap := map[string]rest.Segment{}
	for _, v := range oldSegments {
		oldSegmentsMap[v.Identifier] = v
	}

	for _, seg := range segments {
		os, ok := oldSegmentsMap[seg.Identifier]
		if !ok {
			// If a new flag isn't in the oldFlagMap then the list of old flags are outdated and we'll
			// want to refresh the cache
			return true
		}

		if *os.Version < *seg.Version {
			return true
		}
	}
	return false

	for _, segment := range segments {
		if r.isSegmentOutdated(segment) {
			return true
		}
	}
	return false
}

// Close all resources
func (r FFRepository) Close() {

}

func formatFlagKey(identifier string) string {
	return "flag/" + identifier
}

func formatFlagsKey(envID string) string {
	return "flags/" + envID
}

func formatSegmentKey(identifier string) string {
	return "target-segment/" + identifier
}

func formatSegmentsKey(envID string) string {
	return "target-segments/" + envID
}
