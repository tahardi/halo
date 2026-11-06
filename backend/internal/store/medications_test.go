package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
	"github.com/tahardi/halo/backend/internal/store"
)

func TestStore_PutMedication(t *testing.T) {
	t.Run("happy path - insert then update", func(t *testing.T) {
		// given
		s, _ := newStore(t)
		_, err := s.PutMedication(t.Context(), rizatriptan())
		require.NoError(t, err)
		updated := rizatriptan()
		updated.UsualAmount = 2
		updated.Archived = true

		// when
		got, err := s.PutMedication(t.Context(), updated)

		// then
		require.NoError(t, err)
		assert.Equal(t, updated, got)
		meds, err := s.ListMedications(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []model.Medication{updated}, meds)
	})

	t.Run("happy path - changes only the case of its own name", func(t *testing.T) {
		// given
		s, _ := newStore(t)
		_, err := s.PutMedication(t.Context(), rizatriptan())
		require.NoError(t, err)
		updated := rizatriptan()
		updated.Name = "RIZATRIPTAN"

		// when
		_, err = s.PutMedication(t.Context(), updated)

		// then
		require.NoError(t, err)
		meds, err := s.ListMedications(t.Context())
		require.NoError(t, err)
		assert.Equal(t, []model.Medication{updated}, meds)
	})

	t.Run("error - same name with different case and id", func(t *testing.T) {
		// given
		s, _ := newStore(t)
		_, err := s.PutMedication(t.Context(), rizatriptan())
		require.NoError(t, err)
		clash := rizatriptan()
		clash.ID = "med-other"
		clash.Name = "RIZATRIPTAN"

		// when
		_, err = s.PutMedication(t.Context(), clash)

		// then
		require.ErrorIs(t, err, store.ErrDuplicate)
	})

	t.Run("error - update renames into another medication", func(t *testing.T) {
		// given
		s, _ := newStore(t)
		_, err := s.PutMedication(t.Context(), rizatriptan())
		require.NoError(t, err)
		other := rizatriptan()
		other.ID = "med-other"
		other.Name = "Ibuprofen"
		_, err = s.PutMedication(t.Context(), other)
		require.NoError(t, err)
		other.Name = "rizatriptan"

		// when
		_, err = s.PutMedication(t.Context(), other)

		// then
		require.ErrorIs(t, err, store.ErrDuplicate)
	})

	tests := []struct {
		name string
		edit func(m *model.Medication)
	}{
		{name: "error - invalid kind", edit: func(m *model.Medication) { m.Kind = "daily" }},
		{name: "error - blank name", edit: func(m *model.Medication) { m.Name = "  " }},
		{name: "error - blank unit", edit: func(m *model.Medication) { m.Unit = "  " }},
		{name: "error - zero amount", edit: func(m *model.Medication) { m.UsualAmount = 0 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			s, _ := newStore(t)
			bad := rizatriptan()
			tc.edit(&bad)

			// when
			_, err := s.PutMedication(t.Context(), bad)

			// then
			require.Error(t, err)
			require.NotErrorIs(t, err, store.ErrDuplicate)
		})
	}
}

func TestStore_ListMedications(t *testing.T) {
	t.Run("happy path - sorted by name ignoring case, archived included", func(t *testing.T) {
		// given
		s, _ := newStore(t)
		names := map[string]string{"m-1": "sumatriptan", "m-2": "Ibuprofen", "m-3": "amitriptyline"}
		for id, name := range names {
			m := rizatriptan()
			m.ID = id
			m.Name = name
			m.Archived = id == "m-2"
			_, err := s.PutMedication(t.Context(), m)
			require.NoError(t, err)
		}

		// when
		meds, err := s.ListMedications(t.Context())

		// then
		require.NoError(t, err)
		require.Len(t, meds, 3)
		assert.Equal(t, "amitriptyline", meds[0].Name)
		assert.Equal(t, "Ibuprofen", meds[1].Name)
		assert.True(t, meds[1].Archived)
		assert.Equal(t, "sumatriptan", meds[2].Name)
	})

	t.Run("happy path - empty list is not nil", func(t *testing.T) {
		// given
		s, _ := newStore(t)

		// when
		meds, err := s.ListMedications(t.Context())

		// then
		require.NoError(t, err)
		assert.NotNil(t, meds)
		assert.Empty(t, meds)
	})
}
