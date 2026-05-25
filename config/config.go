package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	PostgreSQL `yaml:"postgres"`
}

type HTTPServer struct {
	Port        string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PostgreSQL struct {
	DSN string `yaml:"dsn" env:"DB_DSN" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); err != nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("failed to read config file: %s", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("failed to read config from env: %s", err)
		}
	}

	return &cfg
}
