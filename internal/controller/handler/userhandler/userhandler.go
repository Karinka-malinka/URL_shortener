package userhandler

import (
	"net/http"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/app/userapp"
	"github.com/URL_shortener/internal/controller/handler"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	UserApp *userapp.Users
	cfg     *config.ConfigData
}

func NewUserHandler(userapp *userapp.Users, cfg *config.ConfigData) *UserHandler {
	return &UserHandler{UserApp: userapp, cfg: cfg}
}

func (lh *UserHandler) RegisterHandler(e *echo.Echo, apiGroup *echo.Group) {

	userGroup := apiGroup.Group("/user")
	userGroup.GET("/urls", lh.UserURLs)

}

func (lh *UserHandler) UserURLs(c echo.Context) error {

	userID, err := handler.GetUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	ca := make(chan []urlapp.URL, 1)
	errc := make(chan error)

	go func() error {

		sURL, err := lh.UserApp.GetUserURLs(c.Request().Context(), userID.String())
		if err != nil {
			errc <- err
			return err
		}

		var outputData []urlapp.URL
		for _, u := range sURL {

			outputData = append(outputData, urlapp.URL{
				Short: lh.cfg.BaseShortAddr + "/" + u.Short,
				Long:  u.Long})
		}

		ca <- outputData
		return nil

	}()

	select {
	case result := <-ca:
		if len(result) == 0 {
			return echo.NewHTTPError(http.StatusNoContent)
		}
		return c.JSON(http.StatusOK, result)
	case err := <-errc:
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case <-c.Request().Context().Done():
		return nil
	}

}
