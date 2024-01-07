package starter

import (
	"context"
	"sync"

	"github.com/URL_shortener/internal/app/url"
)

type App struct {
	urls *url.URLs
}

func NewApp(urlst url.URLStore) *App {
	a := &App{
		urls: url.NewURLs(urlst),
	}
	return a
}

type APIServer interface {
	Start(urls *url.URLs)
	Stop()
}

func (a *App) Serve(ctx context.Context, wg *sync.WaitGroup, hs APIServer) {
	defer wg.Done()
	hs.Start(a.urls)
	<-ctx.Done()
	hs.Stop()
}
