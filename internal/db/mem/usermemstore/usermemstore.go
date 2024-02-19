package usermemstore

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/app/userapp"
)

var _ userapp.UserStore = &UserStore{}

type UserStore struct {
	sync.Mutex
	m map[string]userapp.User
}

func NewUserStore() *UserStore {
	return &UserStore{
		m: make(map[string]userapp.User),
	}
}

func (d *UserStore) Create(ctx context.Context, user userapp.User) error {

	d.Lock()
	defer d.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	d.m[user.UUID.String()] = user

	return nil
}

func (d *UserStore) Read(ctx context.Context, userID string) (*userapp.User, error) {

	d.Lock()
	defer d.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	u, ok := d.m[userID]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

func (d *UserStore) GetUserURLs(ctx context.Context, userID string) ([]urlapp.URL, error) {
	return nil, fmt.Errorf("method is not available")
}
