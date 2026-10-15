package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tahardi/halo/backend/internal/version"
)

func TestString(t *testing.T) {
	// given
	want := "dev"

	// when
	got := version.String()

	// then
	assert.Equal(t, want, got)
}
