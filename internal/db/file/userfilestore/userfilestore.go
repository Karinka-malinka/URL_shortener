package userfilestore

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/app/userapp"
)

var _ userapp.UserStore = &UserStore{}

type UserStore struct {
	sync.Mutex
	file *os.File
	m    map[string]userapp.User
}

func NewFileUsers(filename string) (*UserStore, error) {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

	if err != nil {
		return nil, err
	}

	m := make(map[string]userapp.User)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		var userData userapp.User
		err := json.Unmarshal(scanner.Bytes(), &userData)

		if err != nil {
			return nil, err
		}

		m[userData.UUID.String()] = userData
	}

	f := UserStore{file: file, m: m}

	return &f, nil
}

func (d *UserStore) Create(ctx context.Context, user userapp.User) error {

	d.Lock()
	defer d.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	data, err := json.Marshal(&user)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	d.m[user.UUID.String()] = user

	_, err = d.file.Write(data)
	if err != nil {
		return err
	}

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

func (d *UserStore) DeleteUserURLs(ctx context.Context, shotrURLs []string, userID string) error {
	return fmt.Errorf("method is not available")
}
