package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantOK     bool
		wantStatus int
		wantName   string
	}{
		{name: "happy path - valid body", body: `{"name":"rizatriptan"}`, wantOK: true, wantName: "rizatriptan"},
		{name: "error - not json", body: `nope`, wantStatus: http.StatusBadRequest},
		{name: "error - unknown field", body: `{"name":"x","extra":1}`, wantStatus: http.StatusBadRequest},
		{name: "error - empty body", body: ``, wantStatus: http.StatusBadRequest},
		{name: "error - wrong field type", body: `{"name":1}`, wantStatus: http.StatusBadRequest},
		{
			name:       "error - over 1 MB",
			body:       `{"name":"` + strings.Repeat("a", maxJSONBody) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/x", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			var out struct {
				Name string `json:"name"`
			}

			// when
			ok := decodeJSON(rec, req, &out)

			// then
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.wantName, out.Name)
				return
			}
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}
