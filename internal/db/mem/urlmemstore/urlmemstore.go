package urlmemstore

import (
	"context"
	"database/sql"
	"sync"
)

type URL struct {
	UUID  string `json:"uuid"`
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}

type URLs struct {
	sync.Mutex
	m map[string]URL
}

func NewURLs() *URLs {
	return &URLs{
		m: make(map[string]URL),
	}
}

func (adr *URLs) Shortening(ctx context.Context, u URL) error {
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

func (adr *URLs) Resolve(ctx context.Context, shortURL string) (*URL, error) {
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
