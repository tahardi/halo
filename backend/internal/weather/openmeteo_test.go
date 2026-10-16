package weather_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
	"github.com/tahardi/halo/backend/internal/weather"
)

type call struct {
	path  string
	query url.Values
}

func fixtureServer(t *testing.T, status int, file string) (*httptest.Server, *[]call) {
	t.Helper()
	body, err := os.ReadFile("../../test/testdata/" + file)
	require.NoError(t, err)
	calls := &[]call{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls = append(*calls, call{path: r.URL.Path, query: r.URL.Query()})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv, calls
}

func source(srv *httptest.Server, now time.Time) *weather.OpenMeteo {
	return &weather.OpenMeteo{
		Client:      srv.Client(),
		ForecastURL: srv.URL + "/v1/forecast",
		ArchiveURL:  srv.URL + "/v1/archive",
		Now:         func() time.Time { return now },
	}
}

func hour(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04", s)
	return t.UTC()
}

func TestOpenMeteo_Hours(t *testing.T) {
	t.Run("happy path - recent window uses forecast and skips null hours", func(t *testing.T) {
		// given
		srv, calls := fixtureServer(t, http.StatusOK, "open-meteo-forecast.json")
		o := source(srv, hour("2026-10-13T12:00"))

		// when
		got, err := o.Hours(t.Context(), 44.4812, -73.2149, hour("2026-10-12T00:30"), hour("2026-10-13T00:10"))

		// then
		require.NoError(t, err)
		require.Len(t, *calls, 1)
		c := (*calls)[0]
		assert.Equal(t, "/v1/forecast", c.path)
		assert.Equal(t, "44.48", c.query.Get("latitude"))
		assert.Equal(t, "-73.21", c.query.Get("longitude"))
		assert.Equal(t, "2026-10-12", c.query.Get("start_date"))
		assert.Equal(t, "2026-10-13", c.query.Get("end_date"))
		assert.Equal(t, "pressure_msl,temperature_2m,relative_humidity_2m,precipitation", c.query.Get("hourly"))
		assert.Equal(t, "UTC", c.query.Get("timezone"))
		assert.Equal(t, []model.WeatherHour{
			{LatRound: 44.48, LonRound: -73.21, Hour: hour("2026-10-12T00:00"), PressureHpa: 1018.2,
				TemperatureC: 9.1, HumidityPct: 81, PrecipitationMm: 0, Source: "open-meteo-forecast"},
			{LatRound: 44.48, LonRound: -73.21, Hour: hour("2026-10-12T01:00"), PressureHpa: 1017.9,
				TemperatureC: 8.7, HumidityPct: 83, PrecipitationMm: 0, Source: "open-meteo-forecast"},
			{LatRound: 44.48, LonRound: -73.21, Hour: hour("2026-10-12T02:00"), PressureHpa: 1017.1,
				TemperatureC: 8.2, HumidityPct: 85, PrecipitationMm: 0.1, Source: "open-meteo-forecast"},
			{LatRound: 44.48, LonRound: -73.21, Hour: hour("2026-10-13T00:00"), PressureHpa: 1009.4,
				TemperatureC: 6.5, HumidityPct: 92, PrecipitationMm: 1.2, Source: "open-meteo-forecast"},
		}, got)
	})

	t.Run("happy path - old window uses archive", func(t *testing.T) {
		// given
		srv, calls := fixtureServer(t, http.StatusOK, "open-meteo-archive.json")
		o := source(srv, hour("2026-10-13T12:00"))

		// when
		got, err := o.Hours(t.Context(), 44.48, -73.21, hour("2026-09-22T22:00"), hour("2026-09-23T00:00"))

		// then
		require.NoError(t, err)
		require.Len(t, *calls, 1)
		assert.Equal(t, "/v1/archive", (*calls)[0].path)
		require.Len(t, got, 3)
		assert.Equal(t, "open-meteo-archive", got[0].Source)
	})

	t.Run("happy path - window ending exactly 5 days ago uses forecast", func(t *testing.T) {
		// given
		srv, calls := fixtureServer(t, http.StatusOK, "open-meteo-forecast.json")
		now := hour("2026-10-17T00:00")
		o := source(srv, now)

		// when
		_, err := o.Hours(t.Context(), 44.48, -73.21, hour("2026-10-11T00:00"), now.AddDate(0, 0, -5))

		// then
		require.NoError(t, err)
		assert.Equal(t, "/v1/forecast", (*calls)[0].path)
	})

	t.Run("happy path - hours outside window dropped", func(t *testing.T) {
		// given
		srv, _ := fixtureServer(t, http.StatusOK, "open-meteo-archive.json")
		o := source(srv, hour("2026-10-13T12:00"))

		// when
		got, err := o.Hours(t.Context(), 44.48, -73.21, hour("2026-09-22T23:00"), hour("2026-09-22T23:59"))

		// then
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, hour("2026-09-22T23:00"), got[0].Hour)
	})

	t.Run("error - api error carries reason", func(t *testing.T) {
		// given
		srv, _ := fixtureServer(t, http.StatusBadRequest, "open-meteo-error.json")
		o := source(srv, hour("2026-10-13T12:00"))

		// when
		_, err := o.Hours(t.Context(), 44.48, -73.21, hour("2026-10-12T00:00"), hour("2026-10-13T00:00"))

		// then
		require.ErrorIs(t, err, weather.ErrUpstream)
		assert.Contains(t, err.Error(), "out of allowed range")
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("error - mismatched array lengths", func(t *testing.T) {
		// given
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"hourly":{"time":["2026-10-12T00:00"],"pressure_msl":[],` +
				`"temperature_2m":[1],"relative_humidity_2m":[1],"precipitation":[1]}}`))
		}))
		t.Cleanup(srv.Close)
		o := source(srv, hour("2026-10-13T12:00"))

		// when
		_, err := o.Hours(t.Context(), 44.48, -73.21, hour("2026-10-12T00:00"), hour("2026-10-12T00:00"))

		// then
		assert.ErrorIs(t, err, weather.ErrUpstream)
	})
}
