package urldbstore

import (
	"encoding/json"
	"fmt"

	"github.com/URL_shortener/internal/app/urlapp"
)

type ErrConflict struct {
	Err error
	URL urlapp.URL
}

func NewErrorConflict(err error, URL urlapp.URL) error {
	return &ErrConflict{Err: err, URL: URL}
}

func (e *ErrConflict) Error() string {
	res, err := json.MarshalIndent(e.URL, "", "	")
	if err != nil {
		return fmt.Sprintf("%v : %v", e.Err, err.Error())
	}
	return fmt.Sprintf("%v : %v", e.Err, string(res))
}
