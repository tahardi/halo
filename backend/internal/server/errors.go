package server

import (
	"encoding/json"
	"net/http"

	"github.com/tahardi/halo/backend/internal/model"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "encoding response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, model.ErrorResponse{
		Error: model.ErrorBody{Code: code, Message: message},
	})
}
