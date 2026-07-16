package httpx

import (
	"encoding/json"
	"net/http"
)

// Decode parses the JSON request body into a pointer to T.
func Decode[T any](r *http.Request) (*T, error) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
