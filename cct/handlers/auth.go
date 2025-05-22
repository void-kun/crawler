package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cct/config"
	"cct/models"
	"cct/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response body for login
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at,omitempty"`
	UserID    int    `json:"user_id"`
}

// RegisterRequest represents the request body for registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents the response body for registration
type RegisterResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Login handles POST /auth/login
func Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		logger.Debug().Err(err).Msg("Invalid login request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		logger.Debug().
			Str("email", req.Email).
			Bool("has_password", req.Password != "").
			Msg("Missing required login fields")
		return
	}

	// Get user by email
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		// Don't reveal that the user doesn't exist
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		logger.Debug().
			Str("email", req.Email).
			Err(err).
			Msg("User not found")
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is inactive", http.StatusForbidden)
		logger.Debug().
			Str("email", req.Email).
			Int("user_id", user.ID).
			Msg("Inactive account login attempt")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		logger.Debug().
			Str("email", req.Email).
			Int("user_id", user.ID).
			Msg("Invalid password")
		return
	}

	// Generate API token
	token, err := models.GenerateToken()
	if err != nil {
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to generate token")
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load configuration: "+err.Error(), http.StatusInternalServerError)
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to load configuration")
		return
	}

	// Set token expiration based on configuration
	expiresAt := time.Now().Add(cfg.Auth.TokenExpiry)

	// Create API token in database
	apiToken := models.APIToken{
		UserID:      user.ID,
		Token:       token,
		Description: "Login token generated on " + time.Now().Format(time.RFC3339),
	}

	// Set expiration time
	apiToken.ExpiresAt.Time = expiresAt
	apiToken.ExpiresAt.Valid = true

	if err := models.CreateAPIToken(&apiToken); err != nil {
		http.Error(w, "Failed to create token: "+err.Error(), http.StatusInternalServerError)
		logger.Error().
			Err(err).
			Int("user_id", user.ID).
			Msg("Failed to create API token")
		return
	}

	// Return token to client
	resp := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		UserID:    user.ID,
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Time("expires_at", expiresAt).
		Msg("User logged in successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Register handles POST /auth/register
func Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		logger.Debug().Err(err).Msg("Invalid registration request body")
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		logger.Debug().
			Str("email", req.Email).
			Bool("has_password", req.Password != "").
			Msg("Missing required registration fields")
		return
	}

	// Check if email is already in use
	_, err := models.GetUserByEmail(req.Email)
	if err == nil {
		// Email already exists
		http.Error(w, "Email is already registered", http.StatusConflict)
		logger.Debug().
			Str("email", req.Email).
			Msg("Email already registered")
		return
	} else if !strings.Contains(err.Error(), "not found") {
		// Unexpected error
		http.Error(w, "Failed to check email: "+err.Error(), http.StatusInternalServerError)
		logger.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to check if email exists")
		return
	}

	// Create user
	user := models.User{
		Email:    req.Email,
		IsActive: true, // Users are active by default when self-registering
	}

	if err := models.CreateUser(&user, req.Password); err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		logger.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to create user")
		return
	}

	// Return user info
	resp := RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	logger.Info().
		Int("user_id", user.ID).
		Str("email", user.Email).
		Msg("User registered successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
