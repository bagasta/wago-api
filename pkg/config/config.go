package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	WhatsApp  WhatsAppConfig  `mapstructure:"whatsapp"`
	Langchain LangchainConfig `mapstructure:"langchain"`
	Security  SecurityConfig  `mapstructure:"security"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	URL                string `mapstructure:"url"`
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	Name               string `mapstructure:"name"`
	SSLMode            string `mapstructure:"ssl_mode"`
	MaxConnections     int    `mapstructure:"max_connections"`
	MaxIdleConnections int    `mapstructure:"max_idle_connections"`
}

type WhatsAppConfig struct {
	AutoReconnect bool   `mapstructure:"auto_reconnect"`
	QRTimeout     int    `mapstructure:"qr_timeout"`
	LogLevel      string `mapstructure:"log_level"`
}

type LangchainConfig struct {
	DefaultTimeout string `mapstructure:"default_timeout"`
	MaxRetries     int    `mapstructure:"max_retries"`
	BaseURL        string `mapstructure:"base_url"`
}

type SecurityConfig struct {
	APIKeyHeader      string `mapstructure:"api_key_header"`
	RateLimitRequests int    `mapstructure:"rate_limit_requests"`
	RateLimitWindow   string `mapstructure:"rate_limit_window"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func LoadConfig() (*Config, error) {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
