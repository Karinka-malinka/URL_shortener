package handler

import (
	"io"
	"net/http"

	"github.com/URL_shortener/internal/app/url"
	"github.com/labstack/echo/v4"
)

type Router struct {
	*echo.Echo
	urls *url.URLs
}

func NewRouter(urls *url.URLs) *Router {

	e := echo.New()

	r := &Router{
		Echo: e,
		urls: urls,
	}
	e.POST("/", r.ShortURL)
	e.GET("/:id", r.ResolveURL)

	return r
}

func (rt *Router) ShortURL(c echo.Context) error {

	rBody := c.Request().Body

	defer c.Request().Body.Close()

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

	urlShort := "http://" + c.Request().Host + "/" + nburl.Short

	c.String(http.StatusCreated, urlShort)
	return nil
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
