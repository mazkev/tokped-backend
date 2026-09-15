package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Voucher Model
type Voucher struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Code      string        `bson:"code" json:"code"`
	Discount  int           `bson:"discount" json:"discount"`
	Type      string        `bson:"type" json:"type"` // "flat" atau "percent"
	MinSpend  int           `bson:"min_spend" json:"minSpend"`
	IsActive  bool          `bson:"is_active" json:"isActive"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
}

// Item dalam pesanan
type OrderItem struct {
	ProductID string `bson:"product_id" json:"id"`
	Name      string `bson:"name" json:"name"`
	Price     int    `bson:"price" json:"price"`
	Qty       int    `bson:"qty" json:"qty"`
	Image     string `bson:"image" json:"image"`
	Shop      string `bson:"shop" json:"shop"`
}

// Payment Info untuk simulasi payment gateway
type PaymentInfo struct {
	VANumber      string     `bson:"va_number,omitempty" json:"vaNumber,omitempty"`
	QRCodeData    string     `bson:"qr_code_data,omitempty" json:"qrCodeData,omitempty"`
	ExpiredAt     time.Time  `bson:"expired_at" json:"expiredAt"`
	PaidAt        *time.Time `bson:"paid_at,omitempty" json:"paidAt,omitempty"`
	PaymentStatus string     `bson:"payment_status" json:"paymentStatus"` // "PENDING", "PAID", "EXPIRED"
}

// Order Model
type Order struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	InvoiceNumber string        `bson:"invoice_number" json:"invoiceNumber"` // contoh: INV/20260311/001
	UserID        string        `bson:"user_id" json:"userId"`
	UserName      string        `bson:"user_name" json:"userName"`
	Items         []OrderItem   `bson:"items" json:"items"`
	Total         int           `bson:"total" json:"total"`
	Status        string        `bson:"status" json:"status"` // "Menunggu Pembayaran", "Diproses", "Dikirim", "Selesai"
	PaymentMethod string        `bson:"payment_method" json:"paymentMethod"`
	PaymentInfo   PaymentInfo   `bson:"payment_info" json:"paymentInfo"`
	VoucherUsed   string        `bson:"voucher_used,omitempty" json:"voucherUsed"`
	Reviewed      bool          `bson:"reviewed" json:"reviewed"`
	CreatedAt     time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updatedAt"`
}

// DTO Requests
type CreateOrderRequest struct {
	Items         []OrderItem `json:"items" binding:"required,min=1"`
	PaymentMethod string      `json:"paymentMethod" binding:"required"`
	VoucherCode   string      `json:"voucherCode"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"` // "Diproses", "Dikirim", "Selesai"
}

type ApplyVoucherRequest struct {
	Code string `json:"code" binding:"required"`
}