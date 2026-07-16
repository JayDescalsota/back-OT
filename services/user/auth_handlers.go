package main

import (
	"net/http"

	"github.com/clinicmanager/services/user/service"
	"github.com/clinicmanager/shared/httpx"
	"github.com/clinicmanager/shared/response"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func registerHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[RegisterRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		res, err := svc.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.Created(w, res)
	}
}

func loginHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[LoginRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		res, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, res)
	}
}

func verifyHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			httpx.Error(w, response.Validation("Token is required"))
			return
		}

		err := svc.ValidateEmail(r.Context(), token)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, response.Success("email validated successfully"))
	}
}

func changePasswordHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[ChangePasswordRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		err = svc.ChangePassword(r.Context(), req.Email, req.OldPassword, req.NewPassword)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, response.Success("password changed successfully"))
	}
}
