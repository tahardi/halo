package server

import (
	"encoding/json"
	"errors"
	"net/http"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "request body is larger than 1 MB")
			return false
		}
		writeError(w, http.StatusBadRequest, "bad_request", "request body is not valid JSON")
		return false
	}
	return true
}
