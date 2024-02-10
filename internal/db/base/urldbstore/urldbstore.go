package urldbstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/URL_shortener/internal/app/url"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var _ url.URLStore = &DBURLs{}

type DBURLs struct {
	db *sql.DB
}

func NewDB(ctx context.Context, ps string) (*DBURLs, error) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten (
        "uuid" TEXT PRIMARY KEY,
		"original_url" TEXT,
        "short_url" TEXT,
		"correlation_id" TEXT,
		UNIQUE (original_url, correlation_id)
      )`)

	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS ix_original_url ON shorten (original_url)")
	if err != nil {
		return nil, err
	}

	d := DBURLs{db: db}

	return &d, nil
}

func (d *DBURLs) Close() error {
	return d.db.Close()
}

func (d *DBURLs) Shortening(ctx context.Context, u []url.URL) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, uu := range u {

		_, err = tx.ExecContext(ctx, "INSERT INTO shorten (uuid, short_url, original_url, correlation_id) VALUES($1,$2,$3,$4)", uu.UUID.String(), uu.Short, uu.Long, uu.CorrelationID)

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

func (d *DBURLs) Resolve(ctx context.Context, shortURL string) (*url.URL, error) {

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

	var URL url.URL

	for rows.Next() {

		if err = rows.Scan(&URL.UUID, &URL.Long, &URL.Short, &URL.CorrelationID); err != nil {
			return nil, err
		}
	}

	return &URL, nil
}

func (d *DBURLs) Ping() bool {

	if err := d.db.Ping(); err != nil {
		return false
	}

	return true
}

func (d *DBURLs) getDuplicate(ctx context.Context, originalURL, correlationID string) (*url.URL, error) {

	row := d.db.QueryRowContext(ctx,
		"SELECT uuid, original_url, short_url, correlation_id FROM shorten WHERE original_url = $1 AND correlation_id = $2", originalURL, correlationID)

	var URL url.URL

	if err := row.Scan(&URL.UUID, &URL.Long, &URL.Short, &URL.CorrelationID); err != nil {
		return nil, err
	}

	return &URL, nil
}
