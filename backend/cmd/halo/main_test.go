package main

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDispatch(t *testing.T) {
	t.Run("error - unknown command", func(t *testing.T) {
		// given
		logger := slog.New(slog.DiscardHandler)

		// when
		err := dispatch(t.Context(), []string{"frobnicate"}, logger)

		// then
		require.ErrorIs(t, err, errUnknownCommand)
		require.ErrorContains(t, err, "frobnicate")
	})

	t.Run("error - serve with no config", func(t *testing.T) {
		// given
		t.Setenv("HALO_KEY", "")
		logger := slog.New(slog.DiscardHandler)

		// when
		err := dispatch(t.Context(), nil, logger)

		// then
		require.ErrorContains(t, err, "HALO_KEY is required")
	})
}
