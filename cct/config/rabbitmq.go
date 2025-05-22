package config

import "time"

// RabbitMQConfig holds the configuration for RabbitMQ
type RabbitMQConfig struct {
	URL               string        `mapstructure:"url"`
	QueueName         string        `mapstructure:"queue_name"`
	ExchangeName      string        `mapstructure:"exchange_name"`
	ExchangeType      string        `mapstructure:"exchange_type"`
	RoutingKeys       []string      `mapstructure:"routing_keys"`
	PrefetchCount     int           `mapstructure:"prefetch_count"`
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
}
