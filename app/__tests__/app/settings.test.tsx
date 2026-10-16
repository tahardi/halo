import { fireEvent, render, screen } from '@testing-library/react-native';

import SettingsScreen from '@/app/settings';
import { ApiError, createClient } from '@/api/client';
import { loadSettings, saveSettings } from '@/settings/store';

jest.mock('expo-router', () => ({ Stack: { Screen: () => null } }));
jest.mock('@/settings/store', () => ({ loadSettings: jest.fn(), saveSettings: jest.fn() }));
jest.mock('@/api/client', () => ({ ...jest.requireActual('@/api/client'), createClient: jest.fn() }));

const mockLoad = loadSettings as jest.MockedFunction<typeof loadSettings>;
const mockSave = saveSettings as jest.MockedFunction<typeof saveSettings>;
const mockCreateClient = createClient as jest.MockedFunction<typeof createClient>;
const mockHealth = jest.fn();

async function renderScreen() {
  await render(<SettingsScreen />);
  await screen.findByDisplayValue('https://halo.example.com');
}

describe('SettingsScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockLoad.mockResolvedValue({ baseUrl: 'https://halo.example.com', key: 'secret' });
    mockSave.mockResolvedValue(undefined);
    mockCreateClient.mockReturnValue({ health: mockHealth } as unknown as ReturnType<typeof createClient>);
  });

  test('happy path - prefills the inputs and masks the key', async () => {
    // when
    await renderScreen();

    // then
    expect(screen.getByTestId('settings-key').props.secureTextEntry).toBe(true);
  });

  test('happy path - saves the inputs and shows Saved', async () => {
    // given
    await renderScreen();

    // when
    await fireEvent.changeText(screen.getByTestId('settings-url'), 'https://other.example.com');
    await fireEvent.press(screen.getByTestId('settings-save'));

    // then
    expect(mockSave).toHaveBeenCalledWith({ baseUrl: 'https://other.example.com', key: 'secret' });
    expect(await screen.findByText('Saved')).toBeOnTheScreen();
  });

  test('error - save failure shows the message', async () => {
    // given
    mockSave.mockRejectedValue(new Error('URL must start with https:// or http://'));
    await renderScreen();

    // when
    await fireEvent.press(screen.getByTestId('settings-save'));

    // then
    expect(await screen.findByText('URL must start with https:// or http://')).toBeOnTheScreen();
  });

  test.each([
    ['connected', () => mockHealth.mockResolvedValue({ status: 'ok' }), 'Connected'],
    ['wrong key', () => mockHealth.mockRejectedValue(new ApiError('unauthorized', 401, 'x')), 'Wrong key'],
    ['unreachable', () => mockHealth.mockRejectedValue(new ApiError('network', null, 'x')), "Can't reach backend"],
    ['timeout', () => mockHealth.mockRejectedValue(new ApiError('timeout', null, 'x')), 'Backend timed out'],
  ])('Test connection - %s', async (_name, arrange, text) => {
    // given
    arrange();
    await renderScreen();

    // when
    await fireEvent.press(screen.getByTestId('settings-test'));

    // then
    expect(await screen.findByText(text)).toBeOnTheScreen();
    expect(mockCreateClient).toHaveBeenCalledWith({ baseUrl: 'https://halo.example.com', key: 'secret' });
  });
});
