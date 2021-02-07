package news_test

import (
	"context"
	"errors"
	provider_mock "github.com/cshep4/news-api/internal/mock/provider"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cshep4/news-api/internal/mock/cache"
	"github.com/cshep4/news-api/internal/news"
	"github.com/cshep4/news-api/internal/news/handler/http"
	service "github.com/cshep4/news-api/internal/news/service"
)

type testError string

func (e testError) Error() string { return string(e) }

func TestNew_Error(t *testing.T) {
	testCases := []struct {
		name                   string
		cache                  service.Cache
		expectedErrorParameter string
	}{
		{
			name:                   "cache is empty",
			cache:                  nil,
			expectedErrorParameter: "cache",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := service.New(tc.cache)
			require.Error(t, err)
			require.Nil(t, service)

			ipe, ok := err.(news.InvalidParameterError)
			require.True(t, ok)

			assert.Equal(t, tc.expectedErrorParameter, ipe.Parameter)
		})
	}
}

func TestNew_Success(t *testing.T) {
	testCases := []struct {
		name  string
		cache service.Cache
	}{
		{
			name:  "successfully create service",
			cache: cache_mock.NewMockCache(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := service.New(tc.cache)
			require.NoError(t, err)
			require.NotNil(t, service)

			assert.Implements(t, (*http.NewsService)(nil), service)
		})
	}
}

func TestService_GetFeed_Error(t *testing.T) {
	const testErr = testError("error")
	testCases := []struct {
		name        string
		provider    news.Provider
		mockTimes   int
		getFeedErr  error
		expectedErr error
	}{
		{
			name:        "invalid provider",
			provider:    "invalid provider",
			mockTimes:   0,
			expectedErr: news.ErrProviderNotFound,
		},
		{
			name:        "error getting feed",
			provider:    news.ProviderAll,
			mockTimes:   1,
			getFeedErr:  testErr,
			expectedErr: testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl, ctx := gomock.WithContext(context.Background(), t)
			defer ctrl.Finish()

			cache := cache_mock.NewMockCache(ctrl)
			provider := provider_mock.NewMockProvider(ctrl)

			cache.EXPECT().
				Get(tc.provider, news.CategoryUK).
				Return(nil, false).
				Times(tc.mockTimes)

			provider.EXPECT().
				GetFeed(ctx, news.CategoryUK).
				Return(nil, tc.getFeedErr).
				Times(tc.mockTimes)

			service, err := service.New(cache,
				service.WithProvider(news.ProviderAll, provider),
				service.WithCategory(news.CategoryUK),
			)
			require.NoError(t, err)

			res, err := service.GetFeed(ctx, tc.provider, 0, 0)
			require.Error(t, err)
			require.Nil(t, res)

			assert.True(t, errors.Is(err, tc.expectedErr))
		})
	}
}

func TestService_GetFeed_Success(t *testing.T) {
	ctrl, ctx := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	var (
		item1 = news.Item{DateTime: time.Now().Add(time.Second)}
		item2 = news.Item{DateTime: time.Now()}
	)

	type provider struct {
		name     news.Provider
		provider *provider_mock.MockProvider
		items    []news.Item
		cached   bool
	}
	testCases := []struct {
		name           string
		providers      []provider
		category       news.Category
		provider       news.Provider
		limit          int
		offset         int
		cacheFeed      bool
		expectedResult *news.FeedResponse
	}{
		{
			name: "return cached data",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			expectedResult: &news.FeedResponse{
				Provider: news.ProviderAll,
				Items:    []news.Item{item1, item2},
			},
		},
		{
			name: "retrieve data",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   false,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   false,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			expectedResult: &news.FeedResponse{
				Provider: news.ProviderAll,
				Items:    []news.Item{item1, item2},
			},
		},
		{
			name:      "no providers enabled",
			providers: []provider{},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			expectedResult: &news.FeedResponse{
				Provider: news.ProviderAll,
			},
		},
		{
			name: "limit results",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			limit:     1,
			expectedResult: &news.FeedResponse{
				Provider: news.ProviderAll,
				Items:    []news.Item{item1},
				Limit:    1,
			},
		},
		{
			name: "offset results",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			offset:    1,
			expectedResult: &news.FeedResponse{
				Provider: news.ProviderAll,
				Items:    []news.Item{item2},
				Offset:   1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := cache_mock.NewMockCache(ctrl)

			opts := []service.Option{
				service.WithCategory(tc.category),
			}

			for _, p := range tc.providers {
				feed := &news.Feed{Items: p.items}

				cache.EXPECT().Get(p.name, tc.category).Return(feed, p.cached)

				if !p.cached {
					p.provider.EXPECT().GetFeed(ctx, tc.category).Return(feed, nil)
					cache.EXPECT().Store(p.name, tc.category, *feed)
				}

				opts = append(opts, service.WithProvider(p.name, p.provider))
			}

			service, err := service.New(cache, opts...)
			require.NoError(t, err)

			res, err := service.GetFeed(ctx, tc.provider, tc.offset, tc.limit)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedResult, res)
		})
	}
}

