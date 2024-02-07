package urldbstore

import (
	"encoding/json"
	"fmt"

	"github.com/URL_shortener/internal/app/url"
)

type ErrConflict struct {
	Err error
	URL url.URL
}

func NewErrorConflict(err error, URL url.URL) error {
	return &ErrConflict{Err: err, URL: URL}
}

func (e *ErrConflict) Error() string {
	res, err := json.MarshalIndent(e.URL, "", "	")
	if err != nil {
		return fmt.Sprintf("%v : %v", e.Err, err.Error())
	}
	return fmt.Sprintf("%v : %v", e.Err, string(res))
}
