package store_test

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
	"github.com/tahardi/halo/backend/internal/store"
)

func TestStore_PutAttack(t *testing.T) {
	t.Run("happy path - insert with doses and all fields", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		ended := at("2026-10-13T15:00:00Z")
		severity := 8
		notes := "Trail work, then a short ride."
		other := "jaw pain"
		lat, lon := 44.47, -73.21
		relief := model.ReliefSome
		a := attack("a-1", "2026-10-13T03:00:00-04:00")
		a.EndedAt = &ended
		a.PeakSeverity = &severity
		a.Symptoms = []model.Symptom{model.SymptomVomiting, model.SymptomNausea}
		a.SymptomsOther = &other
		a.Notes = &notes
		a.Lat = &lat
		a.Lon = &lon
		d := dose("d-1", "2026-10-13T07:00:00Z")
		d.Relief = &relief
		a.Doses = []model.Dose{d}

		// when
		got, err := s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		assert.Equal(t, "a-1", got.ID)
		assert.True(t, got.StartedAt.Equal(at("2026-10-13T07:00:00Z")))
		assert.Equal(t, time.UTC, got.StartedAt.Location())
		assert.True(t, got.EndedAt.Equal(ended))
		assert.Equal(t, 8, *got.PeakSeverity)
		assert.Equal(t, []model.Symptom{model.SymptomVomiting, model.SymptomNausea}, got.Symptoms)
		assert.Equal(t, other, *got.SymptomsOther)
		assert.Equal(t, notes, *got.Notes)
		assert.InDelta(t, lat, *got.Lat, 1e-9)
		assert.InDelta(t, lon, *got.Lon, 1e-9)
		assert.Equal(t, "app", got.Source)
		assert.False(t, got.CreatedAt.IsZero())
		require.Len(t, got.Doses, 1)
		assert.Equal(t, "d-1", got.Doses[0].ID)
		assert.Equal(t, model.ReliefSome, *got.Doses[0].Relief)
	})

	t.Run("happy path - update replaces doses and keeps source and created_at", func(t *testing.T) {
		// given
		s, path := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Source = "import"
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z"), dose("d-2", "2026-10-13T10:00:00Z")}
		first, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)
		a.Source = "app"
		a.Doses = []model.Dose{dose("d-2", "2026-10-13T10:00:00Z"), dose("d-3", "2026-10-13T12:00:00Z")}

		// when
		got, err := s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		assert.Equal(t, "import", got.Source)
		assert.True(t, first.CreatedAt.Equal(got.CreatedAt))
		require.Len(t, got.Doses, 2)
		assert.Equal(t, "d-2", got.Doses[0].ID)
		assert.Equal(t, "d-3", got.Doses[1].ID)
		raw, err := sql.Open("sqlite", path)
		require.NoError(t, err)
		t.Cleanup(func() { _ = raw.Close() })
		var kept int
		require.NoError(t, raw.QueryRowContext(
			t.Context(),
			"SELECT count(*) FROM doses WHERE id = 'd-1' AND deleted_at IS NOT NULL",
		).Scan(&kept))
		assert.Equal(t, 1, kept)
	})

	t.Run("happy path - update changes and clears fields", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.EndedAt = new(at("2026-10-13T15:00:00Z"))
		a.PeakSeverity = new(8)
		a.Symptoms = []model.Symptom{model.SymptomNausea}
		a.SymptomsOther = new("jaw pain")
		a.Notes = new("Trail work.")
		a.Lat = new(44.47)
		a.Lon = new(-73.21)
		_, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)
		changed := attack("a-1", "2026-10-13T08:00:00Z")
		changed.Timezone = "America/Denver"
		changed.PeakSeverity = new(3)
		changed.Symptoms = []model.Symptom{model.SymptomVomiting}

		// when
		got, err := s.PutAttack(t.Context(), changed)

		// then
		require.NoError(t, err)
		assert.True(t, got.StartedAt.Equal(changed.StartedAt))
		assert.Equal(t, "America/Denver", got.Timezone)
		assert.Equal(t, 3, *got.PeakSeverity)
		assert.Equal(t, []model.Symptom{model.SymptomVomiting}, got.Symptoms)
		assert.Nil(t, got.EndedAt)
		assert.Nil(t, got.SymptomsOther)
		assert.Nil(t, got.Notes)
		assert.Nil(t, got.Lat)
		assert.Nil(t, got.Lon)
	})

	t.Run("happy path - same attack twice stores one attack and one dose", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}
		_, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)

		// when
		_, err = s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		all, err := s.AllAttacks(t.Context())
		require.NoError(t, err)
		require.Len(t, all, 1)
		assert.Len(t, all[0].Doses, 1)
	})

	t.Run("happy path - put after delete restores the attack and its doses", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Source = "import"
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}
		first, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)
		require.NoError(t, s.DeleteAttack(t.Context(), "a-1"))
		a.Source = "app"

		// when
		_, err = s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		got, err := s.GetAttack(t.Context(), "a-1")
		require.NoError(t, err)
		assert.Equal(t, "import", got.Source)
		assert.True(t, first.CreatedAt.Equal(got.CreatedAt))
		require.Len(t, got.Doses, 1)
		assert.Equal(t, "d-1", got.Doses[0].ID)
	})

	t.Run("happy path - dose can use an archived medication", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		archived := rizatriptan()
		archived.Archived = true
		_, err := s.PutMedication(t.Context(), archived)
		require.NoError(t, err)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}

		// when
		got, err := s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		assert.Len(t, got.Doses, 1)
	})

	t.Run("happy path - concurrent saves all land", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		const count = 20
		errs := make([]error, count)
		var wg sync.WaitGroup

		// when
		for i := range count {
			wg.Go(func() {
				a := attack(fmt.Sprintf("a-%d", i), "2026-10-13T07:00:00Z")
				a.Doses = []model.Dose{dose(fmt.Sprintf("d-%d", i), "2026-10-13T07:00:00Z")}
				_, errs[i] = s.PutAttack(t.Context(), a)
			})
		}
		wg.Wait()

		// then
		for _, err := range errs {
			require.NoError(t, err)
		}
		all, err := s.AllAttacks(t.Context())
		require.NoError(t, err)
		assert.Len(t, all, count)
	})

	t.Run("happy path - nil slices come back empty", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Symptoms = nil
		a.Doses = nil

		// when
		got, err := s.PutAttack(t.Context(), a)

		// then
		require.NoError(t, err)
		assert.NotNil(t, got.Symptoms)
		assert.NotNil(t, got.Doses)
	})

	tests := []struct {
		name string
		edit func(a *model.Attack)
	}{
		{name: "error - ended before started", edit: func(a *model.Attack) {
			a.EndedAt = new(at("2026-10-13T06:00:00Z"))
		}},
		{name: "error - empty id", edit: func(a *model.Attack) { a.ID = "" }},
		{name: "error - empty timezone", edit: func(a *model.Attack) { a.Timezone = "" }},
		{name: "error - severity above 10", edit: func(a *model.Attack) { a.PeakSeverity = new(11) }},
		{name: "error - severity below 0", edit: func(a *model.Attack) { a.PeakSeverity = new(-1) }},
		{name: "error - lat without lon", edit: func(a *model.Attack) { a.Lat = new(44.47) }},
		{name: "error - lat out of range", edit: func(a *model.Attack) {
			a.Lat = new(91.0)
			a.Lon = new(0.0)
		}},
		{name: "error - lon out of range", edit: func(a *model.Attack) {
			a.Lat = new(0.0)
			a.Lon = new(181.0)
		}},
		{name: "error - bad source", edit: func(a *model.Attack) { a.Source = "fax" }},
		{name: "error - bad relief", edit: func(a *model.Attack) { a.Doses[0].Relief = new(model.Relief("lots")) }},
		{name: "error - zero amount", edit: func(a *model.Attack) { a.Doses[0].Amount = 0 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			s, _ := storeWithMed(t)
			a := attack("a-1", "2026-10-13T07:00:00Z")
			a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}
			tc.edit(&a)

			// when
			_, err := s.PutAttack(t.Context(), a)

			// then
			require.Error(t, err)
			all, err := s.AllAttacks(t.Context())
			require.NoError(t, err)
			assert.Empty(t, all)
		})
	}

	t.Run("error - dose id already on another attack", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		first := attack("a-1", "2026-10-13T07:00:00Z")
		first.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}
		_, err := s.PutAttack(t.Context(), first)
		require.NoError(t, err)
		second := attack("a-2", "2026-10-14T07:00:00Z")
		second.Doses = []model.Dose{dose("d-1", "2026-10-14T07:00:00Z")}

		// when
		_, err = s.PutAttack(t.Context(), second)

		// then
		require.ErrorIs(t, err, store.ErrDuplicate)
		_, err = s.GetAttack(t.Context(), "a-2")
		require.ErrorIs(t, err, store.ErrNotFound)
		kept, err := s.GetAttack(t.Context(), "a-1")
		require.NoError(t, err)
		require.Len(t, kept.Doses, 1)
		assert.True(t, kept.Doses[0].TakenAt.Equal(at("2026-10-13T07:00:00Z")))
	})

	t.Run("error - failed update keeps the old doses", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}
		_, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)
		bad := dose("d-2", "2026-10-13T08:00:00Z")
		bad.MedicationID = "med-missing"
		a.Doses = []model.Dose{bad}

		// when
		_, err = s.PutAttack(t.Context(), a)

		// then
		require.Error(t, err)
		got, err := s.GetAttack(t.Context(), "a-1")
		require.NoError(t, err)
		require.Len(t, got.Doses, 1)
		assert.Equal(t, "d-1", got.Doses[0].ID)
	})

	t.Run("error - unknown medication leaves no attack behind", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		bad := dose("d-1", "2026-10-13T07:00:00Z")
		bad.MedicationID = "med-missing"
		a.Doses = []model.Dose{bad}

		// when
		_, err := s.PutAttack(t.Context(), a)

		// then
		require.Error(t, err)
		_, err = s.GetAttack(t.Context(), "a-1")
		require.ErrorIs(t, err, store.ErrNotFound)
	})
}

