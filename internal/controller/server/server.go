package server

import (
	"context"
	"net/http"
	"time"

	"github.com/URL_shortener/internal/app/starter"
	"github.com/URL_shortener/internal/logger"
	"github.com/URL_shortener/internal/service/urlservice"
)

var _ starter.APIServer = &Server{}

type Server struct {
	srv  http.Server
	urls *urlservice.URLServices
}

func NewServer(addr string, h http.Handler) *Server {

	s := &Server{}

	s.srv = http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	s.srv.Shutdown(ctx)
	logger.Log.Infof("server stoped: %s", s.srv.Addr)
	cancel()
}

func (s *Server) Start(urls *urlservice.URLServices) {
	s.urls = urls
	//fmt.Println("server started: ", s.srv.Addr)
	logger.Log.Infof("server started: %s", s.srv.Addr)
	go s.srv.ListenAndServe()
}
