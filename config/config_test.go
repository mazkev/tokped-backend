package config

import (
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg := LoadConfig()
	if cfg.Port == "" {
		t.Errorf("expected default Port, got empty")
	}
	if cfg.DBName == "" {
		t.Errorf("expected default DBName, got empty")
	}
	if cfg.JWTSecret == "" {
		t.Errorf("expected default JWTSecret, got empty")
	}
}

func TestConfig_Validate_Success(t *testing.T) {
	cfg := &Config{
		Port:      "8080",
		MongoURI:  "mongodb://localhost:27017",
		DBName:    "tokopedia_db",
		JWTSecret: "my-secure-jwt-secret-key",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	cfg := &Config{
		Port:      "invalid-port",
		MongoURI:  "mongodb://localhost:27017",
		DBName:    "tokopedia_db",
		JWTSecret: "secret",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatalf("expected error for non-numeric port, got nil")
	}

	cfg.Port = "70000" // out of range
	err = cfg.Validate()
	if err == nil {
		t.Fatalf("expected error for port > 65535, got nil")
	}
}

func TestConfig_Validate_EmptyFields(t *testing.T) {
	cfg := &Config{
		Port:      "8080",
		MongoURI:  "",
		DBName:    "tokopedia_db",
		JWTSecret: "secret",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error on empty MongoURI, got nil")
	}

	cfg.MongoURI = "mongodb://localhost:27017"
	cfg.JWTSecret = ""
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error on empty JWTSecret, got nil")
	}
}
