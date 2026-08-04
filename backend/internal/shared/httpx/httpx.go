package httpx

import (
	"encoding/json"
	"net/http"
)

// Err is the standard error body returned by services.
type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON encodes v as JSON and sets the status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteErr writes an error body with the given status code.
func WriteErr(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, Err{Code: code, Message: msg})
}

// Decode reads and decodes a JSON request body into v.
func Decode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
