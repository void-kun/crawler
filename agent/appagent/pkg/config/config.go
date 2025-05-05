package config

import (
	"time"

	"github.com/spf13/viper"
)

type Stv struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Origin   string `mapstructure:"origin"`
}

// RabbitMQConfig holds the configuration for RabbitMQ
type RabbitMQConfig struct {
	URL               string        `mapstructure:"url"`
	QueueName         string        `mapstructure:"queue_name"`
	ExchangeName      string        `mapstructure:"exchange_name"`
	ExchangeType      string        `mapstructure:"exchange_type"`
	RoutingKeys       []string      `mapstructure:"routing_keys"`
	PriorityTopic     string        `mapstructure:"priority_topic"`
	PrefetchCount     int           `mapstructure:"prefetch_count"`
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
}

// Config holds the configuration for the spider
type Config struct {
	// General settings
	Concurrency int           `mapstructure:"concurrency"`
	Delay       time.Duration `mapstructure:"delay"`
	UserAgent   []string      `mapstructure:"user_agent"`
	MaxDepth    int           `mapstructure:"max_depth"`

	// Headless browser settings
	BrowserPath    string        `mapstructure:"browser_path"`
	BrowserTimeout time.Duration `mapstructure:"browser_timeout"`
	ProxyURL       string        `mapstructure:"proxy_url"`

	// Storage settings
	OutputDir   string `mapstructure:"output_dir"`
	SessionFile string `mapstructure:"session_file"`

	// Source-specific settings
	Stv Stv `mapstructure:"stv"`

	// RabbitMQ settings
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// LoadConfigFromFile loads configuration from a specific file path
func LoadConfigFromFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
