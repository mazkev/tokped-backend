package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Product struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string        `bson:"name" json:"name"`
	Price         int           `bson:"price" json:"price"`
	OriginalPrice int           `bson:"original_price" json:"originalPrice"`
	Discount      int           `bson:"discount" json:"discount"`
	Image         string        `bson:"image" json:"image"`
	Rating        float64       `bson:"rating" json:"rating"`
	Sold          int           `bson:"sold" json:"sold"`
	Shop          string        `bson:"shop" json:"shop"`
	Location      string        `bson:"location" json:"location"`
	Badge         string        `bson:"badge" json:"badge"` // "official" atau "power-merchant"
	Condition     string        `bson:"condition" json:"condition"` // "Baru" atau "Bekas"
	Category      string        `bson:"category" json:"category"`
	Stock         int           `bson:"stock" json:"stock"`
	CreatedAt     time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updatedAt"`
}

type ProductFilter struct {
	Category  string `form:"category"`
	Search    string `form:"search"`
	MinPrice  int    `form:"min_price"`
	MaxPrice  int    `form:"max_price"`
	Condition string `form:"condition"`
	Location  string `form:"location"`
}

type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Price         int     `json:"price" binding:"required,gt=0"`
	OriginalPrice int     `json:"originalPrice"`
	Discount      int     `json:"discount"`
	Image         string  `json:"image" binding:"required"`
	Rating        float64 `json:"rating"`
	Shop          string  `json:"shop"`
	Location      string  `json:"location"`
	Badge         string  `json:"badge"`
	Condition     string  `json:"condition"`
	Category      string  `json:"category" binding:"required"`
	Stock         int     `json:"stock" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name          string  `json:"name"`
	Price         int     `json:"price"`
	OriginalPrice int     `json:"originalPrice"`
	Discount      int     `json:"discount"`
	Image         string  `json:"image"`
	Rating        float64 `json:"rating"`
	Shop          string  `json:"shop"`
	Location      string  `json:"location"`
	Badge         string  `json:"badge"`
	Condition     string  `json:"condition"`
	Category      string  `json:"category"`
	Stock         int     `json:"stock"`
}
