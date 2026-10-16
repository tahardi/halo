package server

import (
	"errors"
	"net/http"
)

const (
	maxUpload   = 50 << 20
	maxJSONBody = 1 << 20
	maxMemory   = 32 << 20
)

func parseMultipart(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "upload is larger than 50 MB")
			return false
		}
		writeError(w, http.StatusBadRequest, "bad_request", "request is not a valid multipart form")
		return false
	}
	return true
}
