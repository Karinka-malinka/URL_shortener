package urlservice

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/URL_shortener/internal/db/file/urlfilestore"
)

type URLServices struct {
	ustore urlfilestore.URLStore
}

func NewURLService(ustore urlfilestore.URLStore) *URLServices {
	return &URLServices{
		ustore: ustore,
	}
}

func (u *URLServices) Shortening(ctx context.Context, longURL string) (string, error) {

	shortURL := generateShortURL()

	err := u.ustore.Shortening(ctx, shortURL, longURL)
	if err != nil {
		return "", fmt.Errorf("create short url: %w", err)
	}

	return shortURL, nil
}

func (u *URLServices) Resolve(ctx context.Context, shortURL string) (string, error) {

	strURL, err := u.ustore.Resolve(ctx, shortURL)

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
