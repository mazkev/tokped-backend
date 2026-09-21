package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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

	fmt.Println("🚀 Berhasil terhubung ke MongoDB (Pool: min 5, max 50):", cfg.DBName)
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
