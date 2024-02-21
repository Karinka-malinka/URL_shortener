package urlhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/controller/handler"
	"github.com/URL_shortener/internal/db/base/urldbstore"
	"github.com/labstack/echo/v4"
)

type URLHandler struct {
	URLApp *urlapp.URLs
	cfg    *config.ConfigData
}

func NewURLHandler(urlapp *urlapp.URLs, cfg *config.ConfigData) *URLHandler {
	return &URLHandler{URLApp: urlapp, cfg: cfg}
}

func (lh *URLHandler) RegisterHandler(e *echo.Echo, apiGroup *echo.Group) {

	e.POST("/", lh.ShortURL)
	e.GET("/:id", lh.ResolveURL)
	e.GET("/ping", lh.Ping)

	shortenGroup := apiGroup.Group("/shorten")
	shortenGroup.POST("", lh.ShortURLJSON)
	shortenGroup.POST("/batch", lh.Batch)

}

func (lh *URLHandler) ShortURL(c echo.Context) error {

	ca := make(chan string, 1)
	errc := make(chan error)

	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	userID, err := handler.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

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

		shortURL, err := lh.URLApp.Shortening(c.Request().Context(), string(body), userID)

		if err != nil {
			errc <- err
			return err
		}

		urlShort := lh.cfg.BaseShortAddr + "/" + shortURL

		ca <- urlShort
		return nil
	}()

	select {
	case result := <-ca:
		return c.String(http.StatusCreated, result)
	case err := <-errc:
		var errConflict *urldbstore.ErrConflict
		if errors.As(err, &errConflict) {
			return c.String(http.StatusConflict, lh.cfg.BaseShortAddr+"/"+errConflict.URL.Short)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

func (lh *URLHandler) ShortURLJSON(c echo.Context) error {

	ca := make(chan string, 1)
	errc := make(chan error)

	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	userID, err := handler.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

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

			shortURL, err := lh.URLApp.Shortening(c.Request().Context(), originalURL, userID)

			if err != nil {
				errc <- err
				return err
			}

			urlShort := lh.cfg.BaseShortAddr + "/" + shortURL
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
				"result": lh.cfg.BaseShortAddr + "/" + errConflict.URL.Short,
			}
			return c.JSON(http.StatusConflict, data)
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}
}

func (lh *URLHandler) Batch(c echo.Context) error {

	ca := make(chan []urlapp.URL, 1)
	errc := make(chan error)

	r := c.Request()

	rBody := r.Body

	defer rBody.Close()

	userID, err := handler.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	if rBody == http.NoBody {
		return echo.NewHTTPError(http.StatusBadRequest, "No body")
	}

	go func() error {

		body, err := io.ReadAll(rBody)

		if err != nil {
			errc <- err
			return err
		}

		var inputData []urlapp.URL

		if err = json.Unmarshal(body, &inputData); err != nil {
			errc <- err
			return err
		}

		shortURL, err := lh.URLApp.Batch(c.Request().Context(), inputData, userID)
		if err != nil {
			errc <- err
			return err
		}

		var outputData []urlapp.URL
		for _, u := range *shortURL {

			outputData = append(outputData, urlapp.URL{
				Short:         lh.cfg.BaseShortAddr + "/" + u.Short,
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

func (lh *URLHandler) ResolveURL(c echo.Context) error {

	uri := c.Param("id")

	URL, err := lh.URLApp.Resolve(c.Request().Context(), uri)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if URL.DeletedFlag {
		return c.JSON(http.StatusGone, "url delete")
	}

	c.Redirect(http.StatusTemporaryRedirect, URL.Long)
	return nil
}

func (lh *URLHandler) Ping(c echo.Context) error {

	if !lh.URLApp.PingDB() {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}
