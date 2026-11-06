import { configuredClient, isConfigured } from '@/api/configured';
import { loadSettings } from '@/settings/store';

jest.mock('@/settings/store', () => ({ loadSettings: jest.fn() }));

const mockLoad = loadSettings as jest.MockedFunction<typeof loadSettings>;

describe('isConfigured', () => {
  test.each([
    [{ baseUrl: '', key: '' }, false],
    [{ baseUrl: 'https://pc', key: '' }, false],
    [{ baseUrl: '', key: 'k' }, false],
    [{ baseUrl: 'https://pc', key: 'k' }, true],
  ])('%p is %p', (settings, want) => {
    expect(isConfigured(settings)).toBe(want);
  });
});

describe('configuredClient', () => {
  test('happy path - returns a client when configured', async () => {
    // given
    mockLoad.mockResolvedValue({ baseUrl: 'https://pc', key: 'k' });

    // when
    const client = await configuredClient();

    // then
    expect(client).not.toBeNull();
  });

  test('error - returns null when not configured', async () => {
    // given
    mockLoad.mockResolvedValue({ baseUrl: '', key: '' });

    // when
    const client = await configuredClient();

    // then
    expect(client).toBeNull();
  });
});
