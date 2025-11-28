package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type JSONErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func WriteJSONError(w http.ResponseWriter, err error, clientMessage string, code int) {

	if err != nil {
		log.Printf("ERROR [%d]: %s | Internal: %v", code, clientMessage, err)
	} else {
		log.Printf("ERROR [%d]: %s", code, clientMessage)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := JSONErrorResponse{
		Status:  "error",
		Message: clientMessage,
	}

	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
