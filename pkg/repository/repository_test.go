package repository

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/harness/ff-golang-server-sdk/logger"
	"github.com/harness/ff-golang-server-sdk/rest"
	"github.com/stretchr/testify/assert"
)

func int64Ptr(i int) *int64 {
	ptr := int64(i)
	return &ptr
}

var (
	featureOne = rest.FeatureConfig{
		DefaultServe:         rest.Serve{},
		Environment:          "123",
		Feature:              "one",
		Kind:                 "boolean",
		OffVariation:         "false",
		Prerequisites:        nil,
		Project:              "one",
		Rules:                nil,
		State:                "on",
		VariationToTargetMap: nil,
		Variations:           nil,
		Version:              int64Ptr(2),
	}

	featureTwo = rest.FeatureConfig{
		DefaultServe:         rest.Serve{},
		Environment:          "123",
		Feature:              "two",
		Kind:                 "boolean",
		OffVariation:         "false",
		Prerequisites:        nil,
		Project:              "two",
		Rules:                nil,
		State:                "on",
		VariationToTargetMap: nil,
		Variations:           nil,
		Version:              int64Ptr(2),
	}

	segmentOne = rest.Segment{
		CreatedAt:  int64Ptr(0),
		Excluded:   nil,
		Identifier: "one",
		Included:   nil,
		ModifiedAt: nil,
		Name:       "one",
		Rules:      nil,
		Tags:       nil,
		Version:    int64Ptr(2),
	}

	segmentTwo = rest.Segment{
		CreatedAt:  int64Ptr(0),
		Excluded:   nil,
		Identifier: "two",
		Included:   nil,
		ModifiedAt: nil,
		Name:       "two",
		Rules:      nil,
		Tags:       nil,
		Version:    int64Ptr(2),
	}
)

type mockCache struct {
	features []rest.FeatureConfig
	segments []rest.Segment
}

func (m *mockCache) Set(key, value interface{}) (evicted bool) {
	skey := key.(string)

	// If we're setting the key for all the flags then we just want to
	// completely overwrite the features slice
	if strings.Contains(skey, "flags/") {
		slice, ok := value.([]rest.FeatureConfig)
		if !ok {
			return false
		}
		m.features = slice
		return false
	}

	if strings.Contains(skey, "flag/") {
		f, ok := value.(rest.FeatureConfig)
		if !ok {
			return false
		}

		// If the features slice is empty we can just append
		if len(m.features) == 0 {
			m.features = append(m.features, f)
			return false
		}

		// Otherwise we need to update any flags that exist
		for i := 0; i < len(m.features); i++ {
			ff := m.features[i]
			if ff.Feature == f.Feature {
				m.features[i] = f
			}
		}
	}

	// If we're setting the key for all the flags then we just want to
	// completely overwrite the features slice
	if strings.Contains(skey, "target-segments/") {
		slice, ok := value.([]rest.Segment)
		if !ok {
			return false
		}
		m.segments = slice
		return false
	}

	if strings.Contains(skey, "target-segment/") {
		s, ok := value.(rest.Segment)
		if !ok {
			return false
		}

		// If the features slice is empty we can just append
		if len(m.segments) == 0 {
			m.segments = append(m.segments, s)
			return false
		}

		// Otherwise we need to update any flags that exist
		for i := 0; i < len(m.segments); i++ {
			ss := m.segments[i]
			if ss.Identifier == s.Identifier {
				m.segments[i] = s
			}
		}
	}

	return false
}

func (m *mockCache) Contains(key interface{}) bool {
	return false
}

func (m *mockCache) Get(key interface{}) (value interface{}, ok bool) {
	s, ok := key.(string)
	if !ok {
		return nil, false
	}

	if s == "flags/123" {
		return m.features, true
	}

	if s == "target-segments/123" {
		return m.segments, true
	}

	if strings.Contains(s, "target-segment") {
		seg := strings.TrimPrefix(s, "target-segment/")

		for _, ss := range m.segments {
			if ss.Identifier == seg {
				return ss, true
			}
		}
		return nil, false
	}

	if strings.Contains(s, "flag") {
		featureName := strings.TrimPrefix(s, "flag/")

		for _, f := range m.features {
			if f.Feature == featureName {
				return f, true
			}
		}
	}
	return nil, false
}

