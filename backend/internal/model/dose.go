package model

import "time"

type Relief string

const (
	ReliefNone Relief = "none"
	ReliefSome Relief = "some"
	ReliefFull Relief = "full"
)

type Dose struct {
	ID           string    `json:"id"`
	MedicationID string    `json:"medication_id"`
	Amount       float64   `json:"amount"`
	TakenAt      time.Time `json:"taken_at"`
	Relief       *Relief   `json:"relief"`
}
