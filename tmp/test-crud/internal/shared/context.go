package shared

import "net/http"

type contextKey string

const (
	// UserIDKey is used to store the authenticated user's ID in request context.
	UserIDKey contextKey = "user_id"
	// TokenJTIKey is used to store the JWT ID in request context.
	TokenJTIKey contextKey = "token_jti"
	// RequestIDKey is used to store the request ID in request context.
	RequestIDKey contextKey = "request_id"
)

// UserID returns the authenticated user's ID from the request context.
func UserID(r *http.Request) (int64, bool) {
	id, ok := r.Context().Value(UserIDKey).(int64)
	return id, ok
}

// TokenJTI returns the JWT ID from the request context.
func TokenJTI(r *http.Request) (string, bool) {
	jti, ok := r.Context().Value(TokenJTIKey).(string)
	return jti, ok
}

// GetRequestID returns the request ID from the request context.
func GetRequestID(r *http.Request) string {
	id, _ := r.Context().Value(RequestIDKey).(string)
	return id
}
