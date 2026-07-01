// Package response provides consistent JSON HTTP response helpers for
// generated handlers. Copy this package into your project — it does not
// import drp at runtime.
package response

import (
	"encoding/json"
	"net/http"
)

// envelope wraps a payload in a {"data": ...} envelope for successful responses.
type envelope struct {
	Data any `json:"data"`
}

// errorEnvelope wraps an error message for error responses.
type errorEnvelope struct {
	Error string `json:"error"`
}

// JSON writes a JSON response with the given HTTP status code.
// The payload is wrapped in {"data": payload}.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: payload})
}

// Error writes a {"error": message} JSON response with the given status code.
func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorEnvelope{Error: message})
}

// NoContent writes a 204 No Content response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// NotFound writes a standard 404 JSON response.
func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "not found")
}

// BadRequest writes a standard 400 JSON response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// UnprocessableEntity writes a standard 422 JSON response.
func UnprocessableEntity(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnprocessableEntity, message)
}

// InternalServerError writes a standard 500 JSON response.
func InternalServerError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}
