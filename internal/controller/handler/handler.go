package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/URL_shortener/internal/app/url"
)

type Router struct {
	*http.ServeMux
	urls *url.URLs
}

func NewRouter(urls *url.URLs) *Router {
	r := &Router{
		ServeMux: http.NewServeMux(),
		urls:     urls,
	}
	r.HandleFunc("/", r.ShortResolveURL)

	return r
}

func (rt *Router) ShortResolveURL(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:

		defer r.Body.Close()

		if r.Body == http.NoBody {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		burl := url.URL{
			Long: string(body),
		}

		nburl, err := rt.urls.Shortening(r.Context(), burl)
		if err != nil {
			http.Error(w, "error url shortening", http.StatusBadRequest)
		}
		urlShort := "http://" + r.Host + "/" + nburl.Short

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(urlShort))

	case http.MethodGet:

		uri := strings.Split(r.RequestURI, "/")
		if len(uri) < 2 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		longURL, err := rt.urls.Resolve(r.Context(), uri[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)

	default:
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
}
