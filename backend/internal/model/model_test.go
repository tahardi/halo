package model_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
)

func TestFixtures_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		file string
		into func() any
	}{
		{"health response", "health-response.json", func() any { return &model.HealthResponse{} }},
		{"attack", "attack.json", func() any { return &model.Attack{} }},
		{"attack page", "attack-page.json", func() any { return &model.AttackPage{} }},
		{"attack detail", "attack-detail.json", func() any { return &model.AttackDetail{} }},
		{"medication list", "medication-list.json", func() any { return &model.MedicationList{} }},
		{"recording extracted", "recording-extracted.json", func() any { return &model.Recording{} }},
		{"recording failed", "recording-failed.json", func() any { return &model.Recording{} }},
		{"error bad request", "error-bad-request.json", func() any { return &model.ErrorResponse{} }},
		{"error unauthorized", "error-unauthorized.json", func() any { return &model.ErrorResponse{} }},
		{"error not found", "error-not-found.json", func() any { return &model.ErrorResponse{} }},
		{"error no speech", "error-no-speech.json", func() any { return &model.ErrorResponse{} }},
		{"error upstream failed", "error-upstream-failed.json", func() any { return &model.ErrorResponse{} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			original, err := os.ReadFile("../../../api/testdata/" + tt.file)
			require.NoError(t, err)
			got := tt.into()

			// when
			decoder := json.NewDecoder(bytes.NewReader(original))
			decoder.DisallowUnknownFields()
			require.NoError(t, decoder.Decode(got))
			reencoded, err := json.Marshal(got)
			require.NoError(t, err)

			// then
			assert.JSONEq(t, string(original), string(reencoded))
		})
	}
}
