package news

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/cshep4/news-api/internal/news"
)

type (
	Provider interface {
		GetFeed(ctx context.Context, category news.Category) (*news.Feed, error)
	}

	Cache interface {
		Get(provider news.Provider, category news.Category) (*news.Feed, bool)
		Store(provider news.Provider, category news.Category, feed news.Feed)
	}

	service struct {
		cache      Cache
		providers  map[news.Provider]Provider
		categories map[news.Category]struct{}
	}
)

func New(cache Cache, opts ...option) (*service, error) {
	if cache == nil {
		return nil, news.InvalidParameterError{Parameter: "cache"}
	}

	s := &service{
		cache:      cache,
		providers:  make(map[news.Provider]Provider),
		categories: make(map[news.Category]struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *service) GetFeedByCategory(ctx context.Context, provider news.Provider, category news.Category, offset, limit int) (*news.FeedResponse, error) {
	if _, ok := s.categories[category]; !ok {
		return nil, news.ErrCategoryNotFound
	}

	items, err := s.getFeeds(ctx, provider, category)
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].DateTime.After(items[j].DateTime)
	})

	return &news.FeedResponse{
		Category: category,
		Provider: provider,
		Items:    s.paginate(items, offset, limit),
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *service) GetFeed(ctx context.Context, provider news.Provider, offset, limit int) (*news.FeedResponse, error) {
	var items []news.Item

	for c := range s.categories {
		feed, err := s.getFeeds(ctx, provider, c)
		if err != nil {
			return nil, err
		}

		items = append(items, feed...)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].DateTime.After(items[j].DateTime)
	})

	return &news.FeedResponse{
		Provider: provider,
		Items:    s.paginate(items, offset, limit),
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *service) getFeeds(ctx context.Context, provider news.Provider, category news.Category) ([]news.Item, error) {
	if provider == news.ProviderAll {
		return s.getAllFeedsForCategory(ctx, category)
	}

	feed, err := s.getFeed(ctx, provider, category)
	if err != nil {
		return nil, err
	}

	return feed.Items, nil
}

func (s *service) getAllFeedsForCategory(ctx context.Context, category news.Category) ([]news.Item, error) {
	var items []news.Item

	for p := range s.providers {
		feed, err := s.getFeed(ctx, p, category)
		if err != nil {
			return nil, err
		}
		items = append(items, feed.Items...)
	}

	return items, nil
}

func (s *service) getFeed(ctx context.Context, provider news.Provider, category news.Category) (*news.Feed, error) {
	newsProvider, ok := s.providers[provider]
	if !ok {
		return nil, news.ErrProviderNotFound
	}

	feed, ok := s.cache.Get(provider, category)
	if ok {
		return feed, nil
	}

	feed, err := newsProvider.GetFeed(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s feed from %s: %w", category, provider, err)
	}

	s.cache.Store(provider, category, *feed)

	return feed, nil
}

func (s *service) paginate(items []news.Item, offset, limit int) []news.Item {
	if offset > len(items) {
		offset = len(items)
	}

	if limit == 0 {
		limit = math.MaxInt32
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	return items[offset:end]
}
