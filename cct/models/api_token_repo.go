package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"cct/utils"
)

// GetAPITokens retrieves all API tokens from the database
func GetAPITokens() ([]APIToken, error) {
	rows, err := utils.DB.Query(`
		SELECT id, user_id, token, description, expires_at, created_at, last_used_at
		FROM api_tokens
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query API tokens: %w", err)
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Token, &t.Description, &t.ExpiresAt, &t.CreatedAt, &t.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan API token row: %w", err)
		}
		tokens = append(tokens, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API token rows: %w", err)
	}

	return tokens, nil
}

// GetAPIToken retrieves an API token by ID
func GetAPIToken(id int) (APIToken, error) {
	var t APIToken
	err := utils.DB.QueryRow(`
		SELECT id, user_id, token, description, expires_at, created_at, last_used_at
		FROM api_tokens
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.UserID, &t.Token, &t.Description, &t.ExpiresAt, &t.CreatedAt, &t.LastUsedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return APIToken{}, fmt.Errorf("API token with ID %d not found", id)
		}
		return APIToken{}, fmt.Errorf("failed to query API token: %w", err)
	}

	return t, nil
}

// GetAPITokensByUser retrieves all API tokens for a specific user
func GetAPITokensByUser(userID int) ([]APIToken, error) {
	rows, err := utils.DB.Query(`
		SELECT id, user_id, token, description, expires_at, created_at, last_used_at
		FROM api_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API tokens by user: %w", err)
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Token, &t.Description, &t.ExpiresAt, &t.CreatedAt, &t.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan API token row: %w", err)
		}
		tokens = append(tokens, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API token rows: %w", err)
	}

	return tokens, nil
}

// GetAPITokenByToken retrieves an API token by its token string
func GetAPITokenByToken(token string) (APIToken, error) {
	var t APIToken
	err := utils.DB.QueryRow(`
		SELECT id, user_id, token, description, expires_at, created_at, last_used_at
		FROM api_tokens
		WHERE token = $1
	`, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.Description, &t.ExpiresAt, &t.CreatedAt, &t.LastUsedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return APIToken{}, fmt.Errorf("API token not found")
		}
		return APIToken{}, fmt.Errorf("failed to query API token by token: %w", err)
	}

	return t, nil
}

// GenerateToken generates a random token string
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateAPIToken creates a new API token in the database
func CreateAPIToken(t *APIToken) error {
	// Generate a random token if not provided
	var err error
	if t.Token == "" {
		t.Token, err = GenerateToken()
		if err != nil {
			return err
		}
	}

	err = utils.DB.QueryRow(`
		INSERT INTO api_tokens (user_id, token, description, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, t.UserID, t.Token, t.Description, t.ExpiresAt).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create API token: %w", err)
	}

	return nil
}

// UpdateAPIToken updates an existing API token
func UpdateAPIToken(t *APIToken) error {
	_, err := utils.DB.Exec(`
		UPDATE api_tokens
		SET description = $1, expires_at = $2
		WHERE id = $3
	`, t.Description, t.ExpiresAt, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update API token: %w", err)
	}

	return nil
}

// UpdateAPITokenLastUsed updates the last_used_at timestamp for an API token
func UpdateAPITokenLastUsed(id int) error {
	_, err := utils.DB.Exec(`
		UPDATE api_tokens
		SET last_used_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to update API token last used: %w", err)
	}

	return nil
}

// DeleteAPIToken deletes an API token by ID
func DeleteAPIToken(id int) error {
	_, err := utils.DB.Exec("DELETE FROM api_tokens WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete API token: %w", err)
	}
	return nil
}

// DeleteExpiredAPITokens deletes all expired API tokens
func DeleteExpiredAPITokens() (int64, error) {
	result, err := utils.DB.Exec(`
		DELETE FROM api_tokens
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired API tokens: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}
