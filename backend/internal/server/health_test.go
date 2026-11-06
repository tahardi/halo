package server_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/server"
)

func TestServer_Health(t *testing.T) {
	want, err := os.ReadFile("../../../api/testdata/health-response.json")
	require.NoError(t, err)

	t.Run("happy path - ok with key", func(t *testing.T) {
		// given
		var logs bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logs, nil))
		req := request(t, http.MethodGet, "/health", nil)

		// when
		rec := serveWithLogger(server.Deps{}, req, logger)

		// then
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, string(want), rec.Body.String())
		assert.Contains(t, logs.String(), "method=GET")
		assert.Contains(t, logs.String(), "path=/health")
		assert.Contains(t, logs.String(), "status=200")
		assert.Contains(t, logs.String(), "duration=")
		assert.NotContains(t, logs.String(), key)
	})

	t.Run("error - no key", func(t *testing.T) {
		// given
		var logs bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logs, nil))
		req := request(t, http.MethodGet, "/health", nil)
		req.Header.Del("Authorization")

		// when
		rec := serveWithLogger(server.Deps{}, req, logger)

		// then
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Equal(t, "unauthorized", errorCode(t, rec))
		assert.Contains(t, logs.String(), "status=401")
	})

	t.Run("error - wrong method", func(t *testing.T) {
		// given
		req := request(t, http.MethodPost, "/health", nil)

		// when
		rec := serve(server.Deps{}, req)

		// then
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("error - unknown route", func(t *testing.T) {
		// given
		req := request(t, http.MethodGet, "/nope", nil)

		// when
		rec := serve(server.Deps{}, req)

		// then
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
