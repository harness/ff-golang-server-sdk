package repository

import (
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
		Version:              int64Ptr(1),
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
		Version:    nil,
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
		Version:    nil,
	}
)

type mockCache struct {
	features []rest.FeatureConfig
	segments []rest.Segment
}

func (m *mockCache) Set(key, value interface{}) (evicted bool) {
	if v, ok := value.([]rest.FeatureConfig); ok {
		for _, f := range v {
			m.features = append(m.features, f)
		}
	}

	if v, ok := value.([]rest.Segment); ok {
		for _, f := range v {
			m.segments = append(m.segments, f)
		}
	}

	return false
}

func (m *mockCache) Contains(key interface{}) bool {
	return false
}

func (m *mockCache) Get(key interface{}) (value interface{}, ok bool) {
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
}

func (m *mockCallback) getOnFlagsStoredCalls() int {
	m.Lock()
	defer m.Unlock()

	return m.onFlagsStored
}

func (m mockCallback) getOnSegmentsStoredCalls() int {
	m.Lock()
	defer m.Unlock()

	return m.onSegmentsStored
}

func (m *mockCallback) OnFlagStored(identifier string) {}

func (m *mockCallback) OnFlagsStored(envID string) {
	m.Lock()
	defer m.Unlock()
	m.onFlagsStored++
}

func (m *mockCallback) OnFlagDeleted(identifier string) {}

func (m *mockCallback) OnSegmentStored(identifier string) {
	m.Lock()
	defer m.Unlock()

	m.onSegmentsStored++
}

func (m *mockCallback) OnSegmentsStored(envID string) {
	m.Lock()
	defer m.Unlock()
	m.onSegmentsStored++
}

func (m *mockCallback) OnSegmentDeleted(identifier string) {}

func TestFFRepository_SetFlags(t *testing.T) {
	type args struct {
		initialLoad bool
		envID       string
		features    []rest.FeatureConfig
	}

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
		"Given initialLoad=false and I try to store two features": {
			args: args{
				initialLoad: false,
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
		"Given initialLoad=false and I try to store two features": {
			args: args{
				initialLoad: false,
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
			assert.Equal(t, tc.expected.callbackCalls, tc.mocks.callback.getOnFlagsStoredCalls())
		})
	}
}
