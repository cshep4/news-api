package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/cshep4/news-api/internal/news"
)

type cache struct {
	mutex sync.Mutex
	clock clockwork.Clock
	feeds map[string]news.Feed
}

func New(clock clockwork.Clock) (*cache, error) {
	if clock == nil {
		return nil, news.InvalidParameterError{Parameter: "clock"}
	}

	return &cache{
		clock: clock,
		mutex: sync.Mutex{},
		feeds: make(map[string]news.Feed),
	}, nil
}

func (c *cache) Get(provider news.Provider, category news.Category) (*news.Feed, bool) {
	hash := c.hash(provider, category)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	feed, ok := c.feeds[hash]
	if !ok {
		return nil, false
	}

	return &feed, true
}

func (c *cache) Store(provider news.Provider, category news.Category, feed news.Feed) {
	hash := c.hash(provider, category)

	c.put(hash, feed)

	go c.invalidateAfterTTL(hash, feed.TTL)
}

func (c *cache) hash(provider news.Provider, category news.Category) string {
	return fmt.Sprintf("%s-%s", provider, category)
}

func (c *cache) put(hash string, feed news.Feed) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.feeds[hash] = feed
}

func (c *cache) invalidateAfterTTL(hash string, ttl int) {
	select {
	case <-c.clock.After(time.Minute * time.Duration(ttl)):
		c.mutex.Lock()
		defer c.mutex.Unlock()

		delete(c.feeds, hash)
	}
}
