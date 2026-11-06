package main

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("LISTEN_ADDR", "127.0.0.1:0")
	t.Setenv("HALO_KEY", "secret")
	t.Setenv("HALO_DB_PATH", filepath.Join(root, "halo.db"))
	t.Setenv("HALO_MEDIA_DIR", filepath.Join(root, "media"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	t.Setenv("WHISPER_MODEL", "/models/ggml-base.en.bin")
}

func TestRun(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		env       map[string]string
		cancelled bool
		wantErr   error
		wantMsg   string
	}{
		{
			name:      "happy path - serve stops when context is cancelled",
			args:      []string{"serve"},
			cancelled: true,
		},
		{
			name:      "happy path - no arguments runs serve",
			args:      nil,
			cancelled: true,
		},
		{
			name:    "error - unknown command",
			args:    []string{"frobnicate"},
			wantErr: errUnknownCommand,
			wantMsg: "unknown command: frobnicate",
		},
		{
			name:    "error - no arguments with no key",
			args:    nil,
			env:     map[string]string{"HALO_KEY": ""},
			wantMsg: "HALO_KEY is required",
		},
		{
			name:    "error - serve with no key",
			args:    []string{"serve"},
			env:     map[string]string{"HALO_KEY": ""},
			wantMsg: "HALO_KEY is required",
		},
		{
			name:    "error - serve with media folder that cannot be created",
			args:    []string{"serve"},
			env:     map[string]string{"HALO_MEDIA_DIR": "/dev/null/media"},
			wantMsg: "creating media folder",
		},
		{
			name:    "error - serve with bad listen address",
			args:    []string{"serve"},
			env:     map[string]string{"LISTEN_ADDR": "nope"},
			wantMsg: "missing port in address",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			setValidEnv(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tc.cancelled {
				cancel()
			}
			logger := slog.New(slog.DiscardHandler)

			// when
			err := run(ctx, tc.args, logger)

			// then
			if tc.wantMsg == "" {
				require.NoError(t, err)
				return
			}
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			}
			require.ErrorContains(t, err, tc.wantMsg)
		})
	}
}
