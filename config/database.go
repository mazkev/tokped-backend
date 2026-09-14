package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Config struct {
	Port        string
	MongoURI    string
	DBName      string
	JWTSecret   string
	FrontendURL string
	GinMode     string
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
		GinMode:     getEnv("GIN_MODE", "release"),
	}
}

func ConnectDB(cfg *Config) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optimasi connection pool untuk lingkungan VPS hemat resource
	clientOptions := options.Client().
		ApplyURI(cfg.MongoURI).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("❌ Gagal inisialisasi MongoDB client: %v", err)
	}

	// Ping database untuk memastikan koneksi aktif
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ Gagal terhubung ke MongoDB: %v", err)
	}

	fmt.Println("✅ Berhasil terhubung ke MongoDB (Pool: min 5, max 50):", cfg.DBName)
	return client.Database(cfg.DBName)
}

// EnsureIndexes membuat database indexes untuk performa query cepat & integritas data
func EnsureIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Index Unik Email User
	_, _ = db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// 2. Index Produk untuk pencarian & filtering
	_, _ = db.Collection("products").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "category", Value: 1}}},
		{Keys: bson.D{{Key: "price", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})

	// 3. Index Order berdasarkan User ID & waktu order
	_, _ = db.Collection("orders").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})

	// 4. Index Unik Voucher Code
	_, _ = db.Collection("vouchers").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// 5. Index Review berdasarkan Product ID
	_, _ = db.Collection("reviews").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "product_id", Value: 1}},
	})

	fmt.Println("⚡ Database indexes MongoDB berhasil diinisialisasi!")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
