package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Config struct {
	Port        string
	MongoURI    string
	DBName      string
	JWTSecret   string
	FrontendURL string
}

func LoadConfig() *Config {
	// Membaca file .env jika tersedia
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:      getEnv("DB_NAME", "tokopedia_db"),
		JWTSecret:   getEnv("JWT_SECRET", "default_secret_key_123"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func ConnectDB(cfg *Config) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("❌ Gagal inisialisasi MongoDB client: %v", err)
	}

	// Ping database untuk memastikan koneksi aktif
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ Gagal terhubung ke MongoDB: %v", err)
	}

	fmt.Println("✅ Berhasil terhubung ke MongoDB:", cfg.DBName)
	return client.Database(cfg.DBName)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
