package url

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/URL_shortener/internal/db/file/urlfilestore"
)

// инверсия зависимостей к базе данных
type URLStore interface {
	Shortening(ctx context.Context, shortURL, longURL string) error
	Resolve(ctx context.Context, shortURL string) (*urlfilestore.URL, error)
}

type URLs struct {
	adrstore URLStore
}

func NewURLs(adrstore URLStore) *URLs {
	return &URLs{
		adrstore: adrstore,
	}
}

func (u *URLs) Shortening(ctx context.Context, longURL string) (string, error) {

	shortURL := generateShortURL()

	err := u.adrstore.Shortening(ctx, shortURL, longURL)
	if err != nil {
		return "", fmt.Errorf("create short url: %w", err)
	}

	return shortURL, nil
}

func (u *URLs) Resolve(ctx context.Context, shortURL string) (string, error) {

	strURL, err := u.adrstore.Resolve(ctx, shortURL)

	if err != nil {
		return "", fmt.Errorf("read long url: %w", err)
	}

	return strURL.Long, nil
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
