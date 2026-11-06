package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/transcribe"
)

func TestWhisper_Transcribe(t *testing.T) {
	whisperBin := os.Getenv("WHISPER_BIN")
	modelPath := os.Getenv("WHISPER_MODEL")
	if whisperBin == "" || modelPath == "" {
		if os.Getenv("CI") == "" {
			t.Skip("set WHISPER_BIN and WHISPER_MODEL (make whisper-build whisper-model)")
		}
		t.Fatal("WHISPER_BIN and WHISPER_MODEL must be set in CI")
	}

	t.Run("happy path - transcribes spoken clip", func(t *testing.T) {
		// given
		w := transcribe.NewWhisper("ffmpeg", whisperBin, modelPath, t.TempDir())

		// when
		got, err := w.Transcribe(t.Context(), "../testdata/migraine-summary.m4a")

		// then
		require.NoError(t, err)
		got = strings.ToLower(got)
		assert.Contains(t, got, "migraine")
		assert.Contains(t, got, "threw up")
	})

	t.Run("error - corrupt audio", func(t *testing.T) {
		// given
		audioPath := filepath.Join(t.TempDir(), "corrupt.m4a")
		require.NoError(t, os.WriteFile(audioPath, []byte("not audio"), 0o600))
		tempDir := t.TempDir()
		w := transcribe.NewWhisper("ffmpeg", whisperBin, modelPath, tempDir)

		// when
		_, err := w.Transcribe(t.Context(), audioPath)

		// then
		require.ErrorIs(t, err, transcribe.ErrRunning)
		require.ErrorContains(t, err, "ffmpeg: ")
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}
