package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonboulle/clockwork"
	"golang.org/x/sync/errgroup"

	"github.com/cshep4/news-api/internal/log"
	"github.com/cshep4/news-api/internal/news"
	"github.com/cshep4/news-api/internal/news/cache"
	httphandler "github.com/cshep4/news-api/internal/news/handler/http"
	newsservice "github.com/cshep4/news-api/internal/news/service"
	"github.com/cshep4/news-api/internal/provider/bbc"
	"github.com/cshep4/news-api/internal/provider/sky"
	"github.com/cshep4/news-api/internal/secret"
	httptransport "github.com/cshep4/news-api/internal/transport/http"
)

const (
	serviceName = "news-api"
	logLevel    = "info"
	version     = "v1.0.0"
)

func start(ctx context.Context) error {
	var s secret.Secrets

	if err := s.Load(); err != nil {
		return fmt.Errorf("failed to load secrets: %w", err)
	}

	cache, err := cache.New(clockwork.NewRealClock())
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	client := &http.Client{
		Timeout: time.Second,
	}

	skyProvider, err := sky.New(s.SkyURL, client)
	if err != nil {
		return fmt.Errorf("failed to create sky provider: %w", err)
	}

	bbcProvider, err := bbc.New(s.BBCURL, client)
	if err != nil {
		return fmt.Errorf("failed to create bbc provider: %w", err)
	}

	service, err := newsservice.New(cache,
		newsservice.WithProvider(news.ProviderSky, skyProvider),
		newsservice.WithProvider(news.ProviderBBC, bbcProvider),
		newsservice.WithCategory(news.CategoryUK),
		newsservice.WithCategory(news.CategoryTechnology),
	)
	if err != nil {
		return fmt.Errorf("failed to create news service: %w", err)
	}

	handler, err := httphandler.New(service)
	if err != nil {
		return fmt.Errorf("failed to create http handler: %w", err)
	}

	newsServer := httptransport.New(
		httptransport.WithLogger(serviceName, logLevel),
		httptransport.WithRouter(handler),
	)

	healthServer := httptransport.New(
		httptransport.WithPort(8082),
		httptransport.WithRegisterer(httptransport.Health()),
		httptransport.WithRegisterer(httptransport.Live()),
		httptransport.WithRegisterer(httptransport.Version(version)),
	)

	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return newsServer.Start(ctx)
	})

	g.Go(func() error {
		return healthServer.Start(ctx)
	})

	g.Go(func() error {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		select {
		case <-c:
			cancel()
			return nil
		case <-ctx.Done():
			g.Go(func() error { return newsServer.Stop(ctx) })
			g.Go(func() error { return healthServer.Stop(ctx) })
			return ctx.Err()
		}
	})

	return g.Wait()
}

func main() {
	ctx := log.WithServiceName(context.Background(), log.New(logLevel), serviceName)
	if err := start(ctx); err != nil {
		log.Error(ctx, "startup_error", log.ErrorParam(err))
	}
}
