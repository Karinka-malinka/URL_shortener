package urlmemstore

import (
	"context"
	"database/sql"
	"sync"

	"github.com/URL_shortener/internal/app/urlapp"
)

var _ urlapp.URLStore = &URLs{}

type URLs struct {
	sync.Mutex
	m map[string]urlapp.URL
}

func NewURLs() *URLs {
	return &URLs{
		m: make(map[string]urlapp.URL),
	}
}

func (adr *URLs) Shortening(ctx context.Context, u []urlapp.URL) error {
	adr.Lock()
	defer adr.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, uu := range u {
		adr.m[uu.Short] = uu
	}
	return nil
}

func (adr *URLs) Resolve(ctx context.Context, shortURL string) (*urlapp.URL, error) {
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
