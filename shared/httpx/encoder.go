package httpx

import (
	"encoding/json"
	"net/http"
)

// WriteJSON sets the content-type, status code, and serializes the value as JSON.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// OK writes a 200 OK JSON response.
func OK(w http.ResponseWriter, v any) {
	WriteJSON(w, http.StatusOK, v)
}

// Created writes a 201 Created JSON response.
func Created(w http.ResponseWriter, v any) {
	WriteJSON(w, http.StatusCreated, v)
}
