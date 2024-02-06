package urldbstore

import (
	"context"
	"database/sql"
	"errors"

	"github.com/URL_shortener/internal/app/url"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/lib/pq"
)

var _ url.URLStore = &DBURLs{}

var ErrConflict = errors.New("data conflict")

type DBURLs struct {
	db *sql.DB
}

func NewDB(ctx context.Context, ps string) (*DBURLs, error) {

	db, err := sql.Open("postgres", ps)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten (
        "uuid" TEXT PRIMARY KEY,
		"original_url" TEXT UNIQUE,
        "short_url" TEXT,
		"correlation_id" TEXT
      )`)

	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS ix_shorten_original_url ON shorten (original_url)")
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

		stmt, err := tx.PrepareContext(ctx,
			"INSERT INTO shorten (uuid, original_url, short_url, correlation_id) VALUES($1,$2,$3,$4) ON CONFLICT (original_url) DO NOTHING")

		if err != nil {
			return err
		}

		defer stmt.Close()

		_, err = stmt.ExecContext(ctx, uu.UUID.String(), uu.Long, uu.Short, uu.CorrelationID)

		if err != nil {
			// проверяем, что ошибка сигнализирует о потенциальном нарушении целостности данных
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return ErrConflict
			}
			return err
		}
	}

	return tx.Commit()

}

func (d *DBURLs) Resolve(ctx context.Context, shortURL string) (*url.URL, error) {

	var rows *sql.Rows

	rows, err := d.db.QueryContext(ctx,
		"SELECT * FROM shorten WHERE short_url=$1", shortURL)

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
