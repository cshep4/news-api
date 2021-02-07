package sky_test

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cshep4/news-api/internal/news"
	service "github.com/cshep4/news-api/internal/news/service"
	"github.com/cshep4/news-api/internal/provider/sky"
)

type testError string

func (e testError) Error() string { return string(e) }

type errorRoundTripper struct{ err error }

func (e errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, e.err
}

func TestNew_Error(t *testing.T) {
	testCases := []struct {
		name                   string
		url                    string
		client                 *http.Client
		expectedErrorParameter string
	}{
		{
			name:                   "url is empty",
			url:                    "",
			expectedErrorParameter: "url",
		},
		{
			name:                   "client is invalid",
			url:                    "https://test.com",
			client:                 nil,
			expectedErrorParameter: "client",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, err := sky.New(tc.url, tc.client)
			require.Error(t, err)
			require.Nil(t, adapter)

			ipe, ok := err.(news.InvalidParameterError)
			require.True(t, ok)

			assert.Equal(t, tc.expectedErrorParameter, ipe.Parameter)
		})
	}
}

func TestNew_Success(t *testing.T) {
	testCases := []struct {
		name   string
		url    string
		client *http.Client
	}{
		{
			name:   "successfully create adapter",
			url:    "test url",
			client: &http.Client{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, err := sky.New(tc.url, tc.client)
			require.NoError(t, err)
			require.NotNil(t, adapter)

			assert.Implements(t, (*service.Provider)(nil), adapter)
		})
	}
}

func TestAdapter_GetFeed_Error(t *testing.T) {
	testCases := []struct {
		name       string
		client     *http.Client
		statusCode int
		expectedEr string
	}{
		{
			name: "request error",
			client: &http.Client{
				Transport: errorRoundTripper{err: testError("error")},
			},
			statusCode: http.StatusOK,
			expectedEr: "failed to do request",
		},
		{
			name:       "status code not 200",
			statusCode: http.StatusTeapot,
			expectedEr: "unexpected status code",
		},
		{
			name:       "invalid response body",
			statusCode: http.StatusOK,
			expectedEr: "failed to unmarshal body",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer s.Close()

			if tc.client == nil {
				tc.client = s.Client()
			}

			adapter, err := sky.New(s.URL, tc.client)
			require.NoError(t, err)
			require.NotNil(t, adapter)

			_, err = adapter.GetFeed(context.Background(), "category")
			require.Error(t, err)

			assert.Contains(t, err.Error(), tc.expectedEr)
		})
	}
}

func TestAdapter_GetFeed_Success(t *testing.T) {
	const (
		title       = "title"
		description = "description"
		link        = "link"
		imageURL    = "image url"
		language    = "language"
		copyright   = "copyright"
		ttl         = 1
	)

	now := time.Now().Round(time.Second)

	testCases := []struct {
		name           string
		apiResponse    sky.Response
		expectedResult *news.Feed
	}{
		{
			name: "successful request",
			apiResponse: sky.Response{
				Channel: sky.Channel{
					Title:         title,
					Description:   description,
					Link:          link,
					LastBuildDate: sky.ResTime(now),
					Copyright:     copyright,
					Language:      language,
					TTL:           ttl,
					Items: []sky.Item{{
						Title:       title,
						Link:        link,
						Description: description,
						PubDate:     sky.ResTime(now),
						Thumbnail:   sky.Thumbnail{URL: imageURL},
					}},
				},
			},
			expectedResult: &news.Feed{
				Title:       title,
				Description: description,
				Link:        link,
				Language:    language,
				Copyright:   copyright,
				DateTime:    now,
				TTL:         ttl,
				Items: []news.Item{
					{
						Category:    "category",
						Provider:    news.ProviderSky,
						Title:       title,
						Link:        link,
						Description: description,
						Thumbnail:   imageURL,
						DateTime:    now,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				require.NoError(t, xml.NewEncoder(w).Encode(tc.apiResponse))
			}))
			defer s.Close()

			adapter, err := sky.New(s.URL, s.Client())
			require.NoError(t, err)
			require.NotNil(t, adapter)

			res, err := adapter.GetFeed(context.Background(), "category")
			require.NoError(t, err)

			assert.Equal(t, tc.expectedResult, res)
		})
	}
}
