package weather

import (
	"context"
	"math"
	"time"

	"github.com/tahardi/halo/backend/internal/model"
)

type Source interface {
	Hours(ctx context.Context, lat float64, lon float64, from time.Time, to time.Time) ([]model.WeatherHour, error)
}

const roundFactor = 100

func Round(v float64) float64 {
	return math.Round(v*roundFactor) / roundFactor
}
