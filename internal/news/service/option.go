package news

import "github.com/cshep4/news-api/internal/news"

type option func(*service)

func WithProvider(key news.Provider, provider Provider) option {
	return func(s *service) {
		s.providers[key] = provider
	}
}

func WithCategory(category news.Category) option {
	return func(s *service) {
		s.categories[category] = struct {}{}
	}
}
