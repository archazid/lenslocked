package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"archazid.io/lenslocked/rand"
)

const (
	// Default time that a PasswordReset is valid for.
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID     int
	UserID int
	// Token is only set when a PasswordReset is being created.
	Token     string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB *sql.DB
	// Bytes to use when generating each password reset token.
	// If this value is not set or less than const MinBytesPerToken then it wil be ignored.
	BytesPerToken int
	// Valid duration time for a PasswordReset
	Duration time.Duration
}

func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {
	// Verify we have a valid email address for a user
	email = strings.ToLower(email)
	var userID int
	row := service.DB.QueryRow(`
		SELECT id
		FROM users
		WHERE email = $1;
	`, email)
	err := row.Scan(&userID)
	if err != nil {
		// TODO: Consider returning a specific error when the user does not exist.
		return nil, fmt.Errorf("create: %w", err)
	}

	// Create the token
	bytesPerToken := service.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	// Duration for valid token before expired
	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}

	// Build the PasswordReset
	pwReset := PasswordReset{
		UserID:    userID,
		Token:     token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = service.DB.QueryRow(`
		INSERT INTO
			password_resets (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3) ON
		CONFLICT (user_id) DO
		UPDATE
		SET
			token_hash = $2, expires_at = $3
		RETURNING id;
	`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &pwReset, nil
}

// Consume a token and return the user associated with it, or return error for any invalid token.
func (service *PasswordResetService) Consume(token string) (*User, error) {
	tokenHash := service.hash(token)
	var user User
	var pwReset PasswordReset
	row := service.DB.QueryRow(`
		SELECT
			pr.id,
			pr.expires_at,
			u.id,
			u.email,
			u.password_hash
		FROM
			password_resets pr
		JOIN users u ON
			u.id = pr.user_id
		WHERE
			pr.token_hash = $1;
	`, tokenHash)
	err := row.Scan(&pwReset.ID, &pwReset.ExpiresAt, &user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	if time.Now().After(pwReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", token)
	}
	err = service.delete(pwReset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	return &user, nil
}

func (service *PasswordResetService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (service *PasswordResetService) delete(id int) error {
	_, err := service.DB.Exec(`
		DELETE FROM password_resets
		WHERE id = $1;
	`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}
