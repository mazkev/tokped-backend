package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"tokped-backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	FindOrdersByUserID(ctx context.Context, userID string) ([]model.Order, error)
	FindAllOrders(ctx context.Context) ([]model.Order, error)
	FindOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status string) error
	SetOrderReviewed(ctx context.Context, id bson.ObjectID) error

	// Voucher queries
	FindVoucherByCode(ctx context.Context, code string) (*model.Voucher, error)
	FindAllActiveVouchers(ctx context.Context) ([]model.Voucher, error)
	SeedInitialVouchers(ctx context.Context) error
}

type orderRepository struct {
	orderCol   *mongo.Collection
	voucherCol *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) OrderRepository {
	repo := &orderRepository{
		orderCol:   db.Collection("orders"),
		voucherCol: db.Collection("vouchers"),
	}

	// Otomatis seed voucher diskon bawaan
	_ = repo.SeedInitialVouchers(context.Background())

	return repo
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = "Menunggu Konfirmasi"
	order.Reviewed = false

	res, err := r.orderCol.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		order.ID = oid
	}
	return nil
}

func (r *orderRepository) FindOrdersByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.orderCol.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

func (r *orderRepository) FindAllOrders(ctx context.Context) ([]model.Order, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.orderCol.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

func (r *orderRepository) FindOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	var order model.Order
	err := r.orderCol.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	_, err := r.orderCol.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *orderRepository) SetOrderReviewed(ctx context.Context, id bson.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"reviewed":   true,
			"updated_at": time.Now(),
		},
	}
	_, err := r.orderCol.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *orderRepository) FindVoucherByCode(ctx context.Context, code string) (*model.Voucher, error) {
	var voucher model.Voucher
	err := r.voucherCol.FindOne(ctx, bson.M{
		"code":      strings.ToUpper(code),
		"is_active": true,
	}).Decode(&voucher)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &voucher, nil
}

func (r *orderRepository) FindAllActiveVouchers(ctx context.Context) ([]model.Voucher, error) {
	cursor, err := r.voucherCol.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var vouchers []model.Voucher
	if err := cursor.All(ctx, &vouchers); err != nil {
		return nil, err
	}
	if vouchers == nil {
		vouchers = []model.Voucher{}
	}
	return vouchers, nil
}

func (r *orderRepository) SeedInitialVouchers(ctx context.Context) error {
	count, err := r.voucherCol.CountDocuments(ctx, bson.M{})
	if err != nil || count > 0 {
		return nil
	}

	initialVouchers := []interface{}{
		model.Voucher{
			Code: "TOKOPEDIA10", Discount: 10000, Type: "flat", MinSpend: 50000, IsActive: true, CreatedAt: time.Now(),
		},
		model.Voucher{
			Code: "HEMAT20", Discount: 20000, Type: "flat", MinSpend: 100000, IsActive: true, CreatedAt: time.Now(),
		},
	}

	_, err = r.voucherCol.InsertMany(ctx, initialVouchers)
	if err == nil {
		fmt.Println("🎟️ Voucher awal berhasil di-seed: TOKOPEDIA10 & HEMAT20")
	}
	return err
}
