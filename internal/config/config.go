package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel    string `yaml:"log_level" env-default:"info"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HttpServer  `yaml:"http_server"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func Load() *Config {
	// TODO rewrite for specifying default CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Couldn't find Config file at %s", configPath)
	}

	var cfg Config
	// TODO replace cleanenv with more popular lib
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Couldn't read config: %s", err)
	}

	return &cfg
}
