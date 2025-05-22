package config

import (
	"fmt"
	"log" // Standard log for initialization only
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

// AuthConfig holds all authentication-related configuration
type AuthConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	TokenExpiry time.Duration `mapstructure:"token_expiry"`
}

// LoggingConfig holds all logging-related configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Load loads the configuration from config.yml
func Load() (*Config, error) {
	// Set default configuration file
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using default values")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Allow environment variables to override config file
	viper.AutomaticEnv()

	// Parse the configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Convert durations from seconds to time.Duration
	config.Server.ReadTimeout = time.Duration(config.Server.ReadTimeout) * time.Second
	config.Server.WriteTimeout = time.Duration(config.Server.WriteTimeout) * time.Second
	config.Server.IdleTimeout = time.Duration(config.Server.IdleTimeout) * time.Second
	config.Auth.TokenExpiry = time.Duration(config.Auth.TokenExpiry) * time.Hour
	config.RabbitMQ.ReconnectInterval = time.Duration(config.RabbitMQ.ReconnectInterval) * time.Second

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 15)
	viper.SetDefault("server.idle_timeout", 60)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "crawler")
	viper.SetDefault("database.sslmode", "disable")

	// Auth defaults
	viper.SetDefault("auth.enabled", false)
	viper.SetDefault("auth.token_expiry", 720) // 30 days in hours

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.output", "console")
	viper.SetDefault("logging.file_path", "logs/app.log")
	viper.SetDefault("logging.max_size", 10)
	viper.SetDefault("logging.max_backups", 5)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.compress", true)

	// RabbitMQ defaults
	viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("rabbitmq.queue_name", "crawler_tasks")
	viper.SetDefault("rabbitmq.exchange_name", "crawler_exchange")
	viper.SetDefault("rabbitmq.exchange_type", "topic")
	viper.SetDefault("rabbitmq.routing_keys", []string{"crawl.#"})
	viper.SetDefault("rabbitmq.prefetch_count", 1)
	viper.SetDefault("rabbitmq.reconnect_interval", 5) // seconds
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}
