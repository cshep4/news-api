package sky

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cshep4/news-api/internal/news"
)

type adapter struct {
	url    string
	client *http.Client
}

func New(url string, client *http.Client) (*adapter, error) {
	switch {
	case url == "":
		return nil, news.InvalidParameterError{Parameter: "url"}
	case client == nil:
		return nil, news.InvalidParameterError{Parameter: "client"}
	}

	return &adapter{
		url:    url,
		client: client,
	}, nil
}

func (a *adapter) GetFeed(ctx context.Context, category news.Category) (*news.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.buildUrl(category), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}

	var response Response
	if err := xml.Unmarshal(b, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %v", err)
	}

	return response.toFeed(category), nil
}

func (a *adapter) buildUrl(category news.Category) string {
	return fmt.Sprintf("%s/%s.xml", a.url, category)
}
