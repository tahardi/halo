package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tahardi/halo/backend/internal/model"
)

func TestSymptom_Valid(t *testing.T) {
	tests := []struct {
		name    string
		symptom model.Symptom
		want    bool
	}{
		{"aura", model.SymptomAura, true},
		{"brain fog", model.SymptomBrainFog, true},
		{"empty", "", false},
		{"unknown", "headache", false},
		{"wrong case", "Aura", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			symptom := tt.symptom

			// when
			got := symptom.Valid()

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}
