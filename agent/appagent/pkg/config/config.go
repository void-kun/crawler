package config

import (
	"time"

	"github.com/spf13/viper"
)

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

// ControlAPIConfig holds the configuration for the control API
type ControlAPIConfig struct {
	BaseURL                string        `mapstructure:"base_url"`
	Timeout                time.Duration `mapstructure:"timeout"`
	APIKey                 string        `mapstructure:"api_key"`
	AgentName              string        `mapstructure:"agent_name"`
	IPAddress              string        `mapstructure:"ip_address"`
	AgentHeartbeatInterval time.Duration `mapstructure:"agent_heartbeat_interval"`
	ReportResults          bool          `mapstructure:"report_results"`
	ResultsEndpoint        string        `mapstructure:"results_endpoint"`
}

// LoggerConfig holds the configuration for the logger
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
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

	// RabbitMQ settings
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`

	// Control API settings
	ControlAPI ControlAPIConfig `mapstructure:"control_api"`

	// Logger settings
	Logger LoggerConfig `mapstructure:"logger"`
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
