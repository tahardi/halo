package model

import "time"

type RecordingStatus string

const (
	RecordingPending     RecordingStatus = "pending"
	RecordingTranscribed RecordingStatus = "transcribed"
	RecordingExtracted   RecordingStatus = "extracted"
	RecordingConfirmed   RecordingStatus = "confirmed"
	RecordingFailed      RecordingStatus = "failed"
)

type Recording struct {
	ID         string          `json:"id"`
	AttackID   string          `json:"attack_id"`
	RecordedAt time.Time       `json:"recorded_at"`
	Status     RecordingStatus `json:"status"`
	Error      *string         `json:"error"`
	Transcript *string         `json:"transcript"`
	Proposal   *Proposal       `json:"proposal"`
	CreatedAt  time.Time       `json:"created_at"`
}
