package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"tokped-backend/internal/model"
	"tokped-backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID, userName string, req model.CreateOrderRequest) (*model.Order, error)
	GetUserOrders(ctx context.Context, userID string) ([]model.Order, error)
	GetAllOrders(ctx context.Context) ([]model.Order, error)
	GetOrderByID(ctx context.Context, idStr string) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, idStr string, status string) error
	ApplyVoucher(ctx context.Context, code string) (*model.Voucher, error)
	GetActiveVouchers(ctx context.Context) ([]model.Voucher, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: orderRepo}
}

func (s *orderService) CreateOrder(ctx context.Context, userID, userName string, req model.CreateOrderRequest) (*model.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("keranjang belanja kosong")
	}

	// 1. Hitung total belanja
	subtotal := 0
	for _, item := range req.Items {
		if item.Qty <= 0 {
			return nil, errors.New("jumlah barang minimal 1")
		}
		subtotal += item.Price * item.Qty
	}

	discount := 0
	voucherCode := ""

	// 2. Jika ada voucher, validasi
	if req.VoucherCode != "" {
		voucher, err := s.orderRepo.FindVoucherByCode(ctx, req.VoucherCode)
		if err != nil {
			return nil, err
		}
		if voucher == nil {
			return nil, errors.New("kode voucher tidak valid atau sudah kadaluwarsa")
		}
		if subtotal < voucher.MinSpend {
			return nil, fmt.Errorf("minimum belanja untuk voucher ini adalah Rp%d", voucher.MinSpend)
		}

		discount = voucher.Discount
		voucherCode = voucher.Code
	}

	finalTotal := subtotal - discount
	if finalTotal < 0 {
		finalTotal = 0
	}

	// 3. Generate nomor invoice unik
	invNumber := fmt.Sprintf("INV/%s/%04d", time.Now().Format("20060102"), rand.Intn(10000))

	order := &model.Order{
		InvoiceNumber: invNumber,
		UserID:        userID,
		UserName:      userName,
		Items:         req.Items,
		Total:         finalTotal,
		PaymentMethod: req.PaymentMethod,
		VoucherUsed:   voucherCode,
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userID string) ([]model.Order, error) {
	return s.orderRepo.FindOrdersByUserID(ctx, userID)
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	return s.orderRepo.FindAllOrders(ctx)
}

func (s *orderService) GetOrderByID(ctx context.Context, idStr string) (*model.Order, error) {
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, errors.New("ID pesanan tidak valid")
	}

	order, err := s.orderRepo.FindOrderByID(ctx, oid)
	if err != nil || order == nil {
		return nil, errors.New("pesanan tidak ditemukan")
	}

	return order, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, idStr string, status string) error {
	validStatuses := map[string]bool{
		"Menunggu Konfirmasi": true,
		"Diproses":            true,
		"Dikirim":             true,
		"Selesai":             true,
		"Dibatalkan":          true,
	}

	if !validStatuses[status] {
		return errors.New("status pesanan tidak valid")
	}

	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return errors.New("ID pesanan tidak valid")
	}

	return s.orderRepo.UpdateOrderStatus(ctx, oid, status)
}

func (s *orderService) ApplyVoucher(ctx context.Context, code string) (*model.Voucher, error) {
	voucher, err := s.orderRepo.FindVoucherByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if voucher == nil {
		return nil, errors.New("voucher tidak ditemukan atau tidak aktif")
	}
	return voucher, nil
}

func (s *orderService) GetActiveVouchers(ctx context.Context) ([]model.Voucher, error) {
	return s.orderRepo.FindAllActiveVouchers(ctx)
}
