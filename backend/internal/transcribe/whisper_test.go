package transcribe_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/transcribe"
)

const (
	fakeFFmpeg      = "#!/bin/sh\nfor a; do last=$a; done; cp \"$3\" \"$last\"\n"
	fakeFFmpegFail  = "#!/bin/sh\necho 'bad input' >&2\nexit 1\n"
	fakeWhisper     = "#!/bin/sh\nprintf '  Hello from the trail.  \\n'\necho 'log line' >&2\n"
	fakeWhisperFail = "#!/bin/sh\necho 'bad model' >&2\nexit 1\n"
)

func TestWhisper_Transcribe(t *testing.T) {
	tests := []struct {
		name     string
		ffmpeg   string
		whisper  string
		canceled bool
		want     string
		wantErr  error
	}{
		{"happy path - trims output and ignores stderr", fakeFFmpeg, fakeWhisper, false, "Hello from the trail.", nil},
		{"error - whisper fails", fakeFFmpeg, fakeWhisperFail, false, "", transcribe.ErrRunning},
		{"error - ffmpeg fails", fakeFFmpegFail, fakeWhisper, false, "", transcribe.ErrRunning},
		{"error - context canceled", fakeFFmpeg, fakeWhisper, true, "", context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			binDir := t.TempDir()
			audioPath := filepath.Join(binDir, "clip.m4a")
			require.NoError(t, os.WriteFile(audioPath, []byte("audio"), 0o600))
			tempDir := t.TempDir()
			w := transcribe.NewWhisper(
				writeScript(t, binDir, "ffmpeg", tt.ffmpeg),
				writeScript(t, binDir, "whisper-cli", tt.whisper),
				filepath.Join(binDir, "model.bin"),
				tempDir,
			)
			ctx, cancel := context.WithCancel(t.Context())
			if tt.canceled {
				cancel()
			}
			defer cancel()

			// when
			got, err := w.Transcribe(ctx, audioPath)

			// then
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
			entries, err := os.ReadDir(tempDir)
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

func writeScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o700))
	return path
}
