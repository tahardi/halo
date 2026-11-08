import fs from 'node:fs';
import path from 'node:path';

import { ApiError, createClient, messageFor } from '@/api/client';

jest.mock('expo-file-system', () => ({
  File: class extends Blob {
    uri: string;
    constructor(uri: string) {
      super(['audio-bytes'], { type: 'audio/m4a' });
      this.uri = uri;
    }
  },
}));

function fixture(file: string): string {
  return fs.readFileSync(path.join(__dirname, '../../../api/testdata', file), 'utf8');
}

function respond(status: number, body: string | null): jest.Mock {
  return jest.fn().mockResolvedValue(new Response(body, { status }));
}

function client(fetchFn: jest.Mock) {
  return createClient({ baseUrl: 'https://pc.ts.net/', key: 'k', fetchFn: fetchFn as unknown as typeof fetch });
}

function formValue(form: FormData, name: string): unknown {
  if (typeof form.get === 'function') {
    return form.get(name);
  }
  const parts = (form as unknown as { getParts(): { fieldName: string; string?: string }[] }).getParts();
  const part = parts.find((p) => p.fieldName === name);
  return part ? (part.string ?? part) : null;
}

async function codeOf(promise: Promise<unknown>): Promise<string> {
  try {
    await promise;
  } catch (e) {
    return e instanceof ApiError ? e.code : 'not-api-error';
  }
  return 'resolved';
}

