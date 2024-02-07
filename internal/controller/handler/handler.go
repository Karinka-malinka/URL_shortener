package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/url"
	"github.com/URL_shortener/internal/db/base/urldbstore"
	"github.com/URL_shortener/internal/logger"
	"github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

type Router struct {
	*echo.Echo
	urls *url.URLs
	cfg  *config.ConfigData
}

/*type ErrSt struct {
	Err         error
	Description string
}*/

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
		LogError:        true,
	}))

	e.Use(middleware.Decompress())
	e.Use(middleware.Gzip())

	e.POST("/", r.ShortURL)
	e.GET("/:id", r.ResolveURL)
	e.POST("/api/shorten", r.ShortURLJSON)
	e.POST("/api/shorten/batch", r.Batch)
	e.GET("/ping", r.Ping)

	return r
}

func (rt *Router) ShortURL(c echo.Context) error {

	ca := make(chan string, 1)
	errc := make(chan error)

	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	go func() error {

		if rBody == http.NoBody {
			err := fmt.Errorf("%s", "No body")
			errc <- err
			return err
		}

		body, err := io.ReadAll(rBody)

		if err != nil {
			errc <- err
			return err
		}

		shortURL, err := rt.urls.Shortening(c.Request().Context(), string(body))

		if err != nil {
			errc <- err
			return err
		}

		urlShort := rt.cfg.BaseShortAddr + "/" + shortURL

		ca <- urlShort
		return nil
	}()

	select {
	case result := <-ca:
		return c.String(http.StatusCreated, result)
	case err := <-errc:
		var errConflict *urldbstore.ErrConflict
		if errors.As(err, &errConflict) {
			return c.String(http.StatusConflict, rt.cfg.BaseShortAddr+"/"+errConflict.URL.Short)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

func (rt *Router) ShortURLJSON(c echo.Context) error {

	ca := make(chan string, 1)
	errc := make(chan error)

	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	if rBody == http.NoBody {
		return echo.NewHTTPError(http.StatusBadRequest, "No body")
	}

	go func() error {

		body, err := io.ReadAll(rBody)

		if err != nil {
			errc <- err
			return err
		}

		var inputData map[string]string

		if err = json.Unmarshal(body, &inputData); err != nil {
			errc <- err
			return err
		}

		originalURL := inputData["url"]
		if originalURL != "" {

			shortURL, err := rt.urls.Shortening(c.Request().Context(), originalURL)

			if err != nil {
				errc <- err
				return err
			}

			urlShort := rt.cfg.BaseShortAddr + "/" + shortURL
			ca <- urlShort
			return nil
		}

		err = fmt.Errorf("%s", "No url")
		errc <- err
		return err

	}()

	select {
	case result := <-ca:
		data := map[string]interface{}{
			"result": result,
		}
		return c.JSON(http.StatusCreated, data)
	case err := <-errc:
		var errConflict *urldbstore.ErrConflict
		if errors.As(err, &errConflict) {
			data := map[string]interface{}{
				"result": rt.cfg.BaseShortAddr + "/" + errConflict.URL.Short,
			}
			return c.JSON(http.StatusConflict, data)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

func (rt *Router) Batch(c echo.Context) error {

	ca := make(chan []url.URL, 1)
	errc := make(chan error)

	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	if rBody == http.NoBody {
		return echo.NewHTTPError(http.StatusBadRequest, "No body")
	}

	go func() error {

		body, err := io.ReadAll(rBody)

		if err != nil {
			errc <- err
			return err
		}

		var inputData []url.URL

		if err = json.Unmarshal(body, &inputData); err != nil {
			errc <- err
			return err
		}

		shortURL, err := rt.urls.Batch(c.Request().Context(), inputData)
		if err != nil {
			errc <- err
			return err
		}

		var outputData []url.URL
		for _, u := range *shortURL {

			outputData = append(outputData, url.URL{
				Short:         rt.cfg.BaseShortAddr + "/" + u.Short,
				CorrelationID: u.CorrelationID})
		}

		ca <- outputData
		return nil

	}()

	select {
	case result := <-ca:
		return c.JSON(http.StatusCreated, result)
	case err := <-errc:
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

func (rt *Router) ResolveURL(c echo.Context) error {

	uri := c.Param("id")

	originalURL, err := rt.urls.Resolve(c.Request().Context(), uri)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
	return nil
}

func (rt *Router) Ping(c echo.Context) error {

	if !rt.urls.PingDB() {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}
