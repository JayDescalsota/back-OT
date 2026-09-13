package main

import (
	"net/http"

	"github.com/clinicmanager/services/user/service"
	"github.com/clinicmanager/shared/context"
	"github.com/clinicmanager/shared/httpx"
	"github.com/clinicmanager/shared/response"
)

type LogoutRequest struct {
	SessionID string `json:"sessionId"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func acceptInviteHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[AcceptInviteRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		res, err := svc.AcceptInvite(r.Context(), req.Token, req.Name, req.Password)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, res)
	}
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

		_, err = svc.Register(r.Context(), req.Email, req.Password)

		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.Created(w, response.Success("user registered successfully"))
	}
}

func loginHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[LoginRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		res, err := svc.Login(r.Context(), req.Email, req.Password, r.Header.Get("User-Agent"), r.RemoteAddr)
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

func logoutHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[LogoutRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		err = svc.Logout(r.Context(), req.SessionID)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, response.Success("logged out successfully"))
	}
}

func logoutAllHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := context.UserIDFromCtx(r.Context())
		if userID == "" {
			httpx.Error(w, response.Unauthorized("unauthorized"))
			return
		}

		err := svc.LogoutAll(r.Context(), userID)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, response.Success("logged out from all devices successfully"))
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func forgotPasswordHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[ForgotPasswordRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		err = svc.ForgotPassword(r.Context(), req.Email)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, response.Success("if the email exists, a reset link has been sent"))
	}
}

func resetPasswordHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[ResetPasswordRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		err = svc.ResetPassword(r.Context(), req.Token, req.NewPassword)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, response.Success("password reset successfully"))
	}
}

func refreshTokenHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.Decode[RefreshTokenRequest](r)
		if err != nil {
			httpx.Error(w, response.Validation("invalid request body"))
			return
		}

		res, err := svc.RefreshToken(r.Context(), req.RefreshToken)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OK(w, res)
	}
}
