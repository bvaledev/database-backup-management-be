package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if status != http.StatusNoContent && payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("erro ao codificar JSON: %v", err)
			http.Error(w, "erro interno ao gerar resposta", http.StatusInternalServerError)
		}
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func JSONError(w http.ResponseWriter, status int, msg string) {
	JSONResponse(w, status, ErrorResponse{Error: msg})
}
