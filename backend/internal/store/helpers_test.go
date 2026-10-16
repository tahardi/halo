package store_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
	"github.com/tahardi/halo/backend/internal/store"
)

func newStore(t *testing.T) (*store.Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "halo.db")
	s, err := store.Open(t.Context(), path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s, path
}

func at(text string) time.Time {
	t, err := time.Parse(time.RFC3339, text)
	if err != nil {
		panic(err)
	}
	return t
}

func rizatriptan() model.Medication {
	return model.Medication{
		ID:          "med-riza",
		Name:        "Rizatriptan",
		UsualAmount: 1,
		Unit:        "tablet",
		Kind:        model.MedicationAcute,
	}
}

func attack(id string, started string) model.Attack {
	return model.Attack{
		ID:        id,
		StartedAt: at(started),
		Timezone:  "America/New_York",
		Symptoms:  []model.Symptom{},
		Doses:     []model.Dose{},
	}
}

func dose(id string, taken string) model.Dose {
	return model.Dose{ID: id, MedicationID: "med-riza", Amount: 1, TakenAt: at(taken)}
}
