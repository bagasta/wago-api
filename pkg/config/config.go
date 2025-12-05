package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
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
	// Load variables from .env if it exists so local overrides work out of the box.
	_ = gotenv.Load()

	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	bindEnvs()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func bindEnvs() {
	keys := []string{
		"server.port",
		"server.name",
		"server.env",
		"database.url",
		"database.host",
		"database.port",
		"database.user",
		"database.password",
		"database.name",
		"database.ssl_mode",
		"database.max_connections",
		"database.max_idle_connections",
		"whatsapp.auto_reconnect",
		"whatsapp.qr_timeout",
		"whatsapp.log_level",
		"langchain.default_timeout",
		"langchain.max_retries",
		"langchain.base_url",
		"security.api_key_header",
		"security.rate_limit_requests",
		"security.rate_limit_window",
		"logging.level",
		"logging.format",
	}

	for _, key := range keys {
		_ = viper.BindEnv(key)
	}
}
