package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Review struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID   string        `bson:"order_id" json:"orderId"`
	ProductID string        `bson:"product_id" json:"productId"`
	UserID    string        `bson:"user_id" json:"userId"`
	UserName  string        `bson:"user_name" json:"userName"`
	Rating    int           `bson:"rating" json:"rating"` // 1 - 5
	Comment   string        `bson:"comment" json:"comment"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
}

type CreateReviewRequest struct {
	OrderID   string `json:"orderId" binding:"required"`
	ProductID string `json:"productId" binding:"required"`
	Rating    int    `json:"rating" binding:"required,min=1,max=5"`
	Comment   string `json:"comment" binding:"required"`
}
