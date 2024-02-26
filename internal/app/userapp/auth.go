package userapp

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTCustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func (ua *Users) newToken(cfg config.ConfigData, user User) (string, error) {

	validityPeriodAccessToken, err := strconv.Atoi(cfg.ValidityPeriodAccessToken)
	if err != nil {
		return "", err
	}

	accessToken := ua.getTokensWithClaims(validityPeriodAccessToken, user)

	accessTokenString, err := accessToken.SignedString([]byte(cfg.SecretKeyForAccessToken))
	if err != nil {
		return "", err
	}

	return accessTokenString, nil
}

func (ua *Users) getTokensWithClaims(ValidityPeriod int, user User) (accessToken *jwt.Token) {

	accessTokenClaims := &JWTCustomClaims{
		UserID: user.UUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(ValidityPeriod))),
		},
	}

	accessToken = jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	return accessToken
}

func (ua *Users) parseToken(accessToken, secretKey string) (bool, *JWTCustomClaims, error) {

	token, err := jwt.ParseWithClaims(accessToken, &JWTCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil

	})

	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			logger.Log.Infof("error in parsing token. error: %v", err)
			return false, nil, err
		}
	}

	userClaims := token.Claims.(*JWTCustomClaims)

	return token.Valid, userClaims, nil
}
