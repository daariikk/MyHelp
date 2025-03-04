package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string `yaml:"env" env-default:"local"`
	DatabaseBaseUrl string `yaml:"database_connection_url" env-required:"true"`
	HTTPServer      `yaml:"http_server"`
	JWT             `yaml:"jwt"`
	Services        `yaml:"services"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type JWT struct {
	AccessSecretKey  string        `yaml:"access_secret_key" env-required:"true"`
	RefreshSecretKey string        `yaml:"refresh_secret_key" env-required:"true"`
	ExpireAccess     time.Duration `yaml:"expire_access" env-default:"30m"`
	ExpireRefresh    time.Duration `yaml:"expire_refresh" env-default:"24h"`
}

type Services struct {
	AccountService     string `yaml:"account_service" env-required:"true"`
	AppointmentService string `yaml:"appointment_service" env-required:"true"`
	PolyclinicService  string `yaml:"polyclinic_service" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("API_GATEWAY_CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
