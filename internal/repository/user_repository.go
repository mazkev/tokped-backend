package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tokped-backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	SeedAdmin(ctx context.Context) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	repo := &userRepository{
		collection: db.Collection("users"),
	}

	// Otomatis seed akun admin bawaan Tokopedia jika belum ada
	_ = repo.SeedAdmin(context.Background())

	return repo
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		user.ID = oid
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			if email == "admin@tokopedia.com" {
				_ = r.SeedAdmin(ctx)
				if errRetry := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user); errRetry == nil {
					return &user, nil
				}
			}
			return nil, nil // User tidak ditemukan
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// SeedAdmin membuat akun admin bawaan jika belum ada di database
func (r *userRepository) SeedAdmin(ctx context.Context) error {
	adminEmail := "admin@tokopedia.com"
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": adminEmail})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Admin sudah ada
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Name:     "Admin Tokopedia",
		Email:    adminEmail,
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := r.Create(ctx, admin); err != nil {
		return err
	}
	fmt.Println("👑 Akun default Admin berhasil dibuat: admin@tokopedia.com / admin123")
	return nil
}
