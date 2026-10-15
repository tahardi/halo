export type Symptom =
  | 'aura'
  | 'nausea'
  | 'vomiting'
  | 'light_sensitivity'
  | 'sound_sensitivity'
  | 'one_sided'
  | 'throbbing'
  | 'neck_pain'
  | 'dizziness'
  | 'brain_fog';

export const SYMPTOMS: Symptom[] = [
  'aura',
  'nausea',
  'vomiting',
  'light_sensitivity',
  'sound_sensitivity',
  'one_sided',
  'throbbing',
  'neck_pain',
  'dizziness',
  'brain_fog',
];

export type Relief = 'none' | 'some' | 'full';

export type Dose = {
  id: string;
  medication_id: string;
  amount: number;
  taken_at: string;
  relief: Relief | null;
};

export type Attack = {
  id: string;
  started_at: string;
  ended_at: string | null;
  timezone: string;
  peak_severity: number | null;
  symptoms: Symptom[];
  symptoms_other: string | null;
  notes: string | null;
  lat: number | null;
  lon: number | null;
  source: 'app' | 'import';
  doses: Dose[];
  created_at: string;
  updated_at: string;
};

export type AttackPage = { attacks: Attack[]; next_before: string | null };

export type MedicationKind = 'acute' | 'preventive';

export type Medication = {
  id: string;
  name: string;
  usual_amount: number;
  unit: string;
  kind: MedicationKind;
  archived: boolean;
};

export type RecordingStatus = 'pending' | 'transcribed' | 'extracted' | 'confirmed' | 'failed';

export type ProposedDose = { medication_id: string; amount: number; taken_at: string };

export type UnknownMedication = { name: string; amount: number | null; taken_at: string | null };

export type Proposal = {
  started_at: string | null;
  ended_at: string | null;
  peak_severity: number | null;
  symptoms: Symptom[];
  doses: ProposedDose[];
  unknown_medications: UnknownMedication[];
  notes: string | null;
  warnings: string[];
};

export type Recording = {
  id: string;
  attack_id: string;
  recorded_at: string;
  status: RecordingStatus;
  error: string | null;
  transcript: string | null;
  proposal: Proposal | null;
  created_at: string;
};

export type WeatherSummary = {
  hours_recorded: number;
  hours_expected: number;
  pressure_at_start_hpa: number | null;
  pressure_change_24h_hpa: number | null;
  temperature_at_start_c: number | null;
  humidity_at_start_pct: number | null;
};

export type AttackDetail = { attack: Attack; recordings: Recording[]; weather: WeatherSummary | null };

export type HealthResponse = { status: 'ok' };

export type ErrorCode =
  | 'bad_request'
  | 'unauthorized'
  | 'not_found'
  | 'payload_too_large'
  | 'no_speech'
  | 'upstream_failed'
  | 'internal'
  | 'network'
  | 'timeout'
  | 'unknown';
