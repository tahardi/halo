package model

import "slices"

type Symptom string

const (
	SymptomAura             Symptom = "aura"
	SymptomNausea           Symptom = "nausea"
	SymptomVomiting         Symptom = "vomiting"
	SymptomLightSensitivity Symptom = "light_sensitivity"
	SymptomSoundSensitivity Symptom = "sound_sensitivity"
	SymptomOneSided         Symptom = "one_sided"
	SymptomThrobbing        Symptom = "throbbing"
	SymptomNeckPain         Symptom = "neck_pain"
	SymptomDizziness        Symptom = "dizziness"
	SymptomBrainFog         Symptom = "brain_fog"
)

var Symptoms = []Symptom{
	SymptomAura, SymptomNausea, SymptomVomiting, SymptomLightSensitivity, SymptomSoundSensitivity,
	SymptomOneSided, SymptomThrobbing, SymptomNeckPain, SymptomDizziness, SymptomBrainFog,
}

func (s Symptom) Valid() bool {
	return slices.Contains(Symptoms, s)
}
