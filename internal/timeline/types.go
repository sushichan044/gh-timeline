package timeline

import (
	"encoding/json"
	"net/url"
	"time"
)

// DateTime wraps [time.Time] with GitHub's RFC3339 JSON format.
type DateTime struct{ time.Time }

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	dt.Time = t
	return nil
}

// URI wraps [url.URL] with JSON string unmarshaling.
type URI struct{ URL *url.URL }

func (u *URI) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := url.Parse(s)
	if err != nil {
		return err
	}
	u.URL = parsed
	return nil
}
