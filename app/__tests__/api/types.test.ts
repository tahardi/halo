import fs from 'node:fs';
import path from 'node:path';

import { SYMPTOMS } from '@/api/types';
import type { Attack, AttackDetail, AttackPage, HealthResponse, Medication, Recording } from '@/api/types';

function fixture<T>(file: string): T {
  return JSON.parse(fs.readFileSync(path.join(__dirname, '../../../api/testdata', file), 'utf8')) as T;
}

describe('fixtures', () => {
  test('load as their types', () => {
    const health = fixture<HealthResponse>('health-response.json');
    const attack = fixture<Attack>('attack.json');
    const page = fixture<AttackPage>('attack-page.json');
    const detail = fixture<AttackDetail>('attack-detail.json');
    const meds = fixture<{ medications: Medication[] }>('medication-list.json');
    const recording = fixture<Recording>('recording-extracted.json');

    expect(health.status).toBe('ok');
    expect(attack.doses).toHaveLength(2);
    expect(page.attacks[0].id).toBe(attack.id);
    expect(detail.recordings[0].proposal?.doses).toHaveLength(2);
    expect(meds.medications[0].kind).toBe('acute');
    expect(recording.status).toBe('extracted');
  });
});

describe('SYMPTOMS', () => {
  test('lists the ten backend symptoms once each', () => {
    expect(SYMPTOMS).toEqual([
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
    ]);
  });
});
