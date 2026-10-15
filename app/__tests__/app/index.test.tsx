import { render, screen } from '@testing-library/react-native';

import Index from '@/app/index';

describe('Index', () => {
  test('shows the app name', async () => {
    await render(<Index />);

    expect(screen.getByTestId('app-title')).toHaveTextContent('Halo');
  });
});
