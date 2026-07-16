package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"/tmp/test-crud/internal/config"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims holds the JWT payload (minimal — no PII).
type Claims struct {
	UserID    int64  `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair contains both access and refresh tokens.
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	TokenType             string    `json:"token_type"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// GenerateTokenPair creates both access and refresh tokens for the given user.
func GenerateTokenPair(userID int64, cfg *config.Config) (*TokenPair, *Claims, *Claims, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(cfg.AccessTokenExpiryMinutes) * time.Minute)
	refreshExp := now.AddDate(0, 0, cfg.RefreshTokenExpiryDays)

	accessToken, accessClaims, err := generateToken(userID, cfg, TokenTypeAccess, accessExp, now)
	if err != nil {
		return nil, nil, nil, err
	}
	refreshToken, refreshClaims, err := generateToken(userID, cfg, TokenTypeRefresh, refreshExp, now)
	if err != nil {
		return nil, nil, nil, err
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
	}, accessClaims, refreshClaims, nil
}

func generateToken(userID int64, cfg *config.Config, tokenType string, exp, now time.Time) (string, *Claims, error) {
	jti, err := newJTI()
	if err != nil {
		return "", nil, err
	}

	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", nil, err
	}
	return signed, claims, nil
}

// ParseToken verifies and returns the claims from a JWT string.
func ParseToken(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// HashToken returns a SHA-256 hex digest of the token string.
func HashToken(tokenStr string) string {
	sum := sha256.Sum256([]byte(tokenStr))
	return hex.EncodeToString(sum[:])
}

func newJTI() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
