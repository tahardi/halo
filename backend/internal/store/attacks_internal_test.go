package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
)

func TestStore_PutAttack(t *testing.T) {
	t.Run("happy path - update advances updated_at and keeps created_at", func(t *testing.T) {
		// given
		s, err := Open(t.Context(), filepath.Join(t.TempDir(), "halo.db"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = s.Close() })
		created := time.Date(2026, 10, 13, 7, 0, 0, 0, time.UTC)
		updated := created.Add(time.Hour)
		a := model.Attack{ID: "a-1", StartedAt: created, Timezone: "UTC"}
		s.now = func() time.Time { return created }
		_, err = s.PutAttack(t.Context(), a)
		require.NoError(t, err)
		s.now = func() time.Time { return updated }

		// when
		got, err := s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		assert.True(t, created.Equal(got.CreatedAt))
		assert.True(t, updated.Equal(got.UpdatedAt))
	})
}