describe('createClient', () => {
  afterEach(() => {
    jest.useRealTimers();
  });

  test('happy path - health sends the bearer key and strips the trailing slash', async () => {
    // given
    const fetchFn = respond(200, fixture('health-response.json'));

    // when
    const res = await client(fetchFn).health();

    // then
    expect(res).toEqual({ status: 'ok' });
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe('https://pc.ts.net/health');
    expect(init.headers.Authorization).toBe('Bearer k');
  });

  test('happy path - putAttack sends JSON to the attack id', async () => {
    // given
    const body = fixture('attack.json');
    const attack = JSON.parse(body);
    const fetchFn = respond(200, body);

    // when
    const res = await client(fetchFn).putAttack(attack);

    // then
    expect(res).toEqual(attack);
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe(`https://pc.ts.net/attacks/${attack.id}`);
    expect(init.method).toBe('PUT');
    expect(init.headers['Content-Type']).toBe('application/json');
    expect(JSON.parse(init.body)).toEqual(attack);
  });

  test('happy path - listAttacks passes before and limit', async () => {
    // given
    const fetchFn = respond(200, fixture('attack-page.json'));

    // when
    await client(fetchFn).listAttacks('2026-10-13T07:00:00Z', 20);

    // then
    expect(fetchFn.mock.calls[0][0]).toBe('https://pc.ts.net/attacks?before=2026-10-13T07%3A00%3A00Z&limit=20');
  });

  test('happy path - listAttacks without arguments sends no query', async () => {
    // given
    const fetchFn = respond(200, fixture('attack-page.json'));

    // when
    await client(fetchFn).listAttacks();

    // then
    expect(fetchFn.mock.calls[0][0]).toBe('https://pc.ts.net/attacks');
  });

  test('happy path - deleteAttack accepts 204', async () => {
    // given
    const fetchFn = respond(204, null);

    // when
    const res = await client(fetchFn).deleteAttack('a-1');

    // then
    expect(res).toBeUndefined();
    expect(fetchFn.mock.calls[0][1].method).toBe('DELETE');
  });

  test('happy path - listMedications unwraps the list', async () => {
    // given
    const body = fixture('medication-list.json');
    const fetchFn = respond(200, body);

    // when
    const meds = await client(fetchFn).listMedications();

    // then
    expect(meds).toEqual(JSON.parse(body).medications);
  });

  test('happy path - uploadRecording sends id, recorded_at, and audio', async () => {
    // given
    const fetchFn = respond(200, fixture('recording-extracted.json'));

    // when
    await client(fetchFn).uploadRecording('a-1', 'r-1', '2026-10-13T11:30:00-04:00', 'file:///clip.m4a');

    // then
    const [url, init] = fetchFn.mock.calls[0];
    expect(url).toBe('https://pc.ts.net/attacks/a-1/recordings');
    expect(init.method).toBe('POST');
    expect(formValue(init.body, 'id')).toBe('r-1');
    expect(formValue(init.body, 'recorded_at')).toBe('2026-10-13T11:30:00-04:00');
    expect(formValue(init.body, 'audio')).not.toBeNull();
  });

  test('happy path - retry and confirm post to the recording', async () => {
    // given
    const fetchFn = jest.fn().mockImplementation(async () => new Response(fixture('recording-extracted.json')));

    // when
    await client(fetchFn).retryRecording('r-1');
    await client(fetchFn).confirmRecording('r-1');

    // then
    expect(fetchFn.mock.calls.map((c) => [c[0], c[1].method])).toEqual([
      ['https://pc.ts.net/recordings/r-1/retry', 'POST'],
      ['https://pc.ts.net/recordings/r-1/confirm', 'POST'],
    ]);
  });

  test.each([
    ['backend error body', 401, fixture('error-unauthorized.json'), 'unauthorized'],
    ['bare 400', 400, null, 'bad_request'],
    ['bare 404', 404, null, 'not_found'],
    ['bare 413', 413, null, 'payload_too_large'],
    ['bare 422', 422, null, 'no_speech'],
    ['bare 502 from the proxy', 502, null, 'unknown'],
    ['backend 502 body', 502, '{"error":{"code":"upstream_failed","message":"whisper"}}', 'upstream_failed'],
    ['bare 500', 500, null, 'unknown'],
    ['unknown code in body', 400, '{"error":{"code":"weird","message":"x"}}', 'bad_request'],
    ['2xx body fails guard', 200, '{"nope":true}', 'unknown'],
  ])('error - %s maps to %s', async (_name, status, body, code) => {
    // given
    const fetchFn = respond(status, body);

    // when
    const got = await codeOf(client(fetchFn).health());

    // then
    expect(got).toBe(code);
  });

  test('error - fetch rejection is network', async () => {
    // given
    const fetchFn = jest.fn().mockRejectedValue(new TypeError('Network request failed'));

    // when
    const got = await codeOf(client(fetchFn).health());

    // then
    expect(got).toBe('network');
  });

  test.each([
    ['health', 10_000, (c: ReturnType<typeof client>) => c.health()],
    ['getAttack', 30_000, (c: ReturnType<typeof client>) => c.getAttack('a-1')],
    ['uploadRecording', 120_000, (c: ReturnType<typeof client>) => c.uploadRecording('a', 'r', 't', 'file:///x')],
    ['retryRecording', 120_000, (c: ReturnType<typeof client>) => c.retryRecording('r')],
  ])('error - %s times out after %i ms', async (_name, ms, call) => {
    // given
    jest.useFakeTimers();
    const fetchFn = jest
      .fn()
      .mockImplementation(
        (_url: string, init: RequestInit) =>
          new Promise((_resolve, reject) => init.signal?.addEventListener('abort', () => reject(new Error('aborted')))),
      );
    const pending = codeOf(call(client(fetchFn)));

    // when
    jest.advanceTimersByTime(ms - 1);
    await Promise.resolve();
    const early = fetchFn.mock.calls[0][1].signal.aborted;
    jest.advanceTimersByTime(1);

    // then
    expect(early).toBe(false);
    expect(await pending).toBe('timeout');
  });
});

describe('messageFor', () => {
  test.each([
    [new ApiError('network', null, 'x'), "Can't reach backend"],
    [new ApiError('timeout', null, 'x'), 'Backend timed out'],
    [new ApiError('unauthorized', 401, 'x'), 'Wrong key'],
    [new ApiError('no_speech', 422, 'x'), 'No speech found in the recording'],
    [new ApiError('payload_too_large', 413, 'x'), 'Recording is too large'],
    [new ApiError('not_found', 404, 'x'), 'Not found on the backend'],
    [new ApiError('bad_request', 400, 'severity out of range'), 'Rejected: severity out of range'],
    [new Error('boom'), 'Backend error'],
  ])('maps %p', (error, text) => {
    expect(messageFor(error)).toBe(text);
  });
});
