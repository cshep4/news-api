package news

import "time"

const (
	ProviderAll Provider = ""
	ProviderBBC Provider = "bbc"
	ProviderSky Provider = "sky"

	CategoryUK         Category = "uk"
	CategoryTechnology Category = "technology"
)

type (
	Provider string
	Category string

	Feed struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Link        string    `json:"link"`
		Language    string    `json:"language"`
		Copyright   string    `json:"copyright"`
		DateTime    time.Time `json:"dateTime"`
		TTL         int       `json:"ttl"`
		Items       []Item    `json:"items"`
	}

	FeedResponse struct {
		Category Category `json:"category,omitempty"`
		Provider Provider `json:"provider,omitempty"`
		Items    []Item   `json:"items"`
		Limit    int      `json:"limit,omitempty"`
		Offset   int      `json:"offset,omitempty"`
	}

	Item struct {
		Category    Category  `json:"category"`
		Provider    Provider  `json:"provider"`
		Title       string    `json:"title"`
		Link        string    `json:"link"`
		Description string    `json:"description"`
		Thumbnail   string    `json:"thumbnail"`
		DateTime    time.Time `json:"dateTime"`
	}
)
