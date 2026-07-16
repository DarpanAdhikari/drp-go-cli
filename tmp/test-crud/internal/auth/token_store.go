package auth

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTokenRevoked = errors.New("token is revoked")

// TokenStore handles token persistence and validation.
type TokenStore struct {
	db *sql.DB
}

// TokenRecord represents a stored token record (without PII).
type TokenRecord struct {
	JTI       string
	UserID    int64
	TokenType string
	ExpiresAt time.Time
}

// NewTokenStore constructs a new TokenStore.
func NewTokenStore(db *sql.DB) *TokenStore {
	return &TokenStore{db: db}
}

// Store inserts a new token record.
func (s *TokenStore) Store(claims *Claims, tokenHash, ipAddress, userAgent, macAddress, fcmToken string) error {
	_, err := s.db.Exec(
		`INSERT INTO user_tokens (user_id, jti, token_hash, token_type, ip_address, user_agent, mac_address, fcm_token, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`,
		claims.UserID, claims.ID, tokenHash, claims.TokenType, ipAddress, userAgent, macAddress, fcmToken, claims.ExpiresAt.Time,
	)
	return err
}

// UpsertDeviceSession finds an existing active device session for the user
// (by mac_address, falling back to user_agent) and updates it in-place.
// If none is found, a new session is inserted.
// This preserves trust_level, authorized_at, and original created_at.
func (s *TokenStore) UpsertDeviceSession(userID int64, tokenHash, jti, tokenType, ipAddress, userAgent, macAddress, fcmToken string, expiresAt time.Time) error {
	var existingID int64

	// Prefer matching by mac_address
	if macAddress != "" {
		s.db.QueryRow(
			`SELECT id FROM user_tokens WHERE user_id = $1 AND mac_address = $2 AND revoked_at IS NULL AND token_type = $3 LIMIT 1`,
			userID, macAddress, tokenType,
		).Scan(&existingID)
	}

	// Fallback: match by user_agent (only if no mac_address was stored)
	if existingID == 0 {
		s.db.QueryRow(
			`SELECT id FROM user_tokens WHERE user_id = $1 AND user_agent = $2 AND mac_address IS NULL AND revoked_at IS NULL AND token_type = $3 LIMIT 1`,
			userID, userAgent, tokenType,
		).Scan(&existingID)
	}

	if existingID != 0 {
		// Update existing — preserve trust_level, authorized_at, created_at
		if fcmToken != "" {
			_, err := s.db.Exec(
				`UPDATE user_tokens SET jti = $1, token_hash = $2, ip_address = $3, user_agent = $4, mac_address = $5, fcm_token = $6, expires_at = $7, updated_at = NOW() WHERE id = $8`,
				jti, tokenHash, ipAddress, userAgent, macAddress, fcmToken, expiresAt, existingID,
			)
			return err
		}
		_, err := s.db.Exec(
			`UPDATE user_tokens SET jti = $1, token_hash = $2, ip_address = $3, user_agent = $4, mac_address = $5, expires_at = $6, updated_at = NOW() WHERE id = $7`,
			jti, tokenHash, ipAddress, userAgent, macAddress, expiresAt, existingID,
		)
		return err
	}

	// No existing session found — insert new
	_, err := s.db.Exec(
		`INSERT INTO user_tokens (user_id, jti, token_hash, token_type, ip_address, user_agent, mac_address, fcm_token, expires_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`,
		userID, jti, tokenHash, tokenType, ipAddress, userAgent, macAddress, fcmToken, expiresAt,
	)
	return err
}

// EnsureActive checks that a token is valid, not revoked, and not expired.
func (s *TokenStore) EnsureActive(jti, tokenType string) error {
	var revokedAt sql.NullTime
	var expiresAt time.Time
	err := s.db.QueryRow(
		`SELECT ut.revoked_at, ut.expires_at
		 FROM user_tokens ut
		 JOIN users u ON u.id = ut.user_id
		 WHERE ut.jti = $1 AND ut.token_type = $2 AND u.is_active = true
		 LIMIT 1`,
		jti, tokenType,
	).Scan(&revokedAt, &expiresAt)
	if err != nil {
		return err
	}
	if revokedAt.Valid || time.Now().After(expiresAt) {
		return ErrTokenRevoked
	}
	return nil
}

// FindActiveRefreshByHash retrieves a non-revoked refresh token record.
func (s *TokenStore) FindActiveRefreshByHash(tokenHash string) (*TokenRecord, error) {
	var rec TokenRecord
	var revokedAt sql.NullTime
	err := s.db.QueryRow(
		`SELECT ut.user_id, ut.jti, ut.token_type, ut.expires_at, ut.revoked_at
		 FROM user_tokens ut
		 JOIN users u ON u.id = ut.user_id
		 WHERE ut.token_hash = $1 AND ut.token_type = $2 AND u.is_active = true
		 LIMIT 1`,
		tokenHash, TokenTypeRefresh,
	).Scan(&rec.UserID, &rec.JTI, &rec.TokenType, &rec.ExpiresAt, &revokedAt)
	if err != nil {
		return nil, err
	}
	if revokedAt.Valid || time.Now().After(rec.ExpiresAt) {
		return nil, ErrTokenRevoked
	}
	return &rec, nil
}

// RevokeJTI marks a token as revoked.
func (s *TokenStore) RevokeJTI(jti string) error {
	_, err := s.db.Exec(`UPDATE user_tokens SET revoked_at = COALESCE(revoked_at, $1) WHERE jti = $2`, time.Now(), jti)
	return err
}

// RotateRefresh marks the old token as revoked and records the replacement.
func (s *TokenStore) RotateRefresh(oldJTI, newJTI string) error {
	_, err := s.db.Exec(
		`UPDATE user_tokens SET revoked_at = COALESCE(revoked_at, $1), replaced_by_jti = $2
		 WHERE jti = $3 AND token_type = $4`,
		time.Now(), newJTI, oldJTI, TokenTypeRefresh,
	)
	return err
}
