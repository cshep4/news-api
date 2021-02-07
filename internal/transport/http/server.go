package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/cshep4/news-api/internal/log"
)

const defaultPort = 8080

type (
	Router interface {
		Route(*mux.Router)
	}

	Registerer interface {
		Register(*mux.Router)
	}

	server struct {
		routers     []Router
		registerers []Registerer
		middlewares []mux.MiddlewareFunc
		https       *http.Server
		port        int
	}
)

func New(opts ...option) *server {
	s := &server{
		port: defaultPort,
		https: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s server) Start(ctx context.Context) error {
	path := fmt.Sprintf(":%d", s.port)

	router := mux.NewRouter()

	for _, m := range s.middlewares {
		router.Use(m)
	}
	for _, r := range s.routers {
		r.Route(router)
	}
	for _, r := range s.registerers {
		r.Register(router)
	}

	s.https.Addr = path
	s.https.Handler = enableCors(router)

	log.Info(ctx, "http_server_listening", log.SafeParam("path", path))

	err := s.https.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to listen and serve: %v", err)
	}

	return nil
}

func enableCors(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowedHeaders: []string{"*"},
	}).
		Handler(h)
}

func (s server) Stop(ctx context.Context) error {
	err := s.https.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown gracefully: %v", err)
	}

	log.Info(ctx, "http_server_stopped")

	return nil
}
