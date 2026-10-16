package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireKey(t *testing.T) {
	unauthorized, err := os.ReadFile("../../../api/testdata/error-unauthorized.json")
	require.NoError(t, err)

	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{name: "error - no header", header: "", wantStatus: http.StatusUnauthorized},
		{name: "error - wrong key", header: "Bearer wrong", wantStatus: http.StatusUnauthorized},
		{name: "error - no scheme", header: "secret", wantStatus: http.StatusUnauthorized},
		{name: "error - key prefix only", header: "Bearer secre", wantStatus: http.StatusUnauthorized},
		{name: "happy path - bearer", header: "Bearer secret", wantStatus: http.StatusOK},
		{name: "happy path - lowercase scheme", header: "bearer secret", wantStatus: http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := requireKey("secret", inner)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			// when
			handler.ServeHTTP(rec, req)

			// then
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus == http.StatusUnauthorized {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
				assert.JSONEq(t, string(unauthorized), rec.Body.String())
			}
		})
	}
}
