package http

import (
	"net/http"
)

type ServerError = serverError

func (h *handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	h.getFeed(w, r)
}

func (h *handler) GetFeedByCategory(w http.ResponseWriter, r *http.Request) {
	h.getFeedByCategory(w, r)
}
