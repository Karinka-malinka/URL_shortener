package urlfilestore

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var _ URLStore = &fileURLs{}

type URL struct {
	UUID  string `json:"uuid"`
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}

type URLStore interface {
	Shortening(ctx context.Context, shortURL, longURL string) error
	Resolve(ctx context.Context, shortURL string) (*URL, error)
}

type fileURLs struct {
	sync.Mutex
	file        *os.File
	m           map[string]URL
	currentUUID uint32
}

func NewFileURLs(filename string) (*fileURLs, error) {

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)

	if err != nil {
		return nil, err
	}

	m := make(map[string]URL)
	var curUUID uint32

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		var URLData URL
		err := json.Unmarshal(scanner.Bytes(), &URLData)

		if err != nil {
			return nil, err
		}

		m[URLData.Short] = URLData
		curUUID++
	}

	f := fileURLs{file: file, m: m, currentUUID: curUUID}

	return &f, nil
}

func (f *fileURLs) Close() error {
	return f.file.Close()
}

func (f *fileURLs) Shortening(ctx context.Context, shortURL, longURL string) error {

	f.Lock()
	defer f.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	f.currentUUID++
	u := URL{
		UUID:  fmt.Sprintf("%d", f.currentUUID),
		Short: shortURL,
		Long:  longURL}

	data, err := json.Marshal(&u)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	f.m[u.Short] = u

	_, err = f.file.Write(data)
	return err

}

func (f *fileURLs) Resolve(ctx context.Context, shortURL string) (*URL, error) {

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
