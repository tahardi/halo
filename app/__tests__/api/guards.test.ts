import fs from 'node:fs';
import path from 'node:path';

import {
  isAttack,
  isAttackDetail,
  isAttackPage,
  isErrorResponse,
  isHealthResponse,
  isMedicationList,
  isRecording,
} from '@/api/guards';

function fixture(file: string): Record<string, unknown> {
  return JSON.parse(fs.readFileSync(path.join(__dirname, '../../../api/testdata', file), 'utf8'));
}

const attack = fixture('attack.json');

describe('isHealthResponse', () => {
  it.each([
    [fixture('health-response.json'), true],
    [{ status: 3 }, false],
    [null, false],
  ])('%j is %s', (value, want) => {
    expect(isHealthResponse(value)).toBe(want);
  });
});

describe('isAttack', () => {
  it.each([
    ['fixture', attack, true],
    ['active attack', { ...attack, ended_at: null, peak_severity: null }, true],
    ['missing id', { ...attack, id: undefined }, false],
    ['severity as string', { ...attack, peak_severity: '8' }, false],
    ['unknown symptom', { ...attack, symptoms: ['headache'] }, false],
    ['doses not a list', { ...attack, doses: null }, false],
    ['dose missing taken_at', { ...attack, doses: [{ id: 'd', medication_id: 'm', amount: 1, relief: null }] }, false],
    ['array', [attack], false],
  ])('%s is %s', (_name, value, want) => {
    expect(isAttack(value)).toBe(want);
  });
});

describe('isAttackPage', () => {
  it.each([
    [fixture('attack-page.json'), true],
    [{ attacks: [], next_before: null }, true],
    [{ attacks: [{}], next_before: null }, false],
    [{ attacks: [] }, false],
  ])('%j is %s', (value, want) => {
    expect(isAttackPage(value)).toBe(want);
  });
});

describe('isAttackDetail', () => {
  const detail = fixture('attack-detail.json');
  it.each([
    ['fixture', detail, true],
    ['no weather yet', { ...detail, weather: null }, true],
    ['bad recording', { ...detail, recordings: [{ id: 'r' }] }, false],
    ['missing attack', { ...detail, attack: null }, false],
  ])('%s is %s', (_name, value, want) => {
    expect(isAttackDetail(value)).toBe(want);
  });
});

describe('isMedicationList', () => {
  it.each([
    [fixture('medication-list.json'), true],
    [{ medications: [] }, true],
    [{ medications: [{ id: 'x', name: 'X', usual_amount: 1, unit: 'tablet', kind: 'daily', archived: false }] }, false],
    [[], false],
  ])('%j is %s', (value, want) => {
    expect(isMedicationList(value)).toBe(want);
  });
});

describe('isRecording', () => {
  it.each([
    [fixture('recording-extracted.json'), true],
    [fixture('recording-failed.json'), true],
    [{ ...fixture('recording-failed.json'), status: 'done' }, false],
  ])('%j is %s', (value, want) => {
    expect(isRecording(value)).toBe(want);
  });
});

describe('isErrorResponse', () => {
  it.each([
    [fixture('error-not-found.json'), true],
    [fixture('error-no-speech.json'), true],
    [{ error: 'nope' }, false],
  ])('%j is %s', (value, want) => {
    expect(isErrorResponse(value)).toBe(want);
  });
});
