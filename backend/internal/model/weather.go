package model

import "time"

type WeatherHour struct {
	LatRound        float64   `json:"lat_round"`
	LonRound        float64   `json:"lon_round"`
	Hour            time.Time `json:"hour"`
	PressureHpa     float64   `json:"pressure_hpa"`
	TemperatureC    float64   `json:"temperature_c"`
	HumidityPct     float64   `json:"humidity_pct"`
	PrecipitationMm float64   `json:"precipitation_mm"`
	Source          string    `json:"source"`
}

type WeatherSummary struct {
	HoursRecorded        int      `json:"hours_recorded"`
	HoursExpected        int      `json:"hours_expected"`
	PressureAtStartHpa   *float64 `json:"pressure_at_start_hpa"`
	PressureChange24hHpa *float64 `json:"pressure_change_24h_hpa"`
	TemperatureAtStartC  *float64 `json:"temperature_at_start_c"`
	HumidityAtStartPct   *float64 `json:"humidity_at_start_pct"`
}
