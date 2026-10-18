package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func multipartBody(t *testing.T, audio []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("id", "rec-1"))
	part, err := writer.CreateFormFile("audio", "clip.m4a")
	require.NoError(t, err)
	_, err = part.Write(audio)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return &body, writer.FormDataContentType()
}

func TestParseMultipart(t *testing.T) {
	tests := []struct {
		name        string
		audio       []byte
		contentType string
		wantOK      bool
		wantStatus  int
	}{
		{name: "happy path - small upload", audio: []byte("audio"), wantOK: true},
		{name: "error - over 50 MB", audio: make([]byte, 51<<20), wantStatus: http.StatusRequestEntityTooLarge},
		{
			name:        "error - not multipart",
			audio:       []byte("audio"),
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "error - wrong multipart boundary",
			audio:       []byte("audio"),
			contentType: "multipart/form-data; boundary=wrong",
			wantStatus:  http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			body, contentType := multipartBody(t, tc.audio)
			if tc.contentType != "" {
				contentType = tc.contentType
			}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/x", body)
			req.Header.Set("Content-Type", contentType)
			rec := httptest.NewRecorder()

			// when
			ok := parseMultipart(rec, req)

			// then
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, "rec-1", req.FormValue("id"))
				return
			}
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}
