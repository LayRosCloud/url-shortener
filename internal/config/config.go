package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local" env:"ENV"`
	StoragePath string `yaml:"storage_path" env-required:"true" env:"STORAGE_PATH"`
	HttpServer  `yaml:"http_server"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-required:"true" env:"ADDRESS"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s" env:"TIMEOUT"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"4s" env:"IDLE_TIMEOUT"`
	User        string        `yaml:"user" env-required:"true" env:"USER"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

func MustLoad(configPath string) *Config {
	if configPath == "" {
		log.Fatal("Config path is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file %s is not exists", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Config error loading: %s", err)
	}
	return &cfg
}