func TestStore_GetAttack(t *testing.T) {
	tests := []struct {
		name    string
		deleted bool
		id      string
		wantErr error
	}{
		{name: "happy path - found", id: "a-1"},
		{name: "error - unknown id", id: "a-2", wantErr: store.ErrNotFound},
		{name: "error - soft deleted", id: "a-1", deleted: true, wantErr: store.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			s, _ := storeWithMed(t)
			_, err := s.PutAttack(t.Context(), attack("a-1", "2026-10-13T07:00:00Z"))
			require.NoError(t, err)
			if tc.deleted {
				require.NoError(t, s.DeleteAttack(t.Context(), "a-1"))
			}

			// when
			got, err := s.GetAttack(t.Context(), tc.id)

			// then
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "a-1", got.ID)
		})
	}
}

func TestStore_ListAttacks(t *testing.T) {
	starts := map[string]string{
		"a-1": "2026-09-23T19:10:00Z",
		"a-2": "2026-09-28T19:00:00Z",
		"a-3": "2026-10-01T23:30:00Z",
		"a-4": "2026-10-02T23:24:00Z",
	}
	tests := []struct {
		name    string
		before  string
		limit   int
		deleted string
		want    []string
	}{
		{name: "happy path - newest first", before: "2026-12-01T00:00:00Z", limit: 50,
			want: []string{"a-4", "a-3", "a-2", "a-1"}},
		{name: "happy path - limit", before: "2026-12-01T00:00:00Z", limit: 2, want: []string{"a-4", "a-3"}},
		{name: "happy path - before is exclusive", before: "2026-10-01T23:30:00Z", limit: 50,
			want: []string{"a-2", "a-1"}},
		{name: "happy path - deleted skipped", before: "2026-12-01T00:00:00Z", limit: 50, deleted: "a-3",
			want: []string{"a-4", "a-2", "a-1"}},
		{name: "happy path - before with offset", before: "2026-10-01T19:30:00-04:00", limit: 50,
			want: []string{"a-2", "a-1"}},
		{name: "happy path - nothing before", before: "2026-01-01T00:00:00Z", limit: 50, want: []string{}},
		{name: "happy path - limit 0", before: "2026-12-01T00:00:00Z", limit: 0, want: []string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			s, _ := storeWithMed(t)
			for id, start := range starts {
				_, err := s.PutAttack(t.Context(), attack(id, start))
				require.NoError(t, err)
			}
			if tc.deleted != "" {
				require.NoError(t, s.DeleteAttack(t.Context(), tc.deleted))
			}

			// when
			got, err := s.ListAttacks(t.Context(), at(tc.before), tc.limit)

			// then
			require.NoError(t, err)
			assert.NotNil(t, got)
			ids := []string{}
			for _, a := range got {
				ids = append(ids, a.ID)
			}
			assert.Equal(t, tc.want, ids)
		})
	}

	t.Run("happy path - same start sorts by id descending", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		for _, id := range []string{"a-1", "a-3", "a-2"} {
			_, err := s.PutAttack(t.Context(), attack(id, "2026-10-13T07:00:00Z"))
			require.NoError(t, err)
		}

		// when
		got, err := s.ListAttacks(t.Context(), at("2026-12-01T00:00:00Z"), 50)

		// then
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, "a-3", got[0].ID)
		assert.Equal(t, "a-2", got[1].ID)
		assert.Equal(t, "a-1", got[2].ID)
	})

	t.Run("happy path - loads live doses only", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z"), dose("d-2", "2026-10-13T08:00:00Z")}
		_, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)
		a.Doses = []model.Dose{dose("d-2", "2026-10-13T08:00:00Z")}
		_, err = s.PutAttack(t.Context(), a)
		require.NoError(t, err)

		// when
		got, err := s.ListAttacks(t.Context(), at("2026-12-01T00:00:00Z"), 50)

		// then
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Len(t, got[0].Doses, 1)
		assert.Equal(t, "d-2", got[0].Doses[0].ID)
	})
}

