package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Midtrans MidtransConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type MidtransConfig struct {
	ServerKey string
	ClientKey string
	Env       string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		// ignore missing .env in production (env injected by system)
	}

	required := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"JWT_SECRET":  os.Getenv("JWT_SECRET"),
	}
	for key, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", key)
		}
	}

	jwtExpiryHours := 24
	if raw := os.Getenv("JWT_EXPIRE_HOURS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			jwtExpiryHours = parsed
		}
	}

	return &Config{
		App: AppConfig{
			Port: os.Getenv("APP_PORT"),
			Env:  os.Getenv("APP_ENV"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:      os.Getenv("JWT_SECRET"),
			ExpiryHours: jwtExpiryHours,
		},
		Midtrans: MidtransConfig{
			ServerKey: os.Getenv("MIDTRANS_SERVER_KEY"),
			ClientKey: os.Getenv("MIDTRANS_CLIENT_KEY"),
			Env:       os.Getenv("MIDTRANS_ENV"),
		},
	}, nil
}

func (c *Config) MysqlDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}
