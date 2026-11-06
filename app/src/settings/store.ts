import * as SecureStore from 'expo-secure-store';

export type Settings = { baseUrl: string; key: string };

export const DEFAULT_SETTINGS: Settings = { baseUrl: '', key: '' };

const STORE_KEY = 'settings';

export async function loadSettings(): Promise<Settings> {
  const raw = await SecureStore.getItemAsync(STORE_KEY);
  if (!raw) {
    return { ...DEFAULT_SETTINGS };
  }
  try {
    return { ...DEFAULT_SETTINGS, ...(JSON.parse(raw) as Partial<Settings>) };
  } catch {
    return { ...DEFAULT_SETTINGS };
  }
}

export async function saveSettings(settings: Settings): Promise<void> {
  const next = { baseUrl: settings.baseUrl.trim().replace(/\/+$/, ''), key: settings.key.trim() };
  if (next.baseUrl !== '' && !/^https?:\/\//.test(next.baseUrl)) {
    throw new Error('URL must start with https:// or http://');
  }
  await SecureStore.setItemAsync(STORE_KEY, JSON.stringify(next));
}
