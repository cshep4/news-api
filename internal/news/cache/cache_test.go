package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cshep4/news-api/internal/news"
	"github.com/cshep4/news-api/internal/news/cache"
	service "github.com/cshep4/news-api/internal/news/service"
)

func TestNew_Error(t *testing.T) {
	testCases := []struct {
		name                   string
		clock                  clockwork.Clock
		expectedErrorParameter string
	}{
		{
			name:                   "clock is empty",
			clock:                  nil,
			expectedErrorParameter: "clock",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache, err := cache.New(tc.clock)
			require.Error(t, err)
			require.Nil(t, cache)

			ipe, ok := err.(news.InvalidParameterError)
			require.True(t, ok)

			assert.Equal(t, tc.expectedErrorParameter, ipe.Parameter)
		})
	}
}

func TestNew_Success(t *testing.T) {
	testCases := []struct {
		name  string
		clock clockwork.Clock
	}{
		{
			name:  "successfully create cache",
			clock: clockwork.NewFakeClock(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache, err := cache.New(tc.clock)
			require.NoError(t, err)
			require.NotNil(t, cache)

			assert.Implements(t, (*service.Cache)(nil), cache)
		})
	}
}

func TestCache_Get(t *testing.T) {
	const (
		provider = news.Provider("provider")
		category = news.Category("category")
	)
	feed := news.Feed{Title: "feed"}

	testCases := []struct {
		name           string
		provider       news.Provider
		category       news.Category
		expectedFeed   *news.Feed
		expectedExists bool
	}{
		{
			name:           "feed not in cache",
			provider:       "fake provider",
			category:       "fake category",
			expectedFeed:   nil,
			expectedExists: false,
		},
		{
			name:           "feed retrieved from cache",
			provider:       provider,
			category:       category,
			expectedFeed:   &feed,
			expectedExists: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache, err := cache.New(clockwork.NewFakeClock())
			require.NoError(t, err)
			require.NotNil(t, cache)

			cache.Store(provider, category, feed)

			res, ok := cache.Get(tc.provider, tc.category)
			require.Equal(t, tc.expectedExists, ok)

			assert.Equal(t, tc.expectedFeed, res)
		})
	}
}

func TestCache_Store(t *testing.T) {
	const (
		provider = news.Provider("provider")
		category = news.Category("category")
	)
	feed := news.Feed{Title: "feed", TTL: 10}

	testCases := []struct {
		name           string
		advanceTime    int
		expectedFeed   *news.Feed
		expectedExists bool
	}{
		{
			name:           "cache feed",
			advanceTime:    0,
			expectedFeed:   &feed,
			expectedExists: true,
		},
		{
			name:           "cache invalidated after ttl",
			advanceTime:    100,
			expectedFeed:   nil,
			expectedExists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clock := clockwork.NewFakeClock()

			cache, err := cache.New(clock)
			require.NoError(t, err)
			require.NotNil(t, cache)
			
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				cache.Store(provider, category, feed)
				wg.Done()
			}()

			clock.BlockUntil(1)
			clock.Advance(time.Minute * time.Duration(tc.advanceTime))

			// have to sleep to ensure cache invalidation has had chance to run as
			// this is inside a go routine
			time.Sleep(5 * time.Millisecond)

			wg.Wait()

			res, ok := cache.Get(provider, category)
			require.Equal(t, tc.expectedExists, ok)

			assert.Equal(t, tc.expectedFeed, res)
		})
	}
}
