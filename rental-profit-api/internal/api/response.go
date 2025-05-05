package api

import (
	"encoding/json"
	"log"
	"net/http"

	"rental-profit-api/internal/types"
)


func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// Log the marshalling error, which is an internal server issue
		log.Printf("Error marshalling JSON response: %v", err)
		// Send a generic internal server error response
		respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
    if err != nil {
        log.Printf("Error writing response: %v", err)
    }
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, types.ErrorResponse{Message: message})
}