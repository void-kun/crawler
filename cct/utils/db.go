package utils

import (
	"database/sql"
	"fmt"

	"cct/config"
	"cct/pkg/logger"

	_ "github.com/lib/pq"
)

// DB is the database connection
var DB *sql.DB

// InitDB initializes the database connection
func InitDB(dbConfig *config.DatabaseConfig) error {
	// Create connection string
	connStr := dbConfig.GetDSN()

	// Open database connection
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().
		Str("host", dbConfig.Host).
		Int("port", dbConfig.Port).
		Str("database", dbConfig.Name).
		Msg("Database connection established")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		logger.Info().Msg("Database connection closed")
	}
}
