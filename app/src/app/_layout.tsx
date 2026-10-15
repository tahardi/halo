import { Stack } from 'expo-router';

export default function RootLayout() {
  return (
    <Stack screenOptions={{ headerStyle: { backgroundColor: '#000000' }, headerTintColor: '#B0B0B0' }}>
      <Stack.Screen name="index" options={{ title: 'Halo' }} />
    </Stack>
  );
}
