package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/starter"
	"github.com/URL_shortener/internal/controller/handler"
	"github.com/URL_shortener/internal/controller/server"
	"github.com/URL_shortener/internal/db/file/urlfilestore"
	"github.com/URL_shortener/internal/logger"
	"github.com/URL_shortener/internal/service/urlservice"
)

func main() {

	logger.Initialize()

	cfg := config.NewConfig()

	parseFlags(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	urlst, err := urlfilestore.NewFileURLs(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	a := starter.NewApp(urlst)
	urls := urlservice.NewURLService(urlst)
	h := handler.NewRouter(urls, cfg)
	srv := server.NewServer(cfg.RunAddr, h)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go a.Serve(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
