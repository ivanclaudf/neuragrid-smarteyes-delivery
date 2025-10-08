package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// RegisterHealthRoute registers a health check endpoint
func RegisterHealthRoute(r *mux.Router) {
	r.HandleFunc("/api/v1/health", func(w http.ResponseWriter, req *http.Request) {
		version := os.Getenv("VERSION")
		if version == "" {
			version = "unknown"
		}
		resp := HealthResponse{
			Status:  "ok",
			Version: version,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}
