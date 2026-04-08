package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ifuryst/llm-wiki/internal/config"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type PostgresInstance struct {
	Container *pgcontainer.PostgresContainer
	Config    config.PostgresConfig
}

func StartPostgres(tb testing.TB) *PostgresInstance {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := pgcontainer.Run(
		ctx,
		"postgres:17",
		pgcontainer.WithDatabase("llm-wiki"),
		pgcontainer.WithUsername("llm-wiki"),
		pgcontainer.WithPassword("llm-wiki"),
	)
	if err != nil {
		tb.Fatalf("start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		tb.Fatalf("postgres host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		tb.Fatalf("postgres port: %v", err)
	}

	tb.Cleanup(func() {
		_ = testcontainers.TerminateContainer(container)
	})

	return &PostgresInstance{
		Container: container,
		Config: config.PostgresConfig{
			Host:     host,
			Port:     port.Int(),
			User:     "llm-wiki",
			Password: "llm-wiki",
			Database: "llm-wiki",
			SSLMode:  "disable",
		},
	}
}

func MustBaseConfig(pg config.PostgresConfig) config.Config {
	return config.Config{
		Environment: "test",
		LogLevel:    "error",
		AutoMigrate: true,
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
		CLI: config.CLIConfig{
			BaseURL: "http://127.0.0.1:0",
			Timeout: "10s",
		},
		Postgres: pg,
	}
}

func UniqueTenant(prefix string, n int) string {
	return fmt.Sprintf("%s-%d", prefix, n)
}
