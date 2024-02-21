package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/app/userapp"
	"github.com/URL_shortener/internal/controller/handler"
	"github.com/URL_shortener/internal/controller/handler/urlhandler"
	"github.com/URL_shortener/internal/controller/handler/userhandler"
	"github.com/URL_shortener/internal/controller/router"
	"github.com/URL_shortener/internal/controller/server"
	"github.com/URL_shortener/internal/db/base"
	"github.com/URL_shortener/internal/db/base/urldbstore"
	"github.com/URL_shortener/internal/db/base/userdbstore"
	"github.com/URL_shortener/internal/db/file/urlfilestore"
	"github.com/URL_shortener/internal/db/file/userfilestore"
	"github.com/URL_shortener/internal/db/mem/urlmemstore"
	"github.com/URL_shortener/internal/db/mem/usermemstore"
	"github.com/URL_shortener/internal/logger"
)

func main() {

	logger.Initialize()

	cfg := config.NewConfig()

	parseFlags(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	var PostgresDatabase *base.PostgresDatabase
	var urlst urlapp.URLStore
	var userst userapp.UserStore
	var err error

	var registeredHandlers []handler.Handler

	urlst = urlmemstore.NewURLs()
	userst = usermemstore.NewUserStore()

	if cfg.DatabaseDSN != "" {

		PostgresDatabase, err = base.NewDB(ctx, cfg.DatabaseDSN)

		if err != nil {
			logger.Log.Fatalf("error in open database. error: %v", err)
		}

		urlst = urldbstore.NewURLStore(PostgresDatabase.DB)
		userst = userdbstore.NewUserStore(PostgresDatabase.DB)

	} else if cfg.FileStoragePath != "" {
		urlst, err = urlfilestore.NewFileURLs(cfg.FileStoragePath)
		if err != nil {
			logger.Log.Fatal(err.Error())
		}

		userst, err = userfilestore.NewFileUsers("/tmp/user.json")
		if err != nil {
			logger.Log.Fatal(err.Error())
		}
	}

	urlApp := urlapp.NewURLs(urlst)
	urlHandler := urlhandler.NewURLHandler(urlApp, cfg)
	registeredHandlers = append(registeredHandlers, urlHandler)

	userApp := userapp.NewUser(userst)
	userHandler := userhandler.NewUserHandler(userApp, cfg)
	registeredHandlers = append(registeredHandlers, userHandler)

	appRouter := router.NewRouter(*cfg, registeredHandlers, userApp)
	srv := server.NewServer(cfg.RunAddr, appRouter.Echo)

	go srv.Start()

	<-ctx.Done()
	srv.Stop()
	cancel()

}
