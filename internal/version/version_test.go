package version_test

import (
	"testing"

	"github.com/sushichan044/gh-timeline/internal/version"
)

func TestGet(t *testing.T) {
	t.Parallel()
	if v := version.Get(); v == "" {
		t.Fatal("version.Get() returned empty string")
	}
}
