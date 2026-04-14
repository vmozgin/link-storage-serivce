package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Env        string
	Storage    Storage
	HTTPServer HTTPServer
}

type Storage struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type HTTPServer struct {
	Address string
}

func MustLoad() *Config {
	cfg := &Config{
		Env: getEnv("ENV", "local"),
		HTTPServer: HTTPServer{
			Address: getEnv("ADDRESS", "localhost:8080"),
		},
		Storage: Storage{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "12345"),
			Name:     getEnv("DB_NAME", "link-storage"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
	}
	if err := cfg.validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}

	return cfg
}

func (c *Config) validate() error {
	if c.Storage.User == "" {
		return errors.New("DB_USER is required")
	}
	if c.Storage.Password == "" {
		return errors.New("DB_PASSWORD is required")
	}
	if c.Storage.Name == "" {
		return errors.New("DB_NAME is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func (s *Storage) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		s.Host, s.Port, s.User, s.Password, s.Name, s.SSLMode,
	)
}
