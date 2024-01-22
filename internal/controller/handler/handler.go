package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/url"
	"github.com/URL_shortener/internal/logger"
	"github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

type Router struct {
	*echo.Echo
	urls *url.URLs
	cfg  *config.ConfigData
}

func NewRouter(urls *url.URLs, cfg *config.ConfigData) *Router {

	e := echo.New()

	r := &Router{
		Echo: e,
		urls: urls,
		cfg:  cfg,
	}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogStatus:       true,
		LogMethod:       true,
		LogResponseSize: true,
		LogLatency:      true,
		LogValuesFunc:   logger.RequestLogger,
	}))

	e.Use(middleware.Decompress())
	/*
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: func(c echo.Context) bool {
				contentType := [...]string{
					"application/json",
					"text/html",
				}
				ch := c.Response().Header().Get("Content-Type")
				for _, ct := range contentType {
					if strings.Contains(ch, ct) {
						return false
					}
				}
				return true
			},
		}))
	*/
	e.Use(middleware.Gzip())

	e.POST("/", r.ShortURL)
	e.GET("/:id", r.ResolveURL)
	e.POST("/api/shorten", r.ShortURLJSON)

	return r
}

func (rt *Router) ShortURL(c echo.Context) error {

	ca := make(chan string, 1)
	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	go func() error {

		if rBody == http.NoBody {
			return echo.ErrBadRequest
		}

		body, err := io.ReadAll(rBody)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		burl := url.URL{
			Long: string(body),
		}

		nburl, err := rt.urls.Shortening(c.Request().Context(), burl)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		urlShort := rt.cfg.BaseShortAddr + "/" + nburl.Short

		ca <- urlShort
		return nil
	}()

	select {
	case result := <-ca:
		return c.String(http.StatusCreated, result)
	case <-c.Request().Context().Done():
		return nil
	}
}

func (rt *Router) ShortURLJSON(c echo.Context) error {

	ca := make(chan string, 1)
	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	if rBody == http.NoBody {
		return echo.NewHTTPError(http.StatusBadRequest, "No body")
	}

	go func() error {

		body, err := io.ReadAll(rBody)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		var inputData map[string]string

		if err = json.Unmarshal(body, &inputData); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		original_URL := inputData["url"]
		if original_URL == "" {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		burl := url.URL{
			Long: original_URL,
		}

		nburl, err := rt.urls.Shortening(c.Request().Context(), burl)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		urlShort := rt.cfg.BaseShortAddr + "/" + nburl.Short

		ca <- urlShort
		return nil
	}()

	select {
	case result := <-ca:

		data := map[string]interface{}{
			"result": result,
		}
		//c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		return c.JSON(http.StatusCreated, data)
	case <-c.Request().Context().Done():
		return nil
	}
}

func (rt *Router) ResolveURL(c echo.Context) error {

	//uri := strings.Split(r.RequestURI, "/")
	uri := c.Param("id")

	longURL, err := rt.urls.Resolve(c.Request().Context(), uri)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	c.Redirect(http.StatusTemporaryRedirect, longURL)
	return nil
}
