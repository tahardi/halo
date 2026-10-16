package weather_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tahardi/halo/backend/internal/weather"
)

func TestRound(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"happy path - rounds down", 44.4812, 44.48},
		{"happy path - rounds up", 44.4851, 44.49},
		{"happy path - negative", -73.2149, -73.21},
		{"happy path - already rounded", 12.5, 12.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := weather.Round(tt.in)

			// then
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}
