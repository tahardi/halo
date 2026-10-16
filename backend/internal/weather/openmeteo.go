package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tahardi/halo/backend/internal/model"
)

const (
	ForecastURL = "https://api.open-meteo.com/v1/forecast"
	ArchiveURL  = "https://archive-api.open-meteo.com/v1/archive"

	archiveDelay   = 5 * 24 * time.Hour
	requestTimeout = 30 * time.Second
	hourlyFields   = "pressure_msl,temperature_2m,relative_humidity_2m,precipitation"
	maxBody        = 4 << 20
)

var ErrUpstream = errors.New("fetching weather")

type OpenMeteo struct {
	Client      *http.Client
	ForecastURL string
	ArchiveURL  string
	Now         func() time.Time
}

func NewOpenMeteo() *OpenMeteo {
	return &OpenMeteo{
		Client:      &http.Client{Timeout: requestTimeout},
		ForecastURL: ForecastURL,
		ArchiveURL:  ArchiveURL,
		Now:         time.Now,
	}
}

type response struct {
	Reason string `json:"reason"`
	Hourly struct {
		Time          []string   `json:"time"`
		PressureMSL   []*float64 `json:"pressure_msl"`
		Temperature   []*float64 `json:"temperature_2m"`
		Humidity      []*float64 `json:"relative_humidity_2m"`
		Precipitation []*float64 `json:"precipitation"`
	} `json:"hourly"`
}

func (o *OpenMeteo) Hours(
	ctx context.Context,
	lat float64,
	lon float64,
	from time.Time,
	to time.Time,
) ([]model.WeatherHour, error) {
	lat, lon = Round(lat), Round(lon)
	from, to = from.UTC().Truncate(time.Hour), to.UTC().Truncate(time.Hour)

	base, source := o.ForecastURL, "open-meteo-forecast"
	if to.Before(o.Now().UTC().Add(-archiveDelay)) {
		base, source = o.ArchiveURL, "open-meteo-archive"
	}
	q := url.Values{}
	q.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(lon, 'f', -1, 64))
	q.Set("start_date", from.Format(time.DateOnly))
	q.Set("end_date", to.Format(time.DateOnly))
	q.Set("hourly", hourlyFields)
	q.Set("timezone", "UTC")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("building weather request: %w", err)
	}
	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstream, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading body: %w", ErrUpstream, err)
	}
	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("%w: status %d: decoding body: %w", ErrUpstream, resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", ErrUpstream, resp.StatusCode, parsed.Reason)
	}

	h := parsed.Hourly
	n := len(h.Time)
	if len(h.PressureMSL) != n || len(h.Temperature) != n || len(h.Humidity) != n || len(h.Precipitation) != n {
		return nil, fmt.Errorf("%w: hourly arrays have different lengths", ErrUpstream)
	}
	out := []model.WeatherHour{}
	for i, raw := range h.Time {
		at, err := time.Parse("2006-01-02T15:04", raw)
		if err != nil {
			return nil, fmt.Errorf("%w: parsing hour %q: %w", ErrUpstream, raw, err)
		}
		if at.Before(from) || at.After(to) {
			continue
		}
		if h.PressureMSL[i] == nil || h.Temperature[i] == nil || h.Humidity[i] == nil || h.Precipitation[i] == nil {
			continue
		}
		out = append(out, model.WeatherHour{
			LatRound:        lat,
			LonRound:        lon,
			Hour:            at.UTC(),
			PressureHpa:     *h.PressureMSL[i],
			TemperatureC:    *h.Temperature[i],
			HumidityPct:     *h.Humidity[i],
			PrecipitationMm: *h.Precipitation[i],
			Source:          source,
		})
	}
	return out, nil
}
