package http_test

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cshep4/news-api/internal/mock/service"
	"github.com/cshep4/news-api/internal/news"
	handler "github.com/cshep4/news-api/internal/news/handler/http"
)

type testError string

func (e testError) Error() string { return string(e) }

func TestNew_Error(t *testing.T) {
	testCases := []struct {
		name                   string
		service                handler.NewsService
		expectedErrorParameter string
	}{
		{
			name:                   "service is empty",
			service:                nil,
			expectedErrorParameter: "newsService",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, err := handler.New(tc.service)
			require.Error(t, err)
			require.Nil(t, handler)

			ipe, ok := err.(news.InvalidParameterError)
			require.True(t, ok)

			assert.Equal(t, tc.expectedErrorParameter, ipe.Parameter)
		})
	}
}

func TestNew_Success(t *testing.T) {
	testCases := []struct {
		name    string
		service handler.NewsService
	}{
		{
			name:    "successfully create handler",
			service: service_mock.NewMockNewsService(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, err := handler.New(tc.service)
			require.NoError(t, err)
			require.NotNil(t, handler)
		})
	}
}

func TestHandler_GetFeed_Error(t *testing.T) {
	testCases := []struct {
		name               string
		provider           news.Provider
		path               string
		testErr            error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "invalid limit",
			path:               "/?limit=limit",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "limit is invalid",
		},
		{
			name:               "invalid offset",
			path:               "/?offset=offset",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "offset is invalid",
		},
		{
			name:               "category not found",
			path:               "/",
			testErr:            news.ErrCategoryNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedError:      news.ErrCategoryNotFound.Error(),
		},
		{
			name:               "provider not found",
			path:               "/",
			testErr:            news.ErrProviderNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedError:      news.ErrProviderNotFound.Error(),
		},
		{
			name:               "internal error",
			path:               "/",
			testErr:            testError("error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "could not get news feed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := service_mock.NewMockNewsService(ctrl)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()

			if tc.testErr != nil {
				service.EXPECT().GetFeed(req.Context(), tc.provider, 0, 0).Return(nil, tc.testErr)
			}

			h, err := handler.New(service)
			require.NoError(t, err)
			require.NotNil(t, h)

			h.GetFeed(rr, req)

			var responseBody handler.ServerError
			require.NoError(t, json.NewDecoder(rr.Result().Body).Decode(&responseBody))

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			assert.Equal(t, tc.expectedError, responseBody.Message)
		})
	}
}

func TestHandler_GetFeed_Success(t *testing.T) {
	testCases := []struct {
		name             string
		provider         news.Provider
		limit            int
		offset           int
		expectedResponse news.FeedResponse
	}{
		{
			name:     "returns feed response",
			limit:    1,
			offset:   2,
			provider: news.ProviderBBC,
			expectedResponse: news.FeedResponse{
				Provider: news.ProviderBBC,
				Limit:    1,
				Offset:   2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := service_mock.NewMockNewsService(ctrl)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?limit=%d&offset=%d&provider=%s", tc.limit, tc.offset, tc.provider), nil)
			rr := httptest.NewRecorder()

			service.EXPECT().GetFeed(req.Context(), tc.provider, tc.offset, tc.limit).Return(&tc.expectedResponse, nil)

			h, err := handler.New(service)
			require.NoError(t, err)
			require.NotNil(t, h)

			h.GetFeed(rr, req)

			var responseBody news.FeedResponse
			require.NoError(t, json.NewDecoder(rr.Result().Body).Decode(&responseBody))

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tc.expectedResponse, responseBody)
		})
	}
}

func TestHandler_GetFeedByCategory_Error(t *testing.T) {
	testCases := []struct {
		name               string
		provider           news.Provider
		category           news.Category
		path               string
		testErr            error
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "missing category",
			path:               "/%s",
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "category not specified",
		},
		{
			name:               "invalid limit",
			path:               "/%s?limit=limit",
			category:           news.CategoryUK,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "limit is invalid",
		},
		{
			name:               "invalid offset",
			path:               "/%s?offset=offset",
			category:           news.CategoryUK,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "offset is invalid",
		},
		{
			name:               "category not found",
			path:               "/%s",
			category:           news.CategoryUK,
			testErr:            news.ErrCategoryNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedError:      news.ErrCategoryNotFound.Error(),
		},
		{
			name:               "provider not found",
			path:               "/%s",
			category:           news.CategoryUK,
			testErr:            news.ErrProviderNotFound,
			expectedStatusCode: http.StatusNotFound,
			expectedError:      news.ErrProviderNotFound.Error(),
		},
		{
			name:               "internal error",
			path:               "/%s",
			category:           news.CategoryUK,
			testErr:            testError("error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError:      "could not get news feed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := service_mock.NewMockNewsService(ctrl)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf(tc.path, tc.category), nil)
			if tc.category != "" {
				req = mux.SetURLVars(req, map[string]string{
					"category": string(tc.category),
				})
			}

			rr := httptest.NewRecorder()

			if tc.testErr != nil {
				service.EXPECT().GetFeedByCategory(req.Context(), tc.provider, tc.category, 0, 0).Return(nil, tc.testErr)
			}

			h, err := handler.New(service)
			require.NoError(t, err)
			require.NotNil(t, h)

			h.GetFeedByCategory(rr, req)

			var responseBody handler.ServerError
			require.NoError(t, json.NewDecoder(rr.Result().Body).Decode(&responseBody))

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			assert.Equal(t, tc.expectedError, responseBody.Message)
		})
	}
}

func TestHandler_GetFeedByCategory_Success(t *testing.T) {
	testCases := []struct {
		name             string
		provider         news.Provider
		category         news.Category
		limit            int
		offset           int
		expectedResponse news.FeedResponse
	}{
		{
			name:     "returns feed response",
			limit:    1,
			offset:   2,
			provider: news.ProviderBBC,
			expectedResponse: news.FeedResponse{
				Provider: news.ProviderBBC,
				Limit:    1,
				Offset:   2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := service_mock.NewMockNewsService(ctrl)

			path := fmt.Sprintf("/%s?limit=%d&offset=%d&provider=%s", tc.category, tc.limit, tc.offset, tc.provider)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req = mux.SetURLVars(req, map[string]string{
				"category": string(tc.category),
			})
			rr := httptest.NewRecorder()

			service.EXPECT().GetFeedByCategory(req.Context(), tc.provider, tc.category, tc.offset, tc.limit).Return(&tc.expectedResponse, nil)

			h, err := handler.New(service)
			require.NoError(t, err)
			require.NotNil(t, h)

			h.GetFeedByCategory(rr, req)

			var responseBody news.FeedResponse
			require.NoError(t, json.NewDecoder(rr.Result().Body).Decode(&responseBody))

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tc.expectedResponse, responseBody)
		})
	}
}
