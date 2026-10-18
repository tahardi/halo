package server

import (
	"net/http"

	"github.com/tahardi/halo/backend/internal/model"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, model.HealthResponse{Status: "ok"})
}
