import * as SecureStore from 'expo-secure-store';

import { loadSettings, saveSettings } from '@/settings/store';

jest.mock('expo-secure-store', () => {
  const values = new Map<string, string>();
  return {
    getItemAsync: jest.fn(async (k: string) => values.get(k) ?? null),
    setItemAsync: jest.fn(async (k: string, v: string) => {
      values.set(k, v);
    }),
    __values: values,
  };
});

const values = (SecureStore as unknown as { __values: Map<string, string> }).__values;

describe('loadSettings', () => {
  beforeEach(() => values.clear());

  test('happy path - empty store returns defaults', async () => {
    // when
    const s = await loadSettings();

    // then
    expect(s).toEqual({ baseUrl: '', key: '' });
  });

  test('error - corrupt JSON returns defaults', async () => {
    // given
    values.set('settings', '{not json');

    // when
    const s = await loadSettings();

    // then
    expect(s).toEqual({ baseUrl: '', key: '' });
  });
});

describe('saveSettings', () => {
  beforeEach(() => values.clear());

  test('happy path - trims the key and trailing slashes, then round-trips', async () => {
    // when
    await saveSettings({ baseUrl: ' https://pc.ts.net:8443// ', key: ' abc ' });

    // then
    expect(await loadSettings()).toEqual({ baseUrl: 'https://pc.ts.net:8443', key: 'abc' });
  });

  test('error - rejects a URL without a scheme', async () => {
    // when
    const save = saveSettings({ baseUrl: 'pc.ts.net', key: 'k' });

    // then
    await expect(save).rejects.toThrow('URL must start with https:// or http://');
  });
});
