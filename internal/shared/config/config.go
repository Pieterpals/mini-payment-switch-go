package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Telegram TelegramConfig `mapstructure:"telegram"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
	OTel struct {
		CollectorURL string `mapstructure:"collector_url"`
	} `mapstructure:"otel"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// ConnMaxLifetimeDuration parses the connection max lifetime string into a time.Duration.
func (d DatabaseConfig) ConnMaxLifetimeDuration() time.Duration {
	dur, err := time.ParseDuration(d.ConnMaxLifetime)
	if err != nil {
		return 5 * time.Minute // safe default
	}
	return dur
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// KafkaConfig holds Kafka broker and topic settings.
type KafkaConfig struct {
	Brokers       []string         `mapstructure:"brokers"`
	Topic         KafkaTopicConfig `mapstructure:"topic"`
	ConsumerGroup string           `mapstructure:"consumer_group"`
}

// KafkaTopicConfig holds Kafka topic names.
type KafkaTopicConfig struct {
	PaymentSuccess string `mapstructure:"payment_success"`
}

// TelegramConfig holds Telegram Bot API settings for notifications.
type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
}

// Load reads configuration from the given YAML file path and environment variables.
// Environment variables override YAML values using underscore-separated keys.
// Example: APP_PORT=9090 overrides app.port in YAML.
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
