package httpx

import (
	"context"
	"net/http"

	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/response"
)

// HTTPStatus maps custom response.Error codes to HTTP status codes.
func HTTPStatus(err error) int {
	if e, ok := err.(*response.Error); ok {
		switch e.Code {
		case "VALIDATION_ERROR":
			return http.StatusBadRequest
		case "UNAUTHORIZED":
			return http.StatusUnauthorized
		case "FORBIDDEN":
			return http.StatusForbidden
		case "NOT_FOUND":
			return http.StatusNotFound
		}
	}
	return http.StatusInternalServerError
}

// Error formats and writes an error response to the client.
func Error(w http.ResponseWriter, err error) {
	status := HTTPStatus(err)
	if status == http.StatusInternalServerError {
		logger.Error(context.Background(), "Internal Server Error", "error", err.Error())
	}
	WriteJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}
