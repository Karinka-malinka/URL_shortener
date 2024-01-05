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
	m map[string]string
}

func NewURLs() *URLs {
	return &URLs{
		m: make(map[string]string),
	}
}

func (adr *URLs) Shortening(ctx context.Context, u url.URL) error {
	adr.Lock()
	defer adr.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	adr.m[u.Short] = u.Long
	return nil
}

func (adr *URLs) Resolve(ctx context.Context, shortURL string) (string, error) {
	adr.Lock()
	defer adr.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	u, ok := adr.m[shortURL]
	if ok {
		return u, nil
	}
	return "", sql.ErrNoRows
}
