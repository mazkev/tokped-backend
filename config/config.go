package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	MongoURI    string
	DBName      string
	JWTSecret   string
	FrontendURL string
	GinMode     string
}

// LoadConfig membaca environment variables dari .env dengan fallback nilai default
func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:      getEnv("DB_NAME", "tokopedia_db"),
		JWTSecret:   getEnv("JWT_SECRET", "default_secret_key_123"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		GinMode:     getEnv("GIN_MODE", "release"),
	}
}

// Validate memeriksa kelayakan konfigurasi saat startup (Fail-Fast)
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return errors.New("PORT tidak boleh kosong")
	}

	portNum, err := strconv.Atoi(c.Port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return fmt.Errorf("PORT '%s' tidak valid, harus berupa angka antara 1 dan 65535", c.Port)
	}

	if strings.TrimSpace(c.MongoURI) == "" {
		return errors.New("MONGO_URI tidak boleh kosong")
	}

	if strings.TrimSpace(c.DBName) == "" {
		return errors.New("DB_NAME tidak boleh kosong")
	}

	if strings.TrimSpace(c.JWTSecret) == "" {
		return errors.New("JWT_SECRET tidak boleh kosong")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
