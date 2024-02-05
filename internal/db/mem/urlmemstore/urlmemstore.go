package urlmemstore

import (
	"context"
	"database/sql"
	"sync"

	"github.com/URL_shortener/internal/app/url"
)

var _ url.URLStore = &URLs{}

type URLs struct {
	sync.Mutex
	m map[string]url.URL
}

func NewURLs() *URLs {
	return &URLs{
		m: make(map[string]url.URL),
	}
}

func (adr *URLs) Close() error {
	return nil
}

func (adr *URLs) Shortening(ctx context.Context, u url.URL) error {
	adr.Lock()
	defer adr.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	adr.m[u.Short] = u
	return nil
}

func (adr *URLs) Resolve(ctx context.Context, shortURL string) (*url.URL, error) {
	adr.Lock()
	defer adr.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	u, ok := adr.m[shortURL]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

func (adr *URLs) Ping() bool {
	return true
}
