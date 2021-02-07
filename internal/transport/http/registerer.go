package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type registerer struct {
	path    string
	f       func(http.ResponseWriter, *http.Request)
	methods []string
}

func NewRegisterer(path string, f func(http.ResponseWriter, *http.Request), methods ...string) registerer {
	return registerer{
		path:    path,
		f:       f,
		methods: methods,
	}
}

func (h registerer) Register(router *mux.Router) {
	router.HandleFunc(h.path, h.f).
		Methods(h.methods...)
}

func Live() registerer {
	return NewRegisterer("/_live",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		http.MethodGet,
	)
}

func Health() registerer {
	return NewRegisterer("/_health",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
		http.MethodGet,
	)
}

func Version(version string) registerer {
	return NewRegisterer("/_version",
		func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(struct {
				Version string `json:"version"`
			}{version})
		},
		http.MethodGet,
	)
}
