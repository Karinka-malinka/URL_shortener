package urlhandler

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	mathrand "math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/db/file/urlfilestore"
	"github.com/URL_shortener/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// rnd generates new random generator with new source for each binary call
var rnd = func() *mathrand.Rand {
	buf := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(err)
	}
	src := mathrand.NewSource(int64(binary.LittleEndian.Uint64(buf)))
	return mathrand.New(src)
}()

// ASCIIString generates random ASCII string
func ASCIIString(minLen, maxLen int) string {
	var letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFJHIJKLMNOPQRSTUVWXYZ"

	slen := rnd.Intn(maxLen-minLen) + minLen

	s := make([]byte, 0, slen)
	i := 0
	for len(s) < slen {
		idx := rnd.Intn(len(letters) - 1)
		char := letters[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		s = append(s, char)
		i++
	}

	return string(s)
}

// generateTestURL возвращает валидный псевдослучайный URL
func generateTestURL(t *testing.T) string {
	t.Helper()

	var res url.URL

	// generate SCHEME
	res.Scheme = "http"
	res.Host = Domain(5, 15)

	for i := 0; i < rnd.Intn(4); i++ {
		res.Path += "/" + strings.ToLower(ASCIIString(5, 15))
	}

	return res.String()
}

func Domain(minLen, maxLen int, zones ...string) string {
	if minLen == 0 {
		minLen = 5
	}
	if maxLen == 0 {
		maxLen = 15
	}

	// generate ZONE
	var zone string
	switch len(zones) {
	case 1:
		zone = zones[0]
	case 0:
		zones = []string{"com", "ru", "net", "biz", "yandex"}
		zone = zones[rnd.Intn(len(zones))]
	default:
		zone = zones[rnd.Intn(len(zones))]
	}

	// generate HOST
	host := strings.ToLower(ASCIIString(minLen, maxLen))
	return host + "." + strings.TrimLeft(zone, ".")
}

func TestURLHandler_ShortURL_ResolveURL_file(t *testing.T) {

	cfg := config.NewConfig()

	originalURL := generateTestURL(t)
	var shortenURL string

	// Setup
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set(echo.HeaderContentType, echo.MIMETextHTML)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	urlst, err := urlfilestore.NewFileURLs("/tmp/shorturldb.json")
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	urlApp := urlapp.NewURLs(urlst)

	//cfg.BaseShortAddr = ""

	h := NewURLHandler(urlApp, cfg)

	// Assertions
	if assert.NoError(t, h.ShortURL(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	shortenURL = rec.Body.String()

	_, urlParseErr := url.Parse(shortenURL)

	if assert.NoErrorf(t, urlParseErr, "Невозможно распарсить полученный сокращенный URL - %s : %s", shortenURL, err) {

		req = httptest.NewRequest(http.MethodGet, "/", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		sh := strings.Split(shortenURL, "/")
		c.SetParamValues(sh[1])

		// Assertions
		if assert.NoError(t, h.ResolveURL(c)) {
			assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
			assert.Equal(t, originalURL, rec.Header().Get("Location"))
		}
	}
}
