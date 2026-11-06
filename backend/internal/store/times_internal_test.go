package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		give time.Time
		want string
	}{
		{
			name: "happy path - utc",
			give: time.Date(2026, 10, 13, 7, 0, 0, 0, time.UTC),
			want: "2026-10-13T07:00:00Z",
		},
		{
			name: "happy path - offset converts to utc",
			give: time.Date(2026, 10, 13, 3, 0, 0, 0, time.FixedZone("EDT", -4*60*60)),
			want: "2026-10-13T07:00:00Z",
		},
		{
			name: "happy path - fractional seconds dropped",
			give: time.Date(2026, 10, 13, 7, 0, 0, 900_000_000, time.UTC),
			want: "2026-10-13T07:00:00Z",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			// when
			got := formatTime(tc.give)

			// then
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormatTimePtr(t *testing.T) {
	t.Run("happy path - formats a time", func(t *testing.T) {
		// given
		give := time.Date(2026, 10, 13, 7, 0, 0, 0, time.UTC)

		// when
		got := formatTimePtr(&give)

		// then
		require.NotNil(t, got)
		assert.Equal(t, "2026-10-13T07:00:00Z", *got)
	})

	t.Run("happy path - nil stays nil", func(t *testing.T) {
		// given
		// when
		got := formatTimePtr(nil)

		// then
		assert.Nil(t, got)
	})
}

func TestParseTime(t *testing.T) {
	t.Run("happy path - parses stored time", func(t *testing.T) {
		// given
		want := time.Date(2026, 10, 13, 7, 0, 0, 0, time.UTC)

		// when
		got, err := parseTime("2026-10-13T07:00:00Z")

		// then
		require.NoError(t, err)
		assert.True(t, want.Equal(got))
	})

	t.Run("error - bad stored time", func(t *testing.T) {
		// given
		// when
		_, err := parseTime("yesterday")

		// then
		require.ErrorContains(t, err, "parsing stored time")
	})
}

func TestPtr(t *testing.T) {
	t.Run("happy path - valid value", func(t *testing.T) {
		// given
		give := sql.Null[int]{V: 8, Valid: true}

		// when
		got := ptr(give)

		// then
		require.NotNil(t, got)
		assert.Equal(t, 8, *got)
	})

	t.Run("happy path - null gives nil", func(t *testing.T) {
		// given
		// when
		got := ptr(sql.Null[int]{})

		// then
		assert.Nil(t, got)
	})
}
