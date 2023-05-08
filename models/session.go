package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"

	"archazid.io/lenslocked/rand"
)

const (
	// The minimum number of bytes to be used for each session token.
	MinBytesPerToken = 32
)

type Session struct {
	ID     int
	UserID int
	// Token is only set when creating a new session. When looking up a session
	// this will be left empty, as we only store the hash of a session token
	// in our database and we cannot reverse it into a raw token
	Token     string
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
	// BytesPerToken is used to determine how many bytes to use when generating
	// each session token. If this value is not set or less than the
	// MinBytesPerToken const it will be ignored and use MinBytesPerToken instead
	BytesPerToken int
}

// Create a new session for the user provided. The session token
// will bew returned as the Token field on the Session type, but only the hashed
// session token is stored in the database.
func (ss *SessionService) Create(userID int) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	if bytesPerToken < MinBytesPerToken {
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	session := Session{
		UserID:    userID,
		Token:     token,
		TokenHash: ss.hash(token),
	}

	// Updating the database
	row := ss.DB.QueryRow(`
		INSERT INTO
			sessions (user_id, token_hash)
		VALUES ($1, $2) ON CONFLICT (user_id)
		DO
		UPDATE
		SET
			token_hash = $2
		RETURNING id;
	`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &session, nil
}

func (ss *SessionService) Delete(token string) error {
	tokenHash := ss.hash(token)

	_, err := ss.DB.Exec(`
		DELETE FROM sessions
		WHERE token_hash = $1;
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (ss *SessionService) User(token string) (*User, error) {
	// Hash the session token
	tokenHash := ss.hash(token)
	// Create user variable
	var user User

	// Query for the session with that hash
	row := ss.DB.QueryRow(`
		SELECT
			u.id,
			u.email,
			u.password_hash
		FROM
			sessions s
		JOIN users u ON
			u.id = s.user_id
		WHERE
			s.token_hash = $1;
	`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	// Return the user
	return &user, nil
}

func (ss *SessionService) hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	// base64 encode the data into a string
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
