package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	Postgres   `yaml:"postgres"`
	Nats       `yaml:"nats"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}
type Postgres struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     string `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	DBName   string `yaml:"db_name" env-default:"postgres"`
}

type Nats struct {
	ClusterId string `yaml:"cluster_id" env-default:"test-cluster"`
	ClientId  string `yaml:"client_id" env-default:"client-123"`
	Url       string `yaml:"url" env-default:"nats://localhost:4222"`
}

func MustLoad(configPath string) *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	return &cfg
}