func (m *mockCache) Keys() []interface{} {
	return nil
}

func (m *mockCache) Len() int {
	return 0
}

func (m *mockCache) Purge() {
}

func (m *mockCache) Remove(key interface{}) (present bool) {
	return false
}

func (m *mockCache) Resize(size int) (evicted int) {
	return 0
}

func (m *mockCache) Updated() time.Time {
	return time.Now()
}

func (m *mockCache) SetLogger(logger logger.Logger) {
}

type mockCallback struct {
	*sync.Mutex
	onFlagsStored    int
	onSegmentsStored int
	onFlagStored     int
	onSegmentStored  int
}

func (m *mockCallback) getOnFlagsStoredCalls() int {
	m.Lock()
	defer m.Unlock()

	return m.onFlagsStored
}

func (m *mockCallback) getOnFlagStoredCalls() int {
	m.Lock()
	defer m.Unlock()

	return m.onFlagStored
}

func (m mockCallback) getOnSegmentsStoredCalls() int {
	m.Lock()
	defer m.Unlock()

	return m.onSegmentsStored
}

func (m mockCallback) getOnSegmentStoredCalls() int {
	m.Lock()
	defer m.Unlock()

	return m.onSegmentStored
}

func (m *mockCallback) OnFlagStored(identifier string) {
	m.Lock()
	defer m.Unlock()
	m.onFlagStored++
}

func (m *mockCallback) OnFlagsStored(envID string) {
	m.Lock()
	defer m.Unlock()
	m.onFlagsStored++
}

func (m *mockCallback) OnFlagDeleted(identifier string) {}

func (m *mockCallback) OnSegmentStored(identifier string) {
	m.Lock()
	defer m.Unlock()

	m.onSegmentStored++
}

func (m *mockCallback) OnSegmentsStored(envID string) {
	m.Lock()
	defer m.Unlock()
	m.onSegmentsStored++
}

func (m *mockCallback) OnSegmentDeleted(identifier string) {}

func (m *mockCallback) OnSegmentsDeleted(envID string, identifier string) {}

func (m *mockCallback) OnFlagsDeleted(envID string, identifier string) {}

func TestFFRepository_SetFlags(t *testing.T) {
	type args struct {
		initialLoad bool
		envID       string
		features    []rest.FeatureConfig
	}

	outdatedFeatureOne := featureOne
	outdatedFeatureOne.Version = int64Ptr(int(*featureOne.Version - 1))

	outdatedFeatureTwo := featureTwo
	outdatedFeatureTwo.Version = int64Ptr(int(*featureTwo.Version - 1))

	type mocks struct {
		cache    *mockCache
		callback *mockCallback
	}

	type results struct {
		cachedFeatures []rest.FeatureConfig
		callbackCalls  int
	}

	testCases := map[string]struct {
		args     args
		mocks    mocks
		expected results
	}{
		"Given initialLoad=true and I try to store two features": {
			args: args{
				initialLoad: true,
				envID:       "123",
				features:    []rest.FeatureConfig{featureOne, featureTwo},
			},

			mocks: mocks{
				cache:    &mockCache{},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedFeatures: []rest.FeatureConfig{featureOne, featureTwo},
				callbackCalls:  1,
			},
		},
		"Given initialLoad=false, I have a cache with two features and I call SetFlags with a list of flags where one of the flags is newer": {
			args: args{
				initialLoad: false,
				envID:       "123",
				features:    []rest.FeatureConfig{featureOne, featureTwo},
			},

			mocks: mocks{
				cache:    &mockCache{features: []rest.FeatureConfig{outdatedFeatureOne, outdatedFeatureTwo}},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedFeatures: []rest.FeatureConfig{featureOne, featureTwo},
				callbackCalls:  1,
			},
		},
		"Given initialLoad=false, I have a cache with two features and I call SetFlags with a list of flags where none of the flags are newer": {
			args: args{
				initialLoad: false,
				envID:       "123",
				features:    []rest.FeatureConfig{featureOne, featureTwo},
			},

			mocks: mocks{
				cache:    &mockCache{features: []rest.FeatureConfig{featureOne, featureTwo}},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedFeatures: []rest.FeatureConfig{featureOne, featureTwo},
				callbackCalls:  0,
			},
		},
	}

	for desc, tc := range testCases {
		desc := desc
		tc := tc

		t.Run(desc, func(t *testing.T) {

			repo := FFRepository{
				cache:    tc.mocks.cache,
				callback: tc.mocks.callback,
			}
			repo.SetFlags(tc.args.initialLoad, tc.args.envID, tc.args.features...)

			assert.Equal(t, tc.expected.cachedFeatures, tc.mocks.cache.features)
			assert.Equal(t, tc.expected.callbackCalls, tc.mocks.callback.getOnFlagsStoredCalls())
		})
	}
}

