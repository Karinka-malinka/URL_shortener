package urlfilestore

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"

	"github.com/URL_shortener/internal/app/urlapp"
)

var _ urlapp.URLStore = &fileURLs{}

type fileURLs struct {
	sync.Mutex
	file *os.File
	m    map[string]urlapp.URL
}

func NewFileURLs(filename string) (*fileURLs, error) {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

	if err != nil {
		return nil, err
	}

	m := make(map[string]urlapp.URL)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		var URLData urlapp.URL
		err := json.Unmarshal(scanner.Bytes(), &URLData)

		if err != nil {
			return nil, err
		}

		m[URLData.Short] = URLData
	}

	f := fileURLs{file: file, m: m}

	return &f, nil
}

func (f *fileURLs) Shortening(ctx context.Context, u []urlapp.URL) error {

	f.Lock()
	defer f.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, uu := range u {
		data, err := json.Marshal(&uu)
		if err != nil {
			return err
		}

		data = append(data, '\n')

		f.m[uu.Short] = uu

		_, err = f.file.Write(data)
		if err != nil {
			return err
		}
	}

	return nil

}

func (f *fileURLs) Resolve(ctx context.Context, shortURL string) (*urlapp.URL, error) {

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
