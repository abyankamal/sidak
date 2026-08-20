package handler

import (
	"encoding/json"
	"net/http"

	"github.com/abyankamal/sidak/backend/internal/domain"
)

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, statusCode int, errorMsg string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := domain.ErrorResponse{
		Error: errorMsg,
	}
	if len(details) > 0 {
		resp.Details = details
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func Message(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, domain.StandardMessageResponse{Message: message})
}
