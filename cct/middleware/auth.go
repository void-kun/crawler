package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"cct/pkg/logger"
	"cct/utils"
)

// UserIDKey is the context key for the user ID
type UserIDKey string

// UserIDContextKey is the key used to store the user ID in the request context
const UserIDContextKey UserIDKey = "user_id"

// GetUserID gets the user ID from the request context
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDContextKey).(int)
	return userID, ok
}

// publicPaths contains paths that don't require authentication
// Add more paths here if you want to make them accessible without authentication
var publicPaths = map[string]bool{
	"/api/auth/login":    true,
	"/api/auth/register": true,
	// Example: "/api/public/endpoint": true,
}

// AuthMiddleware is a middleware that checks for a valid API token
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for public paths
		if publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}
		// Get the Authorization header
		authHeader := r.Header.Get("Api-Key")
		if authHeader == "" {
			http.Error(w, "Api-Key header is required", http.StatusUnauthorized)
			logger.Debug().Str("path", r.URL.Path).Msg("Missing Api-Key header")
			return
		}

		// Check if the header starts with "Bearer "
		if authHeader == "" {
			http.Error(w, "Invalid Api-Key format. Expected 'TOKEN'", http.StatusUnauthorized)
			logger.Debug().Str("path", r.URL.Path).Msg("Invalid Api-Key format")
			return
		}

		// Extract the token
		token := authHeader
		if token == "" {
			http.Error(w, "Token is required", http.StatusUnauthorized)
			logger.Debug().Str("path", r.URL.Path).Msg("Empty token")
			return
		}

		// Validate the token
		var userID int
		var expiresAt sql.NullTime
		err := utils.DB.QueryRow(`
			SELECT user_id, expires_at
			FROM api_tokens
			WHERE token = $1
		`, token).Scan(&userID, &expiresAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				logger.Debug().Str("path", r.URL.Path).Msg("Invalid token")
				return
			}
			http.Error(w, "Failed to validate token: "+err.Error(), http.StatusInternalServerError)
			logger.Error().Err(err).Str("path", r.URL.Path).Msg("Failed to validate token")
			return
		}

		// Check if the token has expired
		if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
			http.Error(w, "Token has expired", http.StatusUnauthorized)
			logger.Debug().
				Str("path", r.URL.Path).
				Time("expires_at", expiresAt.Time).
				Msg("Token expired")
			return
		}

		// Update last_used_at
		_, err = utils.DB.Exec(`
			UPDATE api_tokens
			SET last_used_at = NOW()
			WHERE token = $1
		`, token)
		if err != nil {
			// Log the error but don't fail the request
			// This is not critical for the request to succeed
			logger.Warn().
				Err(err).
				Str("path", r.URL.Path).
				Int("user_id", userID).
				Msg("Failed to update token last_used_at")
		}

		// Set the user ID in the request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
