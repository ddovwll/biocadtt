package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ddovwll/biocadtt/src/internal/infrastructure/data"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppEnv string `env:"APP_ENV" env-default:"local"`

	ReportsDir string `env:"REPORTS_DIR" env-default:"reports"`

	WatcherDir       string        `env:"WATCHER_DIR" env-default:"."`
	WatcherHeartbeat time.Duration `env:"WATCHER_HEARTBEAT" env-default:"3s"`

	HTTPAddr string `env:"HTTP_ADDR" env-default:":8080"`

	PostgresDSN               string        `env:"POSTGRES_DSN"`
	PostgresMaxConns          int32         `env:"POSTGRES_MAX_CONNS" env-default:"10"`
	PostgresMinConns          int32         `env:"POSTGRES_MIN_CONNS" env-default:"1"`
	PostgresMaxConnLifetime   time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" env-default:"30m"`
	PostgresMaxConnIdleTime   time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" env-default:"5m"`
	PostgresHealthCheckPeriod time.Duration `env:"POSTGRES_HEALTH_CHECK_PERIOD" env-default:"1m"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env config: %w", err)
	}

	if strings.TrimSpace(cfg.ReportsDir) == "" {
		return Config{}, fmt.Errorf("REPORTS_DIR is empty")
	}
	if cfg.PostgresMaxConns < 1 {
		return Config{}, fmt.Errorf("POSTGRES_MAX_CONNS must be >= 1")
	}
	if cfg.PostgresMinConns < 0 {
		return Config{}, fmt.Errorf("POSTGRES_MIN_CONNS must be >= 0")
	}
	if cfg.PostgresMinConns > cfg.PostgresMaxConns {
		return Config{}, fmt.Errorf("POSTGRES_MIN_CONNS must be <= POSTGRES_MAX_CONNS")
	}

	return cfg, nil
}

func (c Config) PostgresConfig() data.PostgresConfig {
	return data.PostgresConfig{
		DSN:               c.PostgresDSN,
		MaxConns:          c.PostgresMaxConns,
		MinConns:          c.PostgresMinConns,
		MaxConnLifetime:   c.PostgresMaxConnLifetime,
		MaxConnIdleTime:   c.PostgresMaxConnIdleTime,
		HealthCheckPeriod: c.PostgresHealthCheckPeriod,
	}
}
