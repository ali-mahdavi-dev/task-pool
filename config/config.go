package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server   Server
	Database Database
}

type Server struct {
	ServiceName  string        `envconfig:"SERVICE_NAME" default:"task-pool"`
	Host         string        `envconfig:"SERVER_HOST" default:"0.0.0.0"`
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"10s"`
	Debug        bool          `envconfig:"SERVER_DEBUG" default:"false"`
}

type Database struct {
	Host               string `envconfig:"DATABASE_HOST" default:"localhost"`
	Port               int    `envconfig:"DATABASE_PORT" default:"5432"`
	Username           string `envconfig:"DATABASE_USERNAME" default:"postgres"`
	Password           string `envconfig:"DATABASE_PASSWORD" default:"postgres"`
	Name               string `envconfig:"DATABASE_NAME" default:"task_pool"`
	SSLMode            string `envconfig:"DATABASE_SSLMODE" default:"disable"`
	MaxOpenConnections int    `envconfig:"DATABASE_MAX_OPEN_CONNECTION" default:"100"`
}

func Load() (*Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load env variable into config struct: %w", err)
	}

	return &cfg, nil
}