func TestFFRepository_SetSegments(t *testing.T) {
	type args struct {
		initialLoad bool
		envID       string
		segments    []rest.Segment
	}

	type mocks struct {
		cache    *mockCache
		callback *mockCallback
	}

	type results struct {
		cachedSegments []rest.Segment
		callbackCalls  int
	}

	outdatedSegmentOne := segmentOne
	outdatedSegmentOne.Version = int64Ptr(int(*segmentOne.Version) - 1)

	outdatedSegmentTwo := segmentTwo
	outdatedSegmentTwo.Version = int64Ptr(int(*segmentTwo.Version) - 1)

	testCases := map[string]struct {
		args     args
		mocks    mocks
		expected results
	}{
		"Given initialLoad=true and I try to store two segments": {
			args: args{
				initialLoad: true,
				envID:       "123",
				segments:    []rest.Segment{segmentOne, segmentTwo},
			},

			mocks: mocks{
				cache:    &mockCache{},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedSegments: []rest.Segment{segmentOne, segmentTwo},
				callbackCalls:  1,
			},
		},

		"Given initialLoad=false, I have a cache with two segments and I call SetSegments with a list of segments where one of the segments is newer": {
			args: args{
				initialLoad: false,
				envID:       "123",
				segments:    []rest.Segment{segmentOne, segmentTwo},
			},

			mocks: mocks{
				cache:    &mockCache{segments: []rest.Segment{outdatedSegmentOne, outdatedSegmentTwo}},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedSegments: []rest.Segment{segmentOne, segmentTwo},
				callbackCalls:  1,
			},
		},
		"Given initialLoad=false, I have a cache with two segments and I call SetSegments with a list of segments where none of the segments are newer": {
			args: args{
				initialLoad: false,
				envID:       "123",
				segments:    []rest.Segment{segmentOne, segmentTwo},
			},

			mocks: mocks{
				cache:    &mockCache{segments: []rest.Segment{segmentOne, segmentTwo}},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedSegments: []rest.Segment{segmentOne, segmentTwo},
				callbackCalls:  0,
			},
		},
	}

	for desc, tc := range testCases {
		desc := desc
		tc := tc

		t.Run(desc, func(t *testing.T) {

			repo := FFRepository{
				cache:    tc.mocks.cache,
				callback: tc.mocks.callback,
			}
			repo.SetSegments(tc.args.initialLoad, tc.args.envID, tc.args.segments...)

			assert.Equal(t, tc.expected.cachedSegments, tc.mocks.cache.segments)
			assert.Equal(t, tc.expected.callbackCalls, tc.mocks.callback.getOnSegmentsStoredCalls())
		})
	}
}

