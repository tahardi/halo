import { loadSettings, type Settings } from '@/settings/store';

import { type Client, createClient } from './client';

export function isConfigured(settings: Settings): boolean {
  return settings.baseUrl !== '' && settings.key !== '';
}

export async function configuredClient(): Promise<Client | null> {
  const settings = await loadSettings();
  return isConfigured(settings) ? createClient(settings) : null;
}
