package auth

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"/tmp/test-crud/internal/config"
	"/tmp/test-crud/internal/shared"
	"/tmp/test-crud/internal/user"
)

// Handler handles authentication-related HTTP requests.
type Handler struct {
	userService user.ServiceInterface
	cfg         *config.Config
	tokenStore  *TokenStore
	validate    *validator.Validate
}

// NewHandler constructs a new Handler.
func NewHandler(db *sql.DB, userService user.ServiceInterface, cfg *config.Config) *Handler {
	return &Handler{
		userService: userService,
		cfg:         cfg,
		tokenStore:  NewTokenStore(db),
		validate:    validator.New(),
	}
}

// Register handles POST /auth/register
//
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.RegisterRequest true "Registration payload"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterRequest
	if !shared.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	u, err := h.userService.Register(req)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	shared.WriteJSON(w, http.StatusCreated, map[string]any{"user": u.ToResponse()})
}

// Login handles POST /auth/login
//
// @Summary Authenticate a user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.LoginRequest true "Login payload"
// @Success 200 {object} auth.TokenPair
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if !shared.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	u, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		shared.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokens, accessClaims, refreshClaims, err := GenerateTokenPair(u.ID, h.cfg)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	// Upsert device session for refresh token (preserves trust level)
	if err := h.tokenStore.UpsertDeviceSession(u.ID, HashToken(tokens.RefreshToken), refreshClaims.ID, TokenTypeRefresh, ipAddress, userAgent, req.MacAddress, req.FCMToken, refreshClaims.ExpiresAt.Time); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	// Store access token
	if err := h.tokenStore.Store(accessClaims, HashToken(tokens.AccessToken), ipAddress, userAgent, req.MacAddress, req.FCMToken); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to store access token")
		return
	}

	shared.WriteJSON(w, http.StatusOK, tokens)
}

// Refresh handles POST /auth/refresh
//
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.RefreshRequest true "Refresh payload"
// @Success 200 {object} auth.TokenPair
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req user.RefreshRequest
	if !shared.DecodeJSON(w, r, &req) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	claims, err := ParseToken(req.RefreshToken, h.cfg)
	if err != nil || claims.TokenType != TokenTypeRefresh {
		shared.WriteError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	record, err := h.tokenStore.FindActiveRefreshByHash(HashToken(req.RefreshToken))
	if err != nil || record.JTI != claims.ID || record.UserID != claims.UserID {
		shared.WriteError(w, http.StatusUnauthorized, "refresh token revoked or expired")
		return
	}

	tokens, accessClaims, refreshClaims, err := GenerateTokenPair(claims.UserID, h.cfg)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	// Upsert device session for the new refresh token
	if err := h.tokenStore.UpsertDeviceSession(claims.UserID, HashToken(tokens.RefreshToken), refreshClaims.ID, TokenTypeRefresh, ipAddress, userAgent, req.MacAddress, req.FCMToken, refreshClaims.ExpiresAt.Time); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	// Store new access token
	if err := h.tokenStore.Store(accessClaims, HashToken(tokens.AccessToken), ipAddress, userAgent, req.MacAddress, req.FCMToken); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to store access token")
		return
	}

	// Revoke old refresh token
	if err := h.tokenStore.RotateRefresh(claims.ID, refreshClaims.ID); err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "failed to rotate refresh token")
		return
	}

	shared.WriteJSON(w, http.StatusOK, tokens)
}

// Logout handles POST /auth/logout
//
// @Summary Logout and revoke tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body user.LogoutRequest true "Logout payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Revoke access token by parsing it from the Authorization header
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		if claims, err := ParseToken(parts[1], h.cfg); err == nil {
			_ = h.tokenStore.RevokeJTI(claims.ID)
		}
	}
	// Also revoke the refresh token from request body
	var req user.LogoutRequest
	if shared.DecodeJSON(w, r, &req) && req.RefreshToken != "" {
		if claims, err := ParseToken(req.RefreshToken, h.cfg); err == nil && claims.TokenType == TokenTypeRefresh {
			_ = h.tokenStore.RevokeJTI(claims.ID)
		}
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "logout successful"})
}
