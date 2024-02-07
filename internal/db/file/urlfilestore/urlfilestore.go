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
	m    map[string]url.URL
}

func NewFileURLs(filename string) (*fileURLs, error) {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

	if err != nil {
		return nil, err
	}

	m := make(map[string]url.URL)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		var URLData url.URL
		err := json.Unmarshal(scanner.Bytes(), &URLData)

		if err != nil {
			return nil, err
		}

		m[URLData.Short] = URLData
	}

	f := fileURLs{file: file, m: m}

	return &f, nil
}

func (f *fileURLs) Close() error {
	return f.file.Close()
}

func (f *fileURLs) Shortening(ctx context.Context, u []url.URL) (*url.URL, error) {

	f.Lock()
	defer f.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := json.Marshal(&u)
	if err != nil {
		return nil, err
	}

	data = append(data, '\n')

	for _, uu := range u {
		f.m[uu.Short] = uu
	}

	_, err = f.file.Write(data)
	return nil, err

}

func (f *fileURLs) Resolve(ctx context.Context, shortURL string) (*url.URL, error) {

	f.Lock()
	defer f.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	u, ok := f.m[shortURL]
	if ok {
		return &u, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fileURLs) Ping() bool {
	return true
}
