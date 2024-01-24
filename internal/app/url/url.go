package url

import (
	"context"
	"fmt"
	"math/rand"
)

type URL struct {
	UUID  string `json:"uuid"`
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}

// инверсия зависимостей к базе данных
type URLStore interface {
	Shortening(ctx context.Context, adr URL) error
	Resolve(ctx context.Context, shortURL string) (string, error)
	CurrentUUID() string
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

	adr.UUID = u.adrstore.CurrentUUID()
	adr.Short = generateShortURL()

	err := u.adrstore.Shortening(ctx, adr)
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

func generateShortURL() string {

	const shortURLLength = 8

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	shortURL := make([]byte, shortURLLength)

	for i := range shortURL {
		shortURL[i] = letters[rand.Intn(len(letters))]
	}

	return string(shortURL)

}
