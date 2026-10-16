import { SYMPTOMS } from './types';
import type {
  Attack,
  AttackDetail,
  AttackPage,
  HealthResponse,
  Medication,
  Recording,
  RecordingStatus,
  Symptom,
} from './types';

type Obj = Record<string, unknown>;

const STATUSES: RecordingStatus[] = ['pending', 'transcribed', 'extracted', 'confirmed', 'failed'];

function isObject(value: unknown): value is Obj {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isString(value: unknown): value is string {
  return typeof value === 'string';
}

function isNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value);
}

function isNullable<T>(value: unknown, check: (v: unknown) => v is T): boolean {
  return value === null || check(value);
}

function isList(value: unknown, check: (v: unknown) => boolean): boolean {
  return Array.isArray(value) && value.every(check);
}

function isSymptom(value: unknown): value is Symptom {
  return SYMPTOMS.includes(value as Symptom);
}

function isDose(value: unknown): boolean {
  return (
    isObject(value) &&
    isString(value.id) &&
    isString(value.medication_id) &&
    isNumber(value.amount) &&
    isString(value.taken_at) &&
    (value.relief === null || ['none', 'some', 'full'].includes(value.relief as string))
  );
}

function isProposal(value: unknown): boolean {
  return (
    isObject(value) &&
    isNullable(value.started_at, isString) &&
    isNullable(value.ended_at, isString) &&
    isNullable(value.peak_severity, isNumber) &&
    isList(value.symptoms, isSymptom) &&
    isList(
      value.doses,
      (d) => isObject(d) && isString(d.medication_id) && isNumber(d.amount) && isString(d.taken_at),
    ) &&
    isList(
      value.unknown_medications,
      (u) => isObject(u) && isString(u.name) && isNullable(u.amount, isNumber) && isNullable(u.taken_at, isString),
    ) &&
    isNullable(value.notes, isString) &&
    isList(value.warnings, isString)
  );
}

function isWeather(value: unknown): boolean {
  return (
    isObject(value) &&
    isNumber(value.hours_recorded) &&
    isNumber(value.hours_expected) &&
    isNullable(value.pressure_at_start_hpa, isNumber) &&
    isNullable(value.pressure_change_24h_hpa, isNumber) &&
    isNullable(value.temperature_at_start_c, isNumber) &&
    isNullable(value.humidity_at_start_pct, isNumber)
  );
}

export function isHealthResponse(value: unknown): value is HealthResponse {
  return isObject(value) && value.status === 'ok';
}

export function isAttack(value: unknown): value is Attack {
  return (
    isObject(value) &&
    isString(value.id) &&
    isString(value.started_at) &&
    isNullable(value.ended_at, isString) &&
    isString(value.timezone) &&
    isNullable(value.peak_severity, isNumber) &&
    isList(value.symptoms, isSymptom) &&
    isNullable(value.symptoms_other, isString) &&
    isNullable(value.notes, isString) &&
    isNullable(value.lat, isNumber) &&
    isNullable(value.lon, isNumber) &&
    (value.source === 'app' || value.source === 'import') &&
    isList(value.doses, isDose) &&
    isString(value.created_at) &&
    isString(value.updated_at)
  );
}

export function isAttackPage(value: unknown): value is AttackPage {
  return isObject(value) && isList(value.attacks, isAttack) && isNullable(value.next_before, isString);
}

export function isMedication(value: unknown): value is Medication {
  return (
    isObject(value) &&
    isString(value.id) &&
    isString(value.name) &&
    isNumber(value.usual_amount) &&
    isString(value.unit) &&
    (value.kind === 'acute' || value.kind === 'preventive') &&
    typeof value.archived === 'boolean'
  );
}

export function isMedicationList(value: unknown): value is { medications: Medication[] } {
  return isObject(value) && isList(value.medications, isMedication);
}

export function isRecording(value: unknown): value is Recording {
  return (
    isObject(value) &&
    isString(value.id) &&
    isString(value.attack_id) &&
    isString(value.recorded_at) &&
    STATUSES.includes(value.status as RecordingStatus) &&
    isNullable(value.error, isString) &&
    isNullable(value.transcript, isString) &&
    (value.proposal === null || isProposal(value.proposal)) &&
    isString(value.created_at)
  );
}

export function isAttackDetail(value: unknown): value is AttackDetail {
  return (
    isObject(value) &&
    isAttack(value.attack) &&
    isList(value.recordings, isRecording) &&
    (value.weather === null || isWeather(value.weather))
  );
}

export function isErrorResponse(value: unknown): value is { error: { code: string; message: string } } {
  return isObject(value) && isObject(value.error) && isString(value.error.code) && isString(value.error.message);
}
