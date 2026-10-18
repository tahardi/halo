package config_test

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/config"
)

func baseEnv() map[string]string {
	return map[string]string{
		"HALO_KEY":          "secret",
		"HALO_DB_PATH":      "/data/halo.db",
		"HALO_MEDIA_DIR":    "/data/media",
		"ANTHROPIC_API_KEY": "sk-test",
		"WHISPER_MODEL":     "/models/ggml-base.en.bin",
	}
}

func with(env map[string]string, key string, value string) map[string]string {
	out := maps.Clone(env)
	out[key] = value
	return out
}

func TestLoad(t *testing.T) {
	defaults := config.Config{
		ListenAddr:      "127.0.0.1:8082",
		Key:             "secret",
		DBPath:          "/data/halo.db",
		MediaDir:        "/data/media",
		LLMProvider:     "anthropic",
		AnthropicAPIKey: "sk-test",
		WhisperBin:      "whisper-cli",
		WhisperModel:    "/models/ggml-base.en.bin",
		FFmpegBin:       "ffmpeg",
	}
	withHome := defaults
	withHome.HomeLat = new(44.47)
	withHome.HomeLon = new(-73.21)
	withHomeBounds := defaults
	withHomeBounds.HomeLat = new(-90.0)
	withHomeBounds.HomeLon = new(180.0)
	overridden := defaults
	overridden.ListenAddr = "127.0.0.1:9000"
	overridden.WhisperBin = "/bin/whisper"
	overridden.FFmpegBin = "/bin/ffmpeg"

	tests := []struct {
		name    string
		env     map[string]string
		want    config.Config
		wantErr error
	}{
		{name: "happy path - defaults applied", env: baseEnv(), want: defaults},
		{
			name: "happy path - overrides",
			env: with(with(with(baseEnv(), "LISTEN_ADDR", "127.0.0.1:9000"),
				"WHISPER_BIN", "/bin/whisper"), "FFMPEG_BIN", "/bin/ffmpeg"),
			want: overridden,
		},
		{
			name: "happy path - home location",
			env:  with(with(baseEnv(), "HALO_HOME_LAT", "44.47"), "HALO_HOME_LON", "-73.21"),
			want: withHome,
		},
		{
			name: "happy path - home location on range bounds",
			env:  with(with(baseEnv(), "HALO_HOME_LAT", "-90"), "HALO_HOME_LON", "180"),
			want: withHomeBounds,
		},
		{
			name: "happy path - explicit anthropic provider",
			env:  with(baseEnv(), "HALO_LLM_PROVIDER", "anthropic"),
			want: defaults,
		},
		{name: "error - key empty", env: with(baseEnv(), "HALO_KEY", ""), wantErr: config.ErrMissingKey},
		{name: "error - key whitespace", env: with(baseEnv(), "HALO_KEY", "  "), wantErr: config.ErrMissingKey},
		{name: "error - db path empty", env: with(baseEnv(), "HALO_DB_PATH", ""), wantErr: config.ErrMissingDBPath},
		{
			name:    "error - media dir empty",
			env:     with(baseEnv(), "HALO_MEDIA_DIR", " "),
			wantErr: config.ErrMissingMediaDir,
		},
		{
			name:    "error - unknown provider",
			env:     with(baseEnv(), "HALO_LLM_PROVIDER", "ollama"),
			wantErr: config.ErrUnknownProvider,
		},
		{
			name:    "error - anthropic key empty",
			env:     with(baseEnv(), "ANTHROPIC_API_KEY", ""),
			wantErr: config.ErrMissingAnthropicKey,
		},
		{
			name:    "error - whisper model empty",
			env:     with(baseEnv(), "WHISPER_MODEL", ""),
			wantErr: config.ErrMissingWhisperModel,
		},
		{
			name:    "error - only home lat",
			env:     with(baseEnv(), "HALO_HOME_LAT", "44.47"),
			wantErr: config.ErrInvalidHome,
		},
		{
			name:    "error - only home lon",
			env:     with(baseEnv(), "HALO_HOME_LON", "-73.21"),
			wantErr: config.ErrInvalidHome,
		},
		{
			name:    "error - home lon not a number",
			env:     with(with(baseEnv(), "HALO_HOME_LAT", "44.47"), "HALO_HOME_LON", "west"),
			wantErr: config.ErrInvalidHome,
		},
		{
			name:    "error - home lat not a number",
			env:     with(with(baseEnv(), "HALO_HOME_LAT", "north"), "HALO_HOME_LON", "-73.21"),
			wantErr: config.ErrInvalidHome,
		},
		{
			name:    "error - home lat out of range",
			env:     with(with(baseEnv(), "HALO_HOME_LAT", "91"), "HALO_HOME_LON", "-73.21"),
			wantErr: config.ErrInvalidHome,
		},
		{
			name:    "error - home lon out of range",
			env:     with(with(baseEnv(), "HALO_HOME_LAT", "44.47"), "HALO_HOME_LON", "-181"),
			wantErr: config.ErrInvalidHome,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			getenv := func(k string) string { return tc.env[k] }

			// when
			got, err := config.Load(getenv)

			// then
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
