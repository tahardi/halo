import { router, Stack } from 'expo-router';
import { Pressable, Text } from 'react-native';

function SettingsButton() {
  return (
    <Pressable testID="open-settings" accessibilityLabel="Settings" onPress={() => router.push('/settings')}>
      <Text>Settings</Text>
    </Pressable>
  );
}

export default function RootLayout() {
  return (
    <Stack screenOptions={{ headerStyle: { backgroundColor: '#000000' }, headerTintColor: '#B0B0B0' }}>
      <Stack.Screen name="index" options={{ title: 'Halo', headerRight: () => <SettingsButton /> }} />
    </Stack>
  );
}
