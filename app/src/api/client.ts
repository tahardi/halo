import { File } from 'expo-file-system';

import {
  isAttack,
  isAttackDetail,
  isAttackPage,
  isHealthResponse,
  isMedication,
  isMedicationList,
  isRecording,
} from './guards';
import type { Attack, AttackDetail, AttackPage, ErrorCode, HealthResponse, Medication, Recording } from './types';

export class ApiError extends Error {
  code: ErrorCode;
  status: number | null;

  constructor(code: ErrorCode, status: number | null, message: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

export type ClientConfig = { baseUrl: string; key: string; fetchFn?: typeof fetch };

export type Client = {
  health(): Promise<HealthResponse>;
  putAttack(attack: Attack): Promise<Attack>;
  listAttacks(before?: string, limit?: number): Promise<AttackPage>;
  getAttack(id: string): Promise<AttackDetail>;
  deleteAttack(id: string): Promise<void>;
  listMedications(): Promise<Medication[]>;
  putMedication(m: Medication): Promise<Medication>;
  uploadRecording(attackId: string, id: string, recordedAt: string, fileUri: string): Promise<Recording>;
  retryRecording(id: string): Promise<Recording>;
  confirmRecording(id: string): Promise<Recording>;
};

const HEALTH_MS = 10_000;
const DEFAULT_MS = 30_000;
const RECORDING_MS = 120_000;

const SERVER_CODES: ErrorCode[] = [
  'bad_request',
  'unauthorized',
  'not_found',
  'payload_too_large',
  'no_speech',
  'upstream_failed',
];

const STATUS_CODES: Record<number, ErrorCode> = {
  400: 'bad_request',
  401: 'unauthorized',
  404: 'not_found',
  413: 'payload_too_large',
  422: 'no_speech',
};

function errorBody(body: unknown): { code: ErrorCode; message: string } | null {
  const err = (body as { error?: { code?: unknown; message?: unknown } } | null)?.error;
  if (typeof err?.code !== 'string' || typeof err.message !== 'string') {
    return null;
  }
  if (!SERVER_CODES.includes(err.code as ErrorCode)) {
    return null;
  }
  return { code: err.code as ErrorCode, message: err.message };
}

function isEmpty(value: unknown): value is null {
  return value === null;
}

function json(method: string, body: unknown): RequestInit {
  return { method, body: JSON.stringify(body), headers: { 'Content-Type': 'application/json' } };
}

export function messageFor(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return 'Backend error';
  }
  switch (error.code) {
    case 'network':
      return "Can't reach backend";
    case 'timeout':
      return 'Backend timed out';
    case 'unauthorized':
      return 'Wrong key';
    case 'no_speech':
      return 'No speech found in the recording';
    case 'payload_too_large':
      return 'Recording is too large';
    case 'not_found':
      return 'Not found on the backend';
    case 'bad_request':
      return `Rejected: ${error.message}`;
    default:
      return 'Backend error';
  }
}

export function createClient(config: ClientConfig): Client {
  const baseUrl = config.baseUrl.trim().replace(/\/+$/, '');
  const fetchFn = config.fetchFn ?? fetch;
  const enc = encodeURIComponent;

  async function request<T>(
    endpoint: string,
    init: RequestInit,
    timeoutMs: number,
    isValid: (value: unknown) => value is T,
  ): Promise<T> {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), timeoutMs);
    let res: Response;
    try {
      res = await fetchFn(baseUrl + endpoint, {
        ...init,
        headers: { ...(init.headers as Record<string, string>), Authorization: `Bearer ${config.key}` },
        signal: controller.signal,
      });
    } catch {
      throw controller.signal.aborted
        ? new ApiError('timeout', null, 'backend timed out')
        : new ApiError('network', null, "can't reach backend");
    } finally {
      clearTimeout(timer);
    }
    const body: unknown = res.status === 204 ? null : await res.json().catch(() => null);
    if (!res.ok) {
      const err = errorBody(body);
      throw new ApiError(
        err?.code ?? STATUS_CODES[res.status] ?? 'unknown',
        res.status,
        err?.message ?? `backend returned ${res.status}`,
      );
    }
    if (!isValid(body)) {
      throw new ApiError('unknown', res.status, 'invalid response');
    }
    return body;
  }

  return {
    health: () => request('/health', { method: 'GET' }, HEALTH_MS, isHealthResponse),
    putAttack: (attack) => request(`/attacks/${enc(attack.id)}`, json('PUT', attack), DEFAULT_MS, isAttack),
    listAttacks: (before, limit) => {
      const params = new URLSearchParams();
      if (before !== undefined) {
        params.set('before', before);
      }
      if (limit !== undefined) {
        params.set('limit', String(limit));
      }
      const query = params.toString();
      return request(`/attacks${query ? `?${query}` : ''}`, { method: 'GET' }, DEFAULT_MS, isAttackPage);
    },
    getAttack: (id) => request(`/attacks/${enc(id)}`, { method: 'GET' }, DEFAULT_MS, isAttackDetail),
    deleteAttack: async (id) => {
      await request(`/attacks/${enc(id)}`, { method: 'DELETE' }, DEFAULT_MS, isEmpty);
    },
    listMedications: async () =>
      (await request('/medications', { method: 'GET' }, DEFAULT_MS, isMedicationList)).medications,
    putMedication: (m) => request(`/medications/${enc(m.id)}`, json('PUT', m), DEFAULT_MS, isMedication),
    uploadRecording: (attackId, id, recordedAt, fileUri) => {
      const form = new FormData();
      form.append('id', id);
      form.append('recorded_at', recordedAt);
      form.append('audio', new File(fileUri), 'summary.m4a');
      return request(`/attacks/${enc(attackId)}/recordings`, { method: 'POST', body: form }, RECORDING_MS, isRecording);
    },
    retryRecording: (id) => request(`/recordings/${enc(id)}/retry`, { method: 'POST' }, RECORDING_MS, isRecording),
    confirmRecording: (id) => request(`/recordings/${enc(id)}/confirm`, { method: 'POST' }, DEFAULT_MS, isRecording),
  };
}
