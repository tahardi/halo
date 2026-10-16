package model

type MedicationKind string

const (
	MedicationAcute      MedicationKind = "acute"
	MedicationPreventive MedicationKind = "preventive"
)

type Medication struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	UsualAmount float64        `json:"usual_amount"`
	Unit        string         `json:"unit"`
	Kind        MedicationKind `json:"kind"`
	Archived    bool           `json:"archived"`
}

type MedicationList struct {
	Medications []Medication `json:"medications"`
}
