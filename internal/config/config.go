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
	Install     InstallConfig
	Postgres    PostgresConfig
	Redis       RedisConfig
	MinIO       MinIOConfig
	OpenSearch  OpenSearchConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type CLIConfig struct {
	BaseURL string
	Timeout string
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

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type OpenSearchConfig struct {
	URL string
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
		Redis: RedisConfig{
			Addr:     v.GetString("redis.addr"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		MinIO: MinIOConfig{
			Endpoint:  v.GetString("minio.endpoint"),
			AccessKey: v.GetString("minio.access_key"),
			SecretKey: v.GetString("minio.secret_key"),
			Bucket:    v.GetString("minio.bucket"),
			UseSSL:    v.GetBool("minio.use_ssl"),
		},
		OpenSearch: OpenSearchConfig{
			URL: v.GetString("opensearch.url"),
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
	v.SetDefault("install.base_url", "")

	v.SetDefault("postgres.host", "127.0.0.1")
	v.SetDefault("postgres.port", 15432)
	v.SetDefault("postgres.user", "llmwiki")
	v.SetDefault("postgres.password", "llmwiki")
	v.SetDefault("postgres.database", "llmwiki")
	v.SetDefault("postgres.sslmode", "disable")

	v.SetDefault("redis.addr", "127.0.0.1:16379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("minio.endpoint", "127.0.0.1:19000")
	v.SetDefault("minio.access_key", "minioadmin")
	v.SetDefault("minio.secret_key", "minioadmin")
	v.SetDefault("minio.bucket", "llm-wiki")
	v.SetDefault("minio.use_ssl", false)

	v.SetDefault("opensearch.url", "http://127.0.0.1:19200")
}
