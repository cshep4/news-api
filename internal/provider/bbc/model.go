package bbc

import (
	"encoding/xml"
	"time"

	"github.com/cshep4/news-api/internal/news"
)

type (
	resTime time.Time

	Response struct {
		XMLName xml.Name `xml:"rss"`
		Text    string   `xml:",chardata"`
		Dc      string   `xml:"dc,attr"`
		Content string   `xml:"content,attr"`
		Atom    string   `xml:"atom,attr"`
		Version string   `xml:"version,attr"`
		Media   string   `xml:"media,attr"`
		Channel Channel  `xml:"channel"`
	}

	Channel struct {
		Text          string  `xml:",chardata"`
		Title         string  `xml:"title"`
		Description   string  `xml:"description"`
		Link          string  `xml:"link"`
		Image         Image   `xml:"image"`
		Generator     string  `xml:"generator"`
		LastBuildDate resTime `xml:"lastBuildDate"`
		Copyright     string  `xml:"copyright"`
		Language      string  `xml:"language"`
		TTL           int     `xml:"ttl"`
		Items         []Item  `xml:"item"`
	}

	Image struct {
		Text  string `xml:",chardata"`
		URL   string `xml:"url"`
		Title string `xml:"title"`
		Link  string `xml:"link"`
	}

	Item struct {
		Text        string  `xml:",chardata"`
		Title       string  `xml:"title"`
		Description string  `xml:"description"`
		Link        string  `xml:"link"`
		Guid        Guid    `xml:"guid"`
		PubDate     resTime `xml:"pubDate"`
	}

	Guid struct {
		Text        string `xml:",chardata"`
		IsPermaLink string `xml:"isPermaLink,attr"`
	}
)

func (r *resTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	parse, err := time.Parse(time.RFC1123, v)
	if err != nil {
		return err
	}
	*r = resTime(parse)

	return nil
}

func (r resTime) MarshalText() ([]byte, error) {
	text := time.Time(r).Format(time.RFC1123)
	return []byte(text), nil
}

func (r *Response) toFeed(category news.Category) *news.Feed {
	var items []news.Item
	for _, i := range r.Channel.Items {
		items = append(items, news.Item{
			Category:    category,
			Provider:    news.ProviderBBC,
			Title:       i.Title,
			Link:        i.Link,
			Description: i.Description,
			Thumbnail:   r.Channel.Image.URL,
			DateTime:    time.Time(i.PubDate),
		})
	}

	return &news.Feed{
		Title:       r.Channel.Title,
		Description: r.Channel.Description,
		Link:        r.Channel.Link,
		Language:    r.Channel.Language,
		Copyright:   r.Channel.Copyright,
		DateTime:    time.Time(r.Channel.LastBuildDate),
		TTL:         r.Channel.TTL,
		Items:       items,
	}
}
