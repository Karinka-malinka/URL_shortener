package userdbstore

import (
	"context"
	"database/sql"

	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/URL_shortener/internal/app/userapp"

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

	rows, err := d.db.QueryContext(ctx,
		"SELECT uuid FROM users WHERE uuid = $1", userID)

	/*qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := qb.Select("uuid").
		From("users").
		Where(squirrel.Eq{"uuid": userID}).
		ToSql()

	rows, err = d.db.QueryContext(ctx, query, args)*/

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

	// Create a query builder instance
	/*qb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := qb.Select("short_url", "original_url").
		From("shorten").
		InnerJoin("users ON users.uuid = shorten.user_id").
		Where(squirrel.Eq{"users.uuid": userID}).
		ToSql()

	*/

	rows, err := d.db.QueryContext(ctx,
		"SELECT short_url, original_url FROM shorten INNER JOIN users ON users.uuid = shorten.user_id WHERE users.uuid = $1", userID)

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
