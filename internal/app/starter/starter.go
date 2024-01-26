package starter

import (
	"context"
	"sync"

	"github.com/URL_shortener/internal/db/file/urlfilestore"
	"github.com/URL_shortener/internal/service/urlservice"
)

type App struct {
	urls *urlservice.URLServices
}

func NewApp(urlst urlfilestore.URLStore) *App {
	a := &App{
		urls: urlservice.NewURLService(urlst),
	}
	return a
}

type APIServer interface {
	Start(urls *urlservice.URLServices)
	Stop()
}

func (a *App) Serve(ctx context.Context, wg *sync.WaitGroup, hs APIServer) {
	defer wg.Done()
	hs.Start(a.urls)
	<-ctx.Done()
	hs.Stop()
}
