package userapp

import (
	"context"
	"fmt"

	"github.com/URL_shortener/cmd/config"
	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/google/uuid"
)

type User struct {
	UUID  uuid.UUID
	Token string
}

type UserStore interface {
	Create(ctx context.Context, user User) error
	Read(ctx context.Context, userID string) (*User, error)
	GetUserURLs(ctx context.Context, userID string) ([]urlapp.URL, error)
}

type Users struct {
	userStore UserStore
}

func NewUser(userStore UserStore) *Users {
	return &Users{
		userStore: userStore,
	}
}

func (ua *Users) Register(ctx context.Context, cfg config.ConfigData) (*User, error) {

	var nuser User

	nuser.UUID = uuid.New()

	if err := ua.userStore.Create(ctx, nuser); err != nil {
		return nil, err
	}

	token, err := ua.newToken(cfg, nuser)
	if err != nil {
		return nil, err
	}

	nuser.Token = token

	return &nuser, nil
}

func (ua *Users) UpdateToken(ctx context.Context, accessToken string, cfg config.ConfigData) (*User, error) {

	valid, userClaims, err := ua.parseToken(accessToken, cfg.SecretKeyForAccessToken)

	if err != nil {
		return nil, err
	}

	user, err := ua.userStore.Read(ctx, userClaims.UserID.String())
	if err != nil {
		return nil, err
	}

	user.Token = accessToken

	if !valid {
		token, err := ua.newToken(cfg, User{
			UUID: userClaims.UserID,
		})
		if err != err {
			return nil, err
		}

		user.Token = token
	}

	return user, nil
}

func (ua *Users) GetUserURLs(ctx context.Context, userID string) ([]urlapp.URL, error) {

	sURL, err := ua.userStore.GetUserURLs(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("read long url: %w", err)
	}

	return sURL, nil
}
