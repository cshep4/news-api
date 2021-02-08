package gen

//go:generate mockgen -destination=internal/mock/service/mock_service.gen.go -package=service_mock github.com/cshep4/news-api/internal/news/handler/http NewsService
//go:generate mockgen -destination=internal/mock/cache/mock_cache.gen.go -package=cache_mock github.com/cshep4/news-api/internal/news/service Cache
//go:generate mockgen -destination=internal/mock/provider/mock_provider.gen.go -package=provider_mock github.com/cshep4/news-api/internal/news/service Provider
