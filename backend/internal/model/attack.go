package model

import "time"

type Attack struct {
	ID            string     `json:"id"`
	StartedAt     time.Time  `json:"started_at"`
	EndedAt       *time.Time `json:"ended_at"`
	Timezone      string     `json:"timezone"`
	PeakSeverity  *int       `json:"peak_severity"`
	Symptoms      []Symptom  `json:"symptoms"`
	SymptomsOther *string    `json:"symptoms_other"`
	Notes         *string    `json:"notes"`
	Lat           *float64   `json:"lat"`
	Lon           *float64   `json:"lon"`
	Source        string     `json:"source"`
	Doses         []Dose     `json:"doses"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AttackPage struct {
	Attacks    []Attack   `json:"attacks"`
	NextBefore *time.Time `json:"next_before"`
}

type AttackDetail struct {
	Attack     Attack          `json:"attack"`
	Recordings []Recording     `json:"recordings"`
	Weather    *WeatherSummary `json:"weather"`
}
