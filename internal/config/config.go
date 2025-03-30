package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Tracker struct {
		APIIssuesURL        string `mapstructure:"TRACKER_API_ISSUES_URL"`
		OrgID               string `mapstructure:"TRACKER_ORG_ID"`
		OAuthToken          string `mapstructure:"TRACKER_OAUTH_TOKEN"`
		InitialHistoryDepth string `mapstructure:"TRACKER_INITIAL_HISTORY_DEPTH"`
		Filter              string `mapstructure:"TRACKER_FILTER"`
	} `mapstructure:",squash"`
	PostgreSQL struct {
		Host     string `mapstructure:"PG_HOST"`
		Port     int    `mapstructure:"PG_PORT"`
		Database string `mapstructure:"PG_DB"`
		User     string `mapstructure:"PG_USER"`
		Password string `mapstructure:"PG_PASSWORD"`
		SSLMode  string `mapstructure:"PG_SSLMODE"`
	} `mapstructure:",squash"`
	App struct {
		LogLevel string `mapstructure:"LOG_LEVEL"`
	} `mapstructure:",squash"`
}

var cfg Config

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("PG_PORT", 5432)
	viper.SetDefault("PG_SSLMODE", "disable")
	viper.SetDefault("LOG_LEVEL", "info")

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validateConfig() error {
	if cfg.Tracker.APIIssuesURL == "" {
		return fmt.Errorf("TRACKER_API_ISSUES_URL is required")
	}
	if cfg.Tracker.OrgID == "" {
		return fmt.Errorf("TRACKER_ORG_ID is required")
	}
	if cfg.Tracker.OAuthToken == "" {
		return fmt.Errorf("TRACKER_OAUTH_TOKEN is required")
	}
	if cfg.PostgreSQL.Host == "" {
		return fmt.Errorf("PG_HOST is required")
	}
	if cfg.PostgreSQL.Database == "" {
		return fmt.Errorf("PG_DB is required")
	}
	if cfg.PostgreSQL.User == "" {
		return fmt.Errorf("PG_USER is required")
	}
	if cfg.PostgreSQL.Password == "" {
		return fmt.Errorf("PG_PASSWORD is required")
	}
	return nil
}

// GetDSN returns the database connection string in the format required by pgx
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.PostgreSQL.Host,
		c.PostgreSQL.Port,
		c.PostgreSQL.Database,
		c.PostgreSQL.User,
		c.PostgreSQL.Password,
		c.PostgreSQL.SSLMode,
	)
}

// GetMigrateDSN returns the database connection string in the format required by migrate
func (c *Config) GetMigrateDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PostgreSQL.User,
		c.PostgreSQL.Password,
		c.PostgreSQL.Host,
		c.PostgreSQL.Port,
		c.PostgreSQL.Database,
		c.PostgreSQL.SSLMode,
	)
}
