package handler

import (
	"fmt"

	"github.com/URL_shortener/internal/app/userapp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegisterHandler(*echo.Echo, *echo.Group)
}

func GetUserID(c echo.Context) (uuid.UUID, error) {

	var userID uuid.UUID

	u := c.Get("user")
	if u != nil {
		u := u.(*jwt.Token)
		claims := u.Claims.(*userapp.JWTCustomClaims)
		userID = claims.UserID
		return userID, nil
	} else {
		u := c.Get("userID").(uuid.UUID)
		if u != uuid.Nil {
			return u, nil
		}
	}

	return uuid.Nil, fmt.Errorf("no token")
}
