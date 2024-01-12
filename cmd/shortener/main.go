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
	"github.com/URL_shortener/internal/db/mem/urlmemstore"
)

func main() {

	cfg := config.NewConfig()

	parseFlags(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	urlst := urlmemstore.NewURLs()
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