func TestStore_AllAttacks(t *testing.T) {
	t.Run("happy path - oldest first with doses, deleted skipped", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		late := attack("a-late", "2026-10-13T07:00:00Z")
		late.Doses = []model.Dose{dose("d-2", "2026-10-13T10:00:00Z"), dose("d-1", "2026-10-13T07:00:00Z")}
		for _, a := range []model.Attack{
			late,
			attack("a-early", "2026-09-23T19:10:00Z"),
			attack("a-gone", "2026-10-01T23:30:00Z"),
		} {
			_, err := s.PutAttack(t.Context(), a)
			require.NoError(t, err)
		}
		require.NoError(t, s.DeleteAttack(t.Context(), "a-gone"))

		// when
		got, err := s.AllAttacks(t.Context())

		// then
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "a-early", got[0].ID)
		assert.Equal(t, "a-late", got[1].ID)
		require.Len(t, got[1].Doses, 2)
		assert.Equal(t, "d-1", got[1].Doses[0].ID)
	})

	t.Run("happy path - same start sorts by id", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)
		for _, id := range []string{"a-2", "a-3", "a-1"} {
			_, err := s.PutAttack(t.Context(), attack(id, "2026-10-13T07:00:00Z"))
			require.NoError(t, err)
		}

		// when
		got, err := s.AllAttacks(t.Context())

		// then
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, "a-1", got[0].ID)
		assert.Equal(t, "a-2", got[1].ID)
		assert.Equal(t, "a-3", got[2].ID)
	})

	t.Run("happy path - empty database gives an empty list", func(t *testing.T) {
		// given
		s, _ := storeWithMed(t)

		// when
		got, err := s.AllAttacks(t.Context())

		// then
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
}

func TestStore_DeleteAttack(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		twice   bool
		wantErr error
	}{
		{name: "happy path - deletes", id: "a-1"},
		{name: "error - unknown id", id: "a-2", wantErr: store.ErrNotFound},
		{name: "error - already deleted", id: "a-1", twice: true, wantErr: store.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			s, _ := storeWithMed(t)
			_, err := s.PutAttack(t.Context(), attack("a-1", "2026-10-13T07:00:00Z"))
			require.NoError(t, err)
			if tc.twice {
				require.NoError(t, s.DeleteAttack(t.Context(), tc.id))
			}

			// when
			err = s.DeleteAttack(t.Context(), tc.id)

			// then
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			_, err = s.GetAttack(t.Context(), tc.id)
			require.ErrorIs(t, err, store.ErrNotFound)
		})
	}

	t.Run("happy path - leaves doses in place", func(t *testing.T) {
		// given
		s, path := storeWithMed(t)
		a := attack("a-1", "2026-10-13T07:00:00Z")
		a.Doses = []model.Dose{dose("d-1", "2026-10-13T07:00:00Z")}
		_, err := s.PutAttack(t.Context(), a)
		require.NoError(t, err)

		// when
		err = s.DeleteAttack(t.Context(), "a-1")

		// then
		require.NoError(t, err)
		var live int
		require.NoError(t, rawDB(t, path).QueryRowContext(
			t.Context(),
			"SELECT count(*) FROM doses WHERE id = 'd-1' AND deleted_at IS NULL",
		).Scan(&live))
		assert.Equal(t, 1, live)
	})
}
