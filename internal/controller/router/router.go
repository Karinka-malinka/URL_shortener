package router

import (
	"errors"
	"net/http"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/userapp"
	"github.com/URL_shortener/internal/controller/handler"
	"github.com/URL_shortener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	middleware "github.com/labstack/echo/v4/middleware"
)

type Router struct {
	Echo    *echo.Echo
	UserAPP *userapp.Users
}

func NewRouter(cfg config.ConfigData, handlers []handler.Handler, userApp *userapp.Users) *Router {

	e := echo.New()

	r := &Router{
		Echo:    e,
		UserAPP: userApp,
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
	e.Use(middleware.Recover())
	e.Use(middleware.Decompress())
	e.Use(middleware.Gzip())

	restrictedConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &userapp.JWTCustomClaims{}
		},
		SigningKey: []byte(cfg.SecretKeyForAccessToken),
		ErrorHandler: func(c echo.Context, err error) error {
			return r.TokenRefresher(c, cfg)
		},
		ContinueOnIgnoredError: true,
		TokenLookup:            "header:Authorization:Bearer ,cookie:access_token",
	}

	e.Use(echojwt.WithConfig(restrictedConfig))

	apiGroup := e.Group("/api")

	for _, handler := range handlers {
		handler.RegisterHandler(e, apiGroup)
	}

	return r
}

func (rt *Router) TokenRefresher(c echo.Context, cfg config.ConfigData) error {

	if c.Path() == "/api/user/urls" {
		c.Response().Writer.WriteHeader(http.StatusUnauthorized)
	}

	cookie, err := c.Cookie("access_token")

	var user *userapp.User

	if err != nil {
		user, err = rt.UserAPP.Register(c.Request().Context(), cfg)
	} else {
		user, err = rt.UserAPP.UpdateToken(c.Request().Context(), cookie.Value, cfg)
		if err != nil {
			user, err = rt.UserAPP.Register(c.Request().Context(), cfg)
		}
	}

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			c.Response().Writer.WriteHeader(http.StatusUnauthorized)
		} else {
			c.Response().Writer.WriteHeader(http.StatusInternalServerError)
		}
	}

	sendResponceToken(c, user.Token)

	c.Set("userID", user.UUID)

	return nil
}

func sendResponceToken(c echo.Context, accessToken string) {

	c.Response().Header().Set("Authorization", "Bearer "+accessToken)

	writeAccessTokenCookie(c, accessToken)

}

func writeAccessTokenCookie(c echo.Context, accessToken string) {

	cookie := new(http.Cookie)

	cookie.Name = "access_token"
	cookie.Value = accessToken
	cookie.HttpOnly = true
	cookie.SameSite = 3
	cookie.Path = "/"

	c.SetCookie(cookie)
}
