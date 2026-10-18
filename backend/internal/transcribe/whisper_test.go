package transcribe_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/transcribe"
)

const (
	fakeFFmpeg       = "#!/bin/sh\nfor a; do last=$a; done; cp \"$3\" \"$last\"\n"
	fakeFFmpegFail   = "#!/bin/sh\necho 'bad input' >&2\nexit 1\n"
	fakeWhisper      = "#!/bin/sh\nprintf '  Hello from the trail.  \\n'\necho 'log line' >&2\n"
	fakeWhisperCat   = "#!/bin/sh\n[ \"$1\" = -m ] && [ -f \"$2\" ] && [ \"$3\" = -f ] && cat \"$4\"\n"
	fakeWhisperLines = "#!/bin/sh\nprintf ' I woke up at three.\\n\\n  I threw up until eleven. \\n'\n"
	fakeWhisperEmpty = "#!/bin/sh\nexit 0\n"
	fakeWhisperFail  = "#!/bin/sh\necho 'bad model' >&2\nexit 1\n"
)

func TestWhisper_Transcribe(t *testing.T) {
	tests := []struct {
		name           string
		ffmpeg         string
		whisper        string
		canceled       bool
		missingTempDir bool
		want           string
		wantErr        error
		wantErrMsg     string
	}{
		{
			name:    "happy path - trims output and ignores stderr",
			ffmpeg:  fakeFFmpeg,
			whisper: fakeWhisper,
			want:    "Hello from the trail.",
		},
		{
			name:    "happy path - passes model and converted wav to whisper",
			ffmpeg:  fakeFFmpeg,
			whisper: fakeWhisperCat,
			want:    "audio",
		},
		{
			name:    "happy path - joins multi-line output",
			ffmpeg:  fakeFFmpeg,
			whisper: fakeWhisperLines,
			want:    "I woke up at three. I threw up until eleven.",
		},
		{
			name:    "happy path - silent clip returns empty transcript",
			ffmpeg:  fakeFFmpeg,
			whisper: fakeWhisperEmpty,
			want:    "",
		},
		{
			name:       "error - whisper fails",
			ffmpeg:     fakeFFmpeg,
			whisper:    fakeWhisperFail,
			wantErr:    transcribe.ErrRunning,
			wantErrMsg: "whisper: exit status 1: bad model",
		},
		{
			name:       "error - ffmpeg fails",
			ffmpeg:     fakeFFmpegFail,
			whisper:    fakeWhisper,
			wantErr:    transcribe.ErrRunning,
			wantErrMsg: "ffmpeg: exit status 1: bad input",
		},
		{
			name:     "error - context canceled",
			ffmpeg:   fakeFFmpeg,
			whisper:  fakeWhisper,
			canceled: true,
			wantErr:  context.Canceled,
		},
		{
			name:           "error - temp dir missing",
			ffmpeg:         fakeFFmpeg,
			whisper:        fakeWhisper,
			missingTempDir: true,
			wantErr:        fs.ErrNotExist,
			wantErrMsg:     "creating temp wav",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			binDir := t.TempDir()
			audioPath := filepath.Join(binDir, "clip.m4a")
			require.NoError(t, os.WriteFile(audioPath, []byte("audio"), 0o600))
			modelPath := filepath.Join(binDir, "model.bin")
			require.NoError(t, os.WriteFile(modelPath, []byte("model"), 0o600))
			tempDir := t.TempDir()
			if tt.missingTempDir {
				tempDir = filepath.Join(tempDir, "missing")
			}
			w := transcribe.NewWhisper(
				writeScript(t, binDir, "ffmpeg", tt.ffmpeg),
				writeScript(t, binDir, "whisper-cli", tt.whisper),
				modelPath,
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
				require.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
			if tt.missingTempDir {
				assert.NoDirExists(t, tempDir)
				return
			}
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
