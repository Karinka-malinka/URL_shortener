package urlfilestore

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"

	"github.com/URL_shortener/internal/app/url"
)

var _ url.URLStore = &fileURLs{}

type fileURLs struct {
	sync.Mutex
	file *os.File
	m    map[string]string
}

func NewFileURLs(filename string) (*fileURLs, error) {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		var URLData url.URL
		err := json.Unmarshal(scanner.Bytes(), &URLData)

		if err != nil {
			return nil, err
		}

		m[URLData.Short] = URLData.Long
	}

	return &fileURLs{file: file, m: m}, nil
}

func (f *fileURLs) Close() error {
	return f.file.Close()
}

func (f *fileURLs) Shortening(ctx context.Context, u url.URL) error {
	f.Lock()
	defer f.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	data, err := json.Marshal(&u)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	f.m[u.Short] = u.Long

	_, err = f.file.Write(data)
	return err

}

func (f *fileURLs) Resolve(ctx context.Context, shortURL string) (string, error) {

	f.Lock()
	defer f.Unlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	u, ok := f.m[shortURL]
	if ok {
		return u, nil
	}
	return "", sql.ErrNoRows
}

func (f *fileURLs) CurrentUUID() int {
	return len(f.m)
}
