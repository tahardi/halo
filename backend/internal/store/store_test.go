package store_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
	"github.com/tahardi/halo/backend/internal/store"
)

func TestOpen(t *testing.T) {
	t.Run("happy path - creates schema and reopens without losing data", func(t *testing.T) {
		// given
		s, path := newStore(t)
		_, err := s.PutMedication(t.Context(), rizatriptan())
		require.NoError(t, err)
		require.NoError(t, s.Close())

		// when
		reopened, err := store.Open(t.Context(), path)

		// then
		require.NoError(t, err)
		t.Cleanup(func() { _ = reopened.Close() })
		meds, err := reopened.ListMedications(t.Context())
		require.NoError(t, err)
		assert.Len(t, meds, 1)

		raw, err := sql.Open("sqlite", path)
		require.NoError(t, err)
		t.Cleanup(func() { _ = raw.Close() })
		var version int
		require.NoError(t, raw.QueryRowContext(t.Context(), "PRAGMA user_version").Scan(&version))
		assert.Equal(t, 1, version)
		var tables int
		require.NoError(t, raw.QueryRowContext(
			t.Context(),
			`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN
				('attacks', 'medications', 'doses', 'recordings', 'extractions', 'weather_hours')`,
		).Scan(&tables))
		assert.Equal(t, 6, tables)
	})

	t.Run("happy path - foreign keys enforced", func(t *testing.T) {
		// given
		s, _ := newStore(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}

		// when
		_, err := s.PutAttack(t.Context(), a)

		// then
		require.Error(t, err)
	})

	t.Run("error - schema newer than binary", func(t *testing.T) {
		// given
		path := filepath.Join(t.TempDir(), "halo.db")
		raw, err := sql.Open("sqlite", path)
		require.NoError(t, err)
		_, err = raw.ExecContext(t.Context(), "PRAGMA user_version = 99")
		require.NoError(t, err)
		require.NoError(t, raw.Close())

		// when
		_, err = store.Open(t.Context(), path)

		// then
		require.ErrorIs(t, err, store.ErrSchemaTooNew)
	})

	t.Run("error - folder does not exist", func(t *testing.T) {
		// given
		path := filepath.Join(t.TempDir(), "missing", "halo.db")

		// when
		_, err := store.Open(t.Context(), path)

		// then
		require.Error(t, err)
	})
}
