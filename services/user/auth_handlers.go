package main

import (
	"encoding/json"
	"net/http"

	"github.com/clinicmanager/services/user/service"
	sharedErrors "github.com/clinicmanager/shared/errors"
)

func registerHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		payload, err := svc.Register(r.Context(), body.Email, body.Password)
		if err != nil {
			code := http.StatusInternalServerError
			if appErr, ok := err.(*sharedErrors.AppError); ok {
				switch appErr.Code {
				case "VALIDATION_ERROR":
					code = http.StatusBadRequest
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}
}

func loginHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		payload, err := svc.Login(r.Context(), body.Email, body.Password)
		if err != nil {
			code := http.StatusInternalServerError
			if appErr, ok := err.(*sharedErrors.AppError); ok {
				switch appErr.Code {
				case "UNAUTHORIZED":
					code = http.StatusUnauthorized
				case "FORBIDDEN":
					code = http.StatusForbidden
				case "VALIDATION_ERROR":
					code = http.StatusBadRequest
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}
}

func verifyHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Token is required"})
			return
		}

		err := svc.ValidateEmail(r.Context(), token)
		if err != nil {
			code := http.StatusInternalServerError
			if appErr, ok := err.(*sharedErrors.AppError); ok {
				switch appErr.Code {
				case "VALIDATION_ERROR":
					code = http.StatusBadRequest
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "email validated successfully"})
	}
}
