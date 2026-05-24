package httpReps

import (
	"encoding/json"
	"net/http"
)

func ResponseWithError(w http.ResponseWriter, code int, message string) {
	ResponseWithsJSON(w, code, map[string]string{"error": message})
}

func ResponseWithsJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
