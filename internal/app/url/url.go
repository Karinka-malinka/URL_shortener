package url

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

type URL struct {
	UUID          uuid.UUID `json:"uuid"`
	Short         string    `json:"short_url"`
	Long          string    `json:"original_url"`
	CorrelationID string    `json:"-"`
}

// инверсия зависимостей к базе данных
type URLStore interface {
	Shortening(ctx context.Context, u []URL) error
	Resolve(ctx context.Context, shortURL string) (*URL, error)
	Ping() bool
	Close() error
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

	var nu []URL

	nu = append(nu, URL{
		UUID:  uuid.New(),
		Short: shortURL,
		Long:  longURL,
	})

	err := u.adrstore.Shortening(ctx, nu)

	if err != nil {
		return shortURL, fmt.Errorf("create short url: %w", err)
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

func (u *URLs) PingDB() bool {
	return u.adrstore.Ping()
}

func (u *URLs) CloseDB() {
	u.adrstore.Close()
}

func (u *URLs) Batch(ctx context.Context, sURL []URL) (*[]URL, error) {

	var nu []URL

	for _, bu := range sURL {

		shortURL := generateShortURL()

		nu = append(nu, URL{
			UUID:          uuid.New(),
			Short:         shortURL,
			Long:          bu.Long,
			CorrelationID: bu.CorrelationID,
		})
	}

	err := u.adrstore.Shortening(ctx, nu)

	if err != nil {
		return &nu, fmt.Errorf("create short url: %w", err)
	}

	return &nu, nil
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
