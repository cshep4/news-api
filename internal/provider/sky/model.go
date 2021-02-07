package sky

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
		Atom    string   `xml:"atom,attr"`
		Media   string   `xml:"media,attr"`
		Version string   `xml:"version,attr"`
		Channel Channel  `xml:"channel"`
	}

	Channel struct {
		Text          string  `xml:",chardata"`
		Link          string  `xml:"link"`
		Title         string  `xml:"title"`
		Image         Image   `xml:"image"`
		Description   string  `xml:"description"`
		Language      string  `xml:"language"`
		Copyright     string  `xml:"copyright"`
		LastBuildDate resTime `xml:"lastBuildDate"`
		Category      string  `xml:"category"`
		TTL           int     `xml:"ttl"`
		Items         []Item  `xml:"item"`
	}

	Link struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	}

	Image struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		URL   string `xml:"url"`
		Link  string `xml:"link"`
	}

	Item struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		PubDate     resTime   `xml:"pubDate"`
		Guid        string    `xml:"guid"`
		Enclosure   Enclosure `xml:"enclosure"`
		Thumbnail   Thumbnail `xml:"thumbnail"`
		Content     Content   `xml:"content"`
	}

	Description struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
	}

	Enclosure struct {
		Text   string `xml:",chardata"`
		URL    string `xml:"url,attr"`
		Length int    `xml:"length,attr"`
		Type   string `xml:"type,attr"`
	}

	Thumbnail struct {
		Text   string `xml:",chardata"`
		URL    string `xml:"url,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
	}

	Content struct {
		Text string `xml:",chardata"`
		Type string `xml:"type,attr"`
		URL  string `xml:"url,attr"`
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


func (r Response) toFeed(category news.Category) *news.Feed {
	var items []news.Item
	for _, i := range r.Channel.Items {
		items = append(items, news.Item{
			Category:    category,
			Provider:    news.ProviderSky,
			Title:       i.Title,
			Link:        i.Link,
			Description: i.Description,
			Thumbnail:   i.Thumbnail.URL,
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
