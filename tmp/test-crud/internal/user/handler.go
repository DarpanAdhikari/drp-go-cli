package user

import (
	"net/http"

	"/tmp/test-crud/internal/shared"
)

// Handler handles user-related HTTP requests.
type Handler struct {
	svc *Service
}

// NewHandler constructs a new Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Me handles GET /me — returns the authenticated user's profile.
//
// @Summary Get current user profile
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} user.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	id, ok := shared.UserID(r)
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "authorization required")
		return
	}
	user, err := h.svc.FindByID(id)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	shared.WriteJSON(w, http.StatusOK, user.ToResponse())
}
