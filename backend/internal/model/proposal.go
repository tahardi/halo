package model

import "time"

type ProposedDose struct {
	MedicationID string    `json:"medication_id"`
	Amount       float64   `json:"amount"`
	TakenAt      time.Time `json:"taken_at"`
}

type UnknownMedication struct {
	Name    string     `json:"name"`
	Amount  *float64   `json:"amount"`
	TakenAt *time.Time `json:"taken_at"`
}

type Proposal struct {
	StartedAt          *time.Time          `json:"started_at"`
	EndedAt            *time.Time          `json:"ended_at"`
	PeakSeverity       *int                `json:"peak_severity"`
	Symptoms           []Symptom           `json:"symptoms"`
	Doses              []ProposedDose      `json:"doses"`
	UnknownMedications []UnknownMedication `json:"unknown_medications"`
	Notes              *string             `json:"notes"`
	Warnings           []string            `json:"warnings"`
}
