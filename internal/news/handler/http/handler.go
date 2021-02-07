package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/cshep4/news-api/internal/log"
	"github.com/cshep4/news-api/internal/news"
)

type (
	NewsService interface {
		GetFeedByCategory(ctx context.Context, provider news.Provider, category news.Category, offset, limit int) (*news.FeedResponse, error)
		GetFeed(ctx context.Context, provider news.Provider, offset, limit int) (*news.FeedResponse, error)
	}

	handler struct {
		newsService NewsService
	}

	serverError struct {
		Message string `json:"message"`
	}
)

func New(newsService NewsService) (*handler, error) {
	if newsService == nil {
		return nil, news.InvalidParameterError{Parameter: "newsService"}
	}

	return &handler{
		newsService: newsService,
	}, nil
}

func (h *handler) Route(router *mux.Router) {
	router.HandleFunc("/", h.getFeed).
		Methods(http.MethodGet)
	router.HandleFunc("/{category}", h.getFeedByCategory).
		Methods(http.MethodGet)
}

func (h *handler) getFeed(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	provider := news.Provider(r.URL.Query().Get("provider"))
	if provider == "" {
		provider = news.ProviderAll
	}

	limit, err := h.intParam(r.URL.Query(), "limit")
	if err != nil {
		h.errorResponse(r.Context(), http.StatusBadRequest, "limit is invalid", w)
		return
	}

	offset, err := h.intParam(r.URL.Query(), "offset")
	if err != nil {
		h.errorResponse(r.Context(), http.StatusBadRequest, "offset is invalid", w)
		return
	}

	res, err := h.newsService.GetFeed(r.Context(), provider, offset, limit)
	if err != nil {
		log.Error(r.Context(), "error_getting_feed",
			log.SafeParam("provider", provider),
			log.SafeParam("limit", limit),
			log.SafeParam("offset", offset),
			log.ErrorParam(err),
		)
	}
	h.sendResponse(r.Context(), w, res, err)
}

func (h *handler) getFeedByCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	category, ok := mux.Vars(r)["category"]
	if !ok {
		h.errorResponse(r.Context(), http.StatusBadRequest, "category not specified", w)
		return
	}

	limit, err := h.intParam(r.URL.Query(), "limit")
	if err != nil {
		h.errorResponse(r.Context(), http.StatusBadRequest, "limit is invalid", w)
		return
	}

	offset, err := h.intParam(r.URL.Query(), "offset")
	if err != nil {
		h.errorResponse(r.Context(), http.StatusBadRequest, "offset is invalid", w)
		return
	}

	provider := news.Provider(r.URL.Query().Get("provider"))
	if provider == "" {
		provider = news.ProviderAll
	}

	res, err := h.newsService.GetFeedByCategory(r.Context(), provider, news.Category(category), offset, limit)
	if err != nil {
		log.Error(r.Context(), "error_getting_feed_category",
			log.SafeParam("category", category),
			log.SafeParam("provider", provider),
			log.SafeParam("limit", limit),
			log.SafeParam("offset", offset),
			log.ErrorParam(err),
		)
	}
	h.sendResponse(r.Context(), w, res, err)
}

func (h *handler) intParam(values url.Values, key string) (int, error) {
	param := values.Get(key)
	if param == "" {
		return 0, nil
	}

	n, err := strconv.Atoi(param)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (h *handler) sendResponse(ctx context.Context, w http.ResponseWriter, res interface{}, err error) {
	switch {
	case err == nil:
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Error(ctx, "encode_response_error", log.ErrorParam(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case errors.Is(err, news.ErrCategoryNotFound),
		errors.Is(err, news.ErrProviderNotFound):
		h.errorResponse(ctx, http.StatusNotFound, err.Error(), w)
	default:
		h.errorResponse(ctx, http.StatusInternalServerError, "could not get news feed", w)
	}
}

func (h *handler) errorResponse(ctx context.Context, status int, message string, w http.ResponseWriter) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(serverError{Message: message}); err != nil {
		log.Error(ctx, "encode_response_error", log.ErrorParam(err))
		return
	}
}
