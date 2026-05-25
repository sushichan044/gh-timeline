//nolint:testpackage // white-box test exercises unexported package types directly.
package timeline

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateTime_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "valid RFC3339 UTC timestamp",
			input: `"2026-01-02T10:00:00Z"`,
			want:  time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			name:  "valid RFC3339 timestamp with offset",
			input: `"2026-01-02T19:00:00+09:00"`,
			want:  time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		{
			name:    "non-RFC3339 string",
			input:   `"2026-01-02"`,
			wantErr: true,
		},
		{
			name:    "non-string JSON value",
			input:   `1234567890`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got DateTime
			err := json.Unmarshal([]byte(tc.input), &got)
			if (err != nil) != tc.wantErr {
				t.Fatalf("UnmarshalJSON(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if !got.Time.Equal(tc.want) {
				t.Errorf("UnmarshalJSON(%q) = %v, want %v", tc.input, got.Time, tc.want)
			}
		})
	}
}

func TestURI_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantStr string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			input:   `"https://github.com/owner/repo"`,
			wantStr: "https://github.com/owner/repo",
		},
		{
			name:    "empty string produces empty URL",
			input:   `""`,
			wantStr: "",
		},
		{
			name:    "non-string JSON value",
			input:   `123`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got URI
			err := json.Unmarshal([]byte(tc.input), &got)
			if (err != nil) != tc.wantErr {
				t.Fatalf("UnmarshalJSON(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if got.URL == nil {
				t.Fatalf("UnmarshalJSON(%q) URL is nil", tc.input)
			}
			if got.URL.String() != tc.wantStr {
				t.Errorf("UnmarshalJSON(%q) = %q, want %q", tc.input, got.URL.String(), tc.wantStr)
			}
		})
	}
}