func TestService_GetFeedByCategory_Error(t *testing.T) {
	const testErr = testError("error")
	testCases := []struct {
		name        string
		provider    news.Provider
		category    news.Category
		mockTimes   int
		getFeedErr  error
		expectedErr error
	}{
		{
			name:        "invalid provider",
			provider:    "invalid provider",
			category:    news.CategoryUK,
			mockTimes:   0,
			expectedErr: news.ErrProviderNotFound,
		},
		{
			name:        "invalid category",
			provider:    news.ProviderAll,
			category:    "invalid category",
			mockTimes:   0,
			expectedErr: news.ErrCategoryNotFound,
		},
		{
			name:        "error getting feed",
			provider:    news.ProviderAll,
			category:    news.CategoryUK,
			mockTimes:   1,
			getFeedErr:  testErr,
			expectedErr: testErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl, ctx := gomock.WithContext(context.Background(), t)
			defer ctrl.Finish()

			cache := cache_mock.NewMockCache(ctrl)
			provider := provider_mock.NewMockProvider(ctrl)

			cache.EXPECT().
				Get(tc.provider, news.CategoryUK).
				Return(nil, false).
				Times(tc.mockTimes)

			provider.EXPECT().
				GetFeed(ctx, news.CategoryUK).
				Return(nil, tc.getFeedErr).
				Times(tc.mockTimes)

			service, err := service.New(cache,
				service.WithProvider(news.ProviderAll, provider),
				service.WithCategory(news.CategoryUK),
			)
			require.NoError(t, err)

			res, err := service.GetFeedByCategory(ctx, tc.provider, tc.category, 0, 0)
			require.Error(t, err)
			require.Nil(t, res)

			assert.True(t, errors.Is(err, tc.expectedErr))
		})
	}
}

func TestService_GetFeedByCategory_Success(t *testing.T) {
	ctrl, ctx := gomock.WithContext(context.Background(), t)
	defer ctrl.Finish()

	var (
		item1 = news.Item{DateTime: time.Now().Add(time.Second)}
		item2 = news.Item{DateTime: time.Now()}
	)

	type provider struct {
		name     news.Provider
		provider *provider_mock.MockProvider
		items    []news.Item
		cached   bool
	}
	testCases := []struct {
		name           string
		providers      []provider
		category       news.Category
		provider       news.Provider
		limit          int
		offset         int
		cacheFeed      bool
		expectedResult *news.FeedResponse
	}{
		{
			name: "return cached data",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			expectedResult: &news.FeedResponse{
				Category: news.CategoryUK,
				Provider: news.ProviderAll,
				Items:    []news.Item{item1, item2},
			},
		},
		{
			name: "retrieve data",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   false,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   false,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			expectedResult: &news.FeedResponse{
				Category: news.CategoryUK,
				Provider: news.ProviderAll,
				Items:    []news.Item{item1, item2},
			},
		},
		{
			name:      "no providers enabled",
			providers: []provider{},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			expectedResult: &news.FeedResponse{
				Category: news.CategoryUK,
				Provider: news.ProviderAll,
			},
		},
		{
			name: "limit results",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			limit:     1,
			expectedResult: &news.FeedResponse{
				Category: news.CategoryUK,
				Provider: news.ProviderAll,
				Items:    []news.Item{item1},
				Limit:    1,
			},
		},
		{
			name: "offset results",
			providers: []provider{
				{
					name:     news.ProviderBBC,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item1},
				},
				{
					name:     news.ProviderSky,
					provider: provider_mock.NewMockProvider(ctrl),
					cached:   true,
					items:    []news.Item{item2},
				},
			},
			category:  news.CategoryUK,
			provider:  news.ProviderAll,
			cacheFeed: false,
			offset:    1,
			expectedResult: &news.FeedResponse{
				Category: news.CategoryUK,
				Provider: news.ProviderAll,
				Items:    []news.Item{item2},
				Offset:   1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := cache_mock.NewMockCache(ctrl)

			opts := []service.Option{
				service.WithCategory(tc.category),
			}

			for _, p := range tc.providers {
				feed := &news.Feed{Items: p.items}

				cache.EXPECT().Get(p.name, tc.category).Return(feed, p.cached)

				if !p.cached {
					p.provider.EXPECT().GetFeed(ctx, tc.category).Return(feed, nil)
					cache.EXPECT().Store(p.name, tc.category, *feed)
				}

				opts = append(opts, service.WithProvider(p.name, p.provider))
			}

			service, err := service.New(cache, opts...)
			require.NoError(t, err)

			res, err := service.GetFeedByCategory(ctx, tc.provider, tc.category, tc.offset, tc.limit)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
