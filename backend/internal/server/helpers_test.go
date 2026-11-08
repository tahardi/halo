package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/model"
	"github.com/tahardi/halo/backend/internal/server"
)

const key = "secret"

func serve(deps server.Deps, req *http.Request) *httptest.ResponseRecorder {
	return serveWithLogger(deps, req, slog.New(slog.DiscardHandler))
}

func serveWithLogger(deps server.Deps, req *http.Request, logger *slog.Logger) *httptest.ResponseRecorder {
	handler := server.NewServer(key, deps, logger)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func request(t *testing.T, method string, path string, body []byte) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequestWithContext(t.Context(), method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+key)
	return req
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var resp model.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp.Error.Code
}
