package url

import (
	"context"
	"fmt"
	"time"

	hashids "github.com/speps/go-hashids"
)

type URL struct {
	Short string `json:"short_url"`
	Long  string `json:"url"`
}

// инверсия зависимостей к базе данных
type URLStore interface {
	Shortening(ctx context.Context, adr URL) error
	Resolve(ctx context.Context, shortURL string) (string, error)
}

type URLs struct {
	adrstore URLStore
}

func NewURLs(adrstore URLStore) *URLs {
	return &URLs{
		adrstore: adrstore,
	}
}

func (u *URLs) Shortening(ctx context.Context, adr URL) (*URL, error) {

	// получаем короткий url как хэш текущего времени
	hd := hashids.NewData()
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	urlID, err := h.Encode([]int{int(now.UnixNano())})
	if err != nil {
		return nil, err
	}

	adr.Short = urlID

	err = u.adrstore.Shortening(ctx, adr)
	if err != nil {
		return nil, fmt.Errorf("create short url: %w", err)
	}

	return &adr, nil
}

func (u *URLs) Resolve(ctx context.Context, shortURL string) (string, error) {
	longURL, err := u.adrstore.Resolve(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("read long url: %w", err)
	}
	if longURL == "" {
		return "", fmt.Errorf("empty long url: %w", err)
	}
	return longURL, nil
}
