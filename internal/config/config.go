package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string
	LogLevel    string
	AutoMigrate bool
	Server      ServerConfig
	CLI         CLIConfig
	Auth        AuthConfig
	Install     InstallConfig
	Postgres    PostgresConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type CLIConfig struct {
	BaseURL string
	Timeout string
}

type AuthConfig struct {
	BootstrapToken         string
	BootstrapTenantID      string
	BootstrapPrincipalName string
}

type InstallConfig struct {
	BaseURL string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.SetEnvPrefix("LLM_WIKI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	setDefaults(v)

	cfg := Config{
		Environment: v.GetString("environment"),
		LogLevel:    v.GetString("log_level"),
		AutoMigrate: v.GetBool("auto_migrate"),
		Server: ServerConfig{
			Host: v.GetString("server.host"),
			Port: v.GetInt("server.port"),
		},
		CLI: CLIConfig{
			BaseURL: v.GetString("cli.base_url"),
			Timeout: v.GetString("cli.timeout"),
		},
		Auth: AuthConfig{
			BootstrapToken:         v.GetString("auth.bootstrap_token"),
			BootstrapTenantID:      v.GetString("auth.bootstrap_ns"),
			BootstrapPrincipalName: v.GetString("auth.bootstrap_principal_name"),
		},
		Install: InstallConfig{
			BaseURL: v.GetString("install.base_url"),
		},
		Postgres: PostgresConfig{
			Host:     v.GetString("postgres.host"),
			Port:     v.GetInt("postgres.port"),
			User:     v.GetString("postgres.user"),
			Password: v.GetString("postgres.password"),
			Database: v.GetString("postgres.database"),
			SSLMode:  v.GetString("postgres.sslmode"),
		},
	}

	if cfg.Server.Port <= 0 {
		return Config{}, fmt.Errorf("server.port must be positive")
	}
	if strings.TrimSpace(cfg.CLI.BaseURL) == "" {
		return Config{}, fmt.Errorf("cli.base_url is required")
	}
	if strings.TrimSpace(cfg.Install.BaseURL) == "" {
		cfg.Install.BaseURL = cfg.CLI.BaseURL
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("environment", "development")
	v.SetDefault("log_level", "info")
	v.SetDefault("auto_migrate", true)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8234)
	v.SetDefault("cli.base_url", "http://127.0.0.1:8234")
	v.SetDefault("cli.timeout", "10s")
	v.SetDefault("auth.bootstrap_token", "dev-bootstrap-token")
	v.SetDefault("auth.bootstrap_ns", "default")
	v.SetDefault("auth.bootstrap_principal_name", "bootstrap-admin")
	v.SetDefault("install.base_url", "")

	v.SetDefault("postgres.host", "127.0.0.1")
	v.SetDefault("postgres.port", 15432)
	v.SetDefault("postgres.user", "llmwiki")
	v.SetDefault("postgres.password", "llmwiki")
	v.SetDefault("postgres.database", "llmwiki")
	v.SetDefault("postgres.sslmode", "disable")
}
