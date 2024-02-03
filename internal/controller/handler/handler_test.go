package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/url"
	"github.com/URL_shortener/internal/db/file/urlfilestore"
	"github.com/URL_shortener/internal/db/mem/urlmemstore"
	"github.com/URL_shortener/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_ShortResolveURL2(t *testing.T) {

	cfg := config.NewConfig()
	//cfg.FileStoragePath = "C:/GoWork/src/github.com/URL_shortener/short-url-db.json"
	cfg.FileStoragePath = "/tmp/short-url-db.json"

	urlst, err := urlfilestore.NewFileURLs(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Fatal(err.Error() + ", path = " + cfg.FileStoragePath)
	}
	urls := url.NewURLs(urlst)
	rt := NewRouter(urls, cfg)
	hts := httptest.NewServer(rt)
	cli := hts.Client()

	cfg.BaseShortAddr = hts.URL
	requeststr := hts.URL + "/"

	type request struct {
		metod string
		body  *strings.Reader
	}

	tests := []struct {
		name           string
		request        request
		wantStatusCode int
	}{
		{
			name: "POST positive test",
			request: request{
				metod: "POST",
				body:  strings.NewReader(`https://practicum.yandex.ru/`),
			},
			wantStatusCode: 201,
		},
		{
			name: "POST negative test",
			request: request{
				metod: "POST",
				body:  strings.NewReader(""),
			},
			wantStatusCode: 405,
		},
		{
			name: "GET positive test",
			request: request{
				metod: "GET",
				body:  strings.NewReader(""),
			},
			wantStatusCode: 200,
		},
		{
			name: "Negative test",
			request: request{
				metod: "PUT",
				body:  strings.NewReader(""),
			},
			wantStatusCode: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, _ := http.NewRequest(tt.request.metod, requeststr, tt.request.body)

			resp, err := cli.Do(r)
			if err != nil {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if tt.request.metod == "POST" && resp.StatusCode == 201 {
				rbody, _ := io.ReadAll(resp.Body)
				require.NoError(t, err)
				err = resp.Body.Close()
				require.NoError(t, err)

				requeststr = string(rbody)
			}

		})
	}
}

func TestRouter_Ping(t *testing.T) {

	cfg := config.NewConfig()

	urlst := urlmemstore.NewURLs()
	urls := url.NewURLs(urlst)
	rt := NewRouter(urls, cfg)
	hts := httptest.NewServer(rt)
	cli := hts.Client()

	cfg.BaseShortAddr = hts.URL
	requeststr := hts.URL + "/ping"

	type request struct {
		metod string
	}

	tests := []struct {
		name           string
		request        request
		wantStatusCode int
	}{
		{
			name: "GET positive test",
			request: request{
				metod: "GET",
			},
			wantStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, _ := http.NewRequest(tt.request.metod, requeststr, nil)

			resp, err := cli.Do(r)
			if err != nil {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

		})
	}
}
