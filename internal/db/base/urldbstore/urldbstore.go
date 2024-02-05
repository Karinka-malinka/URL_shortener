package urldbstore

import (
	"context"
	"database/sql"

	"github.com/URL_shortener/internal/app/url"
	_ "github.com/lib/pq"
)

var _ url.URLStore = &DBURLs{}

type DBURLs struct {
	db *sql.DB
}

func NewDB(ctx context.Context, ps string) (*DBURLs, error) {

	db, err := sql.Open("postgres", ps)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten (
        "uuid" TEXT,
        "short_url" TEXT,
        "original_url" TEXT
      )`)

	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS original_url ON shorten (original_url)")
	if err != nil {
		return nil, err
	}

	d := DBURLs{db: db}

	return &d, nil
}

func (d *DBURLs) Close() error {
	return d.db.Close()
}

func (d *DBURLs) Shortening(ctx context.Context, u url.URL) error {

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO shorten (uuid, short_url, original_url) VALUES($1,$2,$3)")

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, u.UUID.String(), u.Short, u.Long)

	if err != nil {
		return err
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
	defer rows.Close()

	var URL url.URL

	for rows.Next() {

		if err = rows.Scan(&URL); err != nil {
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
