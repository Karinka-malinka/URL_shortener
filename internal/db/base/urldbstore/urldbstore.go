package urldbstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/URL_shortener/internal/app/urlapp"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var _ urlapp.URLStore = &URLStore{}

type URLStore struct {
	db *sql.DB
}

func NewURLStore(db *sql.DB) *URLStore {

	return &URLStore{db: db}
}

func (d *URLStore) Shortening(ctx context.Context, u []urlapp.URL) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, uu := range u {

		_, err = tx.ExecContext(ctx, "INSERT INTO shorten (uuid, short_url, original_url, correlation_id, user_id) VALUES($1,$2,$3,$4,$5)", uu.UUID.String(), uu.Short, uu.Long, uu.CorrelationID, uu.UserID)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				exURL, err := d.getDuplicate(ctx, uu.Long, uu.CorrelationID)
				if err != nil {
					return err
				}
				return NewErrorConflict(err, *exURL)
			}

			return err
		}
	}

	return tx.Commit()
}

func (d *URLStore) getDuplicate(ctx context.Context, originalURL, correlationID string) (*urlapp.URL, error) {

	row := d.db.QueryRowContext(ctx,
		"SELECT uuid, original_url, short_url, correlation_id FROM shorten WHERE original_url = $1 AND correlation_id = $2", originalURL, correlationID)

	var URL urlapp.URL

	if err := row.Scan(&URL.UUID, &URL.Long, &URL.Short, &URL.CorrelationID); err != nil {
		return nil, err
	}

	return &URL, nil
}

func (d *URLStore) Resolve(ctx context.Context, shortURL string) (*urlapp.URL, error) {

	var rows *sql.Rows

	rows, err := d.db.QueryContext(ctx,
		"SELECT uuid, original_url, short_url, correlation_id FROM shorten WHERE short_url=$1", shortURL)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	defer rows.Close()

	var URL urlapp.URL

	for rows.Next() {

		if err = rows.Scan(&URL.UUID, &URL.Long, &URL.Short, &URL.CorrelationID); err != nil {
			return nil, err
		}
	}

	return &URL, nil
}

func (d *URLStore) Ping() bool {

	if err := d.db.Ping(); err != nil {
		return false
	}

	return true
}
