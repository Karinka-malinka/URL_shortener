package urldbstore

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type URL struct {
	UUID  string `json:"uuid"`
	Short string `json:"short_url"`
	Long  string `json:"original_url"`
}

type DBURLs struct {
	db *sql.DB
}

func NewDB(ps string) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		panic(err)
	}
	defer db.Close()

}
