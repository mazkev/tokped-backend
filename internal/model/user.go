package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	Email     string        `bson:"email" json:"email"`
	Password  string        `bson:"password" json:"-"` // Tanda "-" agar password tidak pernah dikirim ke JSON response
	Role      string        `bson:"role" json:"role"`   // "user" atau "admin"
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

// Request DTO (Data Transfer Object)
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
