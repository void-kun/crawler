package models

import (
	"database/sql"
	"fmt"

	"cct/utils"

	"golang.org/x/crypto/bcrypt"
)

// GetUsers retrieves all users from the database
func GetUsers() ([]User, error) {
	rows, err := utils.DB.Query(`
		SELECT id, email, password_hash, is_active, created_at
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// GetUser retrieves a user by ID
func GetUser(id int) (User, error) {
	var u User
	err := utils.DB.QueryRow(`
		SELECT id, email, password_hash, is_active, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user with ID %d not found", id)
		}
		return User{}, fmt.Errorf("failed to query user: %w", err)
	}

	return u, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (User, error) {
	var u User
	err := utils.DB.QueryRow(`
		SELECT id, email, password_hash, is_active, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.IsActive, &u.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user with email %s not found", email)
		}
		return User{}, fmt.Errorf("failed to query user by email: %w", err)
	}

	return u, nil
}

// CreateUser creates a new user in the database
func CreateUser(u *User, password string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.PasswordHash = string(hashedPassword)

	err = utils.DB.QueryRow(`
		INSERT INTO users (email, password_hash, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, u.Email, u.PasswordHash, u.IsActive).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser updates an existing user
func UpdateUser(u *User) error {
	_, err := utils.DB.Exec(`
		UPDATE users
		SET email = $1, is_active = $2
		WHERE id = $3
	`, u.Email, u.IsActive, u.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(id int, password string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = utils.DB.Exec(`
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
	`, string(hashedPassword), id)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// VerifyUserPassword checks if the provided password matches the stored hash
func VerifyUserPassword(id int, password string) (bool, error) {
	var passwordHash string
	err := utils.DB.QueryRow(`
		SELECT password_hash
		FROM users
		WHERE id = $1
	`, id).Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("user with ID %d not found", id)
		}
		return false, fmt.Errorf("failed to query user password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("failed to compare password hash: %w", err)
	}

	return true, nil
}

// DeleteUser deletes a user by ID
func DeleteUser(id int) error {
	_, err := utils.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
