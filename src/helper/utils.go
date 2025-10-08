package helper

import (
	"encoding/json"
	"net/http"
	"os"
)

// Response is an alias to APIResponse for backward compatibility
type Response = APIResponse

// RespondWithJSON sends a JSON response (wrapper for backward compatibility)
func RespondWithJSON(w http.ResponseWriter, statusCode int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
