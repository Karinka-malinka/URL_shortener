package userdbstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/app/userapp"
	"github.com/URL_shortener/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var _ userapp.UserStore = &UserStore{}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {

	return &UserStore{db: db}
}

func (d *UserStore) Create(ctx context.Context, user userapp.User) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO users (uuid) VALUES($1)", user.UUID.String())

	if err != nil {

		return err
	}

	return tx.Commit()
}

func (d *UserStore) Read(ctx context.Context, userID string) (*userapp.User, error) {

	var rows *sql.Rows

	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := qb.Select("uuid").
		From("users").
		Where(squirrel.Eq{"uuid": userID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err = d.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var user userapp.User

	for rows.Next() {

		if err = rows.Scan(&user.UUID); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (d *UserStore) GetUserURLs(ctx context.Context, userID string) ([]urlapp.URL, error) {

	var rows *sql.Rows

	qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := qb.Select("short_url", "original_url").
		From("shorten").
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err = d.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var URLs []urlapp.URL

	for rows.Next() {

		var URL urlapp.URL
		if err = rows.Scan(&URL.Short, &URL.Long); err != nil {
			return nil, err
		}

		URLs = append(URLs, URL)
	}

	return URLs, nil
}

func (d *UserStore) DeleteUserURLs(ctx context.Context, shotrURLs []string, userID string) error {

	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Update("shorten").
		Set("is_deleted", true).
		Where(squirrel.Eq{"short_url": shotrURLs}).
		Where(squirrel.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		return err
	}

	result, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	logger.Log.Infof("Delete %d urls", rowsAffected)
	if len(shotrURLs) != int(rowsAffected) {
		return fmt.Errorf("not successful deletion")
	}
	return nil
}