func TestFFRepository_SetFlag(t *testing.T) {
	type args struct {
		initialLoad bool
		envID       string
		feature     rest.FeatureConfig
	}

	outdatedFeatureOne := featureOne
	outdatedFeatureOne.Version = int64Ptr(int(*featureOne.Version - 1))

	outdatedFeatureTwo := featureTwo
	outdatedFeatureTwo.Version = int64Ptr(int(*featureTwo.Version - 1))

	type mocks struct {
		cache    *mockCache
		callback *mockCallback
	}

	type results struct {
		cachedFeatures []rest.FeatureConfig
		callbackCalls  int
	}

	testCases := map[string]struct {
		args     args
		mocks    mocks
		expected results
	}{
		"Given initialLoad=true and I call SetFlag": {
			args: args{
				initialLoad: true,
				envID:       "123",
				feature:     featureOne,
			},

			mocks: mocks{
				cache:    &mockCache{},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedFeatures: []rest.FeatureConfig{featureOne},
				callbackCalls:  1,
			},
		},
		"Given initialLoad=false, I have a cache with two features and I call SetFlag with an updated version for one of the features": {
			args: args{
				initialLoad: false,
				envID:       "123",
				feature:     featureTwo,
			},

			mocks: mocks{
				cache:    &mockCache{features: []rest.FeatureConfig{featureOne, outdatedFeatureTwo}},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedFeatures: []rest.FeatureConfig{featureOne, featureTwo},
				callbackCalls:  1,
			},
		},
	}

	for desc, tc := range testCases {
		desc := desc
		tc := tc

		t.Run(desc, func(t *testing.T) {

			repo := FFRepository{
				cache:    tc.mocks.cache,
				callback: tc.mocks.callback,
			}
			repo.SetFlag(tc.args.feature, tc.args.initialLoad)

			assert.Equal(t, tc.expected.cachedFeatures, tc.mocks.cache.features)
			assert.Equal(t, tc.expected.callbackCalls, tc.mocks.callback.getOnFlagStoredCalls())
		})
	}
}

func TestFFRepository_SetSegment(t *testing.T) {
	type args struct {
		initialLoad bool
		envID       string
		segment     rest.Segment
	}

	type mocks struct {
		cache    *mockCache
		callback *mockCallback
	}

	type results struct {
		cachedSegments []rest.Segment
		callbackCalls  int
	}

	outdatedSegmentOne := segmentOne
	outdatedSegmentOne.Version = int64Ptr(int(*segmentOne.Version) - 1)

	outdatedSegmentTwo := segmentTwo
	outdatedSegmentTwo.Version = int64Ptr(int(*segmentTwo.Version) - 1)

	testCases := map[string]struct {
		args     args
		mocks    mocks
		expected results
	}{
		"Given initialLoad=true and I call SetFlag": {
			args: args{
				initialLoad: true,
				envID:       "123",
				segment:     segmentOne,
			},

			mocks: mocks{
				cache:    &mockCache{},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedSegments: []rest.Segment{segmentOne},
				callbackCalls:  1,
			},
		},
		"Given initialLoad=false, I have a cache with two segments and I call SetSegment with an updated version for one of the segments": {
			args: args{
				initialLoad: false,
				envID:       "123",
				segment:     segmentTwo,
			},

			mocks: mocks{
				cache:    &mockCache{segments: []rest.Segment{segmentOne, outdatedSegmentTwo}},
				callback: &mockCallback{Mutex: &sync.Mutex{}},
			},

			expected: results{
				cachedSegments: []rest.Segment{segmentOne, segmentTwo},
				callbackCalls:  1,
			},
		},
	}

	for desc, tc := range testCases {
		desc := desc
		tc := tc

		t.Run(desc, func(t *testing.T) {

			repo := FFRepository{
				cache:    tc.mocks.cache,
				callback: tc.mocks.callback,
			}
			repo.SetSegment(tc.args.segment, tc.args.initialLoad)

			assert.Equal(t, tc.expected.cachedSegments, tc.mocks.cache.segments)
			assert.Equal(t, tc.expected.callbackCalls, tc.mocks.callback.getOnSegmentStoredCalls())
		})
	}
}
