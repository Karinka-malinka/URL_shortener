package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/starter"
	"github.com/URL_shortener/internal/app/url"
	"github.com/URL_shortener/internal/controller/handler"
	"github.com/URL_shortener/internal/controller/server"
	"github.com/URL_shortener/internal/db/base/urldbstore"
	"github.com/URL_shortener/internal/db/file/urlfilestore"
	"github.com/URL_shortener/internal/db/mem/urlmemstore"
	"github.com/URL_shortener/internal/logger"
)

func main() {

	logger.Initialize()

	cfg := config.NewConfig()

	parseFlags(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	var urlst url.URLStore
	var err error

	urlst = urlmemstore.NewURLs()

	if cfg.DatabaseDSN != "" {
		urlst, err = urldbstore.NewDB(ctx, cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}
	} else if cfg.FileStoragePath != "" {
		urlst, err = urlfilestore.NewFileURLs(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}
	}

	a := starter.NewApp(urlst)
	urls := url.NewURLs(urlst)
	h := handler.NewRouter(urls, cfg)
	srv := server.NewServer(cfg.RunAddr, h)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go a.Serve(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
