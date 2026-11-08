import { Stack } from 'expo-router';
import { useEffect, useState } from 'react';
import { Button, ScrollView, StyleSheet, Text, TextInput } from 'react-native';

import { createClient, messageFor } from '@/api/client';
import { loadSettings, saveSettings } from '@/settings/store';

export default function SettingsScreen() {
  const [baseUrl, setBaseUrl] = useState('');
  const [key, setKey] = useState('');
  const [status, setStatus] = useState('');

  useEffect(() => {
    let cancelled = false;
    loadSettings()
      .then((s) => {
        if (!cancelled) {
          setBaseUrl(s.baseUrl);
          setKey(s.key);
        }
      })
      .catch((e: unknown) => setStatus(e instanceof Error ? e.message : String(e)));
    return () => {
      cancelled = true;
    };
  }, []);

  async function save() {
    try {
      await saveSettings({ baseUrl, key });
      setStatus('Saved');
    } catch (e) {
      setStatus(e instanceof Error ? e.message : String(e));
    }
  }

  async function testConnection() {
    try {
      await createClient({ baseUrl, key }).health();
      setStatus('Connected');
    } catch (e) {
      setStatus(messageFor(e));
    }
  }

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Stack.Screen options={{ title: 'Settings' }} />
      <Text>Backend URL</Text>
      <TextInput
        testID="settings-url"
        value={baseUrl}
        onChangeText={setBaseUrl}
        autoCapitalize="none"
        autoCorrect={false}
        keyboardType="url"
        style={styles.input}
      />
      <Text>Key</Text>
      <TextInput
        testID="settings-key"
        value={key}
        onChangeText={setKey}
        secureTextEntry
        autoCapitalize="none"
        autoCorrect={false}
        style={styles.input}
      />
      <Button testID="settings-save" title="Save" onPress={save} />
      <Button testID="settings-test" title="Test connection" onPress={testConnection} />
      <Text testID="settings-status">{status}</Text>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: 16, gap: 8 },
  input: { borderWidth: 1, borderColor: '#ccc', borderRadius: 8, padding: 8 },
});
