package service

import (
	"context"
	"strings"
	"testing"

	"tokped-backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockOrderRepo struct {
	ordersByID map[bson.ObjectID]*model.Order
	vouchers   map[string]*model.Voucher
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		ordersByID: make(map[bson.ObjectID]*model.Order),
		vouchers: map[string]*model.Voucher{
			"HEMAT20": {
				Code:     "HEMAT20",
				Discount: 20000,
				MinSpend: 50000,
				IsActive: true,
			},
		},
	}
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *model.Order) error {
	if order.ID.IsZero() {
		order.ID = bson.NewObjectID()
	}
	m.ordersByID[order.ID] = order
	return nil
}

func (m *mockOrderRepo) FindOrdersByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	var result []model.Order
	for _, o := range m.ordersByID {
		if o.UserID == userID {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (m *mockOrderRepo) FindAllOrders(ctx context.Context) ([]model.Order, error) {
	var result []model.Order
	for _, o := range m.ordersByID {
		result = append(result, *o)
	}
	return result, nil
}

func (m *mockOrderRepo) FindOrderByID(ctx context.Context, id bson.ObjectID) (*model.Order, error) {
	o, exists := m.ordersByID[id]
	if !exists {
		return nil, nil
	}
	return o, nil
}

func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, id bson.ObjectID, status string) error {
	if o, exists := m.ordersByID[id]; exists {
		o.Status = status
		return nil
	}
	return nil
}

func (m *mockOrderRepo) SetOrderReviewed(ctx context.Context, id bson.ObjectID) error {
	return nil
}

func (m *mockOrderRepo) PayOrder(ctx context.Context, id bson.ObjectID) error {
	if o, exists := m.ordersByID[id]; exists {
		o.Status = "Diproses"
		o.PaymentInfo.PaymentStatus = "PAID"
		return nil
	}
	return nil
}

func (m *mockOrderRepo) FindVoucherByCode(ctx context.Context, code string) (*model.Voucher, error) {
	v, exists := m.vouchers[strings.ToUpper(code)]
	if !exists {
		return nil, nil
	}
	return v, nil
}

func (m *mockOrderRepo) FindAllActiveVouchers(ctx context.Context) ([]model.Voucher, error) {
	var list []model.Voucher
	for _, v := range m.vouchers {
		if v.IsActive {
			list = append(list, *v)
		}
	}
	return list, nil
}

func (m *mockOrderRepo) SeedInitialVouchers(ctx context.Context) error {
	return nil
}

func TestCreateOrder_Success(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	req := model.CreateOrderRequest{
		Items: []model.OrderItem{
			{ProductID: "prod-1", Name: "Kopi Arabika", Price: 50000, Qty: 2},
			{ProductID: "prod-2", Name: "Cangkir Keramik", Price: 25000, Qty: 1},
		},
		PaymentMethod: "BCA Virtual Account",
	}

	order, err := svc.CreateOrder(context.Background(), "user-123", "Budi Santoso", req)
	if err != nil {
		t.Fatalf("expected order created, got error: %v", err)
	}

	// Subtotal: 50000*2 + 25000*1 = 125000
	if order.Total != 125000 {
		t.Errorf("expected total 125000, got: %d", order.Total)
	}

	if !strings.HasPrefix(order.InvoiceNumber, "INV/") {
		t.Errorf("expected invoice prefix INV/, got: %s", order.InvoiceNumber)
	}

	if order.Status != "Menunggu Pembayaran" {
		t.Errorf("expected status 'Menunggu Pembayaran', got: %s", order.Status)
	}

	if order.PaymentInfo.VANumber == "" {
		t.Errorf("expected generated VA number")
	}
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	req := model.CreateOrderRequest{
		Items: []model.OrderItem{},
	}

	_, err := svc.CreateOrder(context.Background(), "user-123", "Budi", req)
	if err == nil {
		t.Fatalf("expected error on empty items, got nil")
	}
	if err.Error() != "keranjang belanja kosong" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCreateOrder_InvalidQty(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	req := model.CreateOrderRequest{
		Items: []model.OrderItem{
			{ProductID: "prod-1", Name: "Produk", Price: 10000, Qty: 0},
		},
	}

	_, err := svc.CreateOrder(context.Background(), "user-123", "Budi", req)
	if err == nil {
		t.Fatalf("expected error on zero qty, got nil")
	}
	if err.Error() != "jumlah barang minimal 1" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestCreateOrder_WithVoucherDiscount(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	req := model.CreateOrderRequest{
		Items: []model.OrderItem{
			{ProductID: "prod-1", Name: "Headphone Bluetooth", Price: 100000, Qty: 1},
		},
		PaymentMethod: "QRIS",
		VoucherCode:   "HEMAT20", // Diskon 20.000, min spend 50.000
	}

	order, err := svc.CreateOrder(context.Background(), "user-123", "Budi", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total = 100000 - 20000 = 80000
	if order.Total != 80000 {
		t.Errorf("expected total 80000 after discount, got: %d", order.Total)
	}

	if order.VoucherUsed != "HEMAT20" {
		t.Errorf("expected voucher used HEMAT20, got: %s", order.VoucherUsed)
	}

	// Verifikasi QRIS payload
	if order.PaymentInfo.QRCodeData == "" {
		t.Errorf("expected QRIS qrCodeData to be generated")
	}
}

func TestCreateOrder_VoucherBelowMinSpend(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	req := model.CreateOrderRequest{
		Items: []model.OrderItem{
			{ProductID: "prod-1", Name: "Barang Murah", Price: 20000, Qty: 1}, // < 50000 min spend
		},
		VoucherCode: "HEMAT20",
	}

	_, err := svc.CreateOrder(context.Background(), "user-123", "Budi", req)
	if err == nil {
		t.Fatalf("expected error when subtotal is below min spend, got nil")
	}
	if !strings.Contains(err.Error(), "minimum belanja") {
		t.Errorf("expected min spend error, got: %v", err)
	}
}

func TestPayOrder_Success(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	orderID := bson.NewObjectID()
	order := &model.Order{
		ID:            orderID,
		UserID:        "user-owner",
		Status:        "Menunggu Pembayaran",
		PaymentMethod: "QRIS",
	}
	_ = repo.CreateOrder(context.Background(), order)

	updated, err := svc.PayOrder(context.Background(), orderID.Hex(), "user-owner", false)
	if err != nil {
		t.Fatalf("pay order failed: %v", err)
	}

	if updated.Status != "Diproses" {
		t.Errorf("expected status 'Diproses', got: %s", updated.Status)
	}
}

func TestPayOrder_UnauthorizedUser(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	orderID := bson.NewObjectID()
	order := &model.Order{
		ID:            orderID,
		UserID:        "user-owner",
		Status:        "Menunggu Pembayaran",
		PaymentMethod: "QRIS",
	}
	_ = repo.CreateOrder(context.Background(), order)

	// User lain mencoba membayar (bukan admin)
	_, err := svc.PayOrder(context.Background(), orderID.Hex(), "user-impostor", false)
	if err == nil {
		t.Fatalf("expected error when unauthorized user tries to pay, got nil")
	}
	if !strings.Contains(err.Error(), "akses ditolak") {
		t.Errorf("expected access denied error, got: %v", err)
	}
}

func TestUpdateOrderStatus_ValidAndInvalid(t *testing.T) {
	repo := newMockOrderRepo()
	svc := NewOrderService(repo)

	orderID := bson.NewObjectID()
	order := &model.Order{
		ID:     orderID,
		Status: "Menunggu Pembayaran",
	}
	_ = repo.CreateOrder(context.Background(), order)

	// Valid status
	err := svc.UpdateOrderStatus(context.Background(), orderID.Hex(), "Dikirim")
	if err != nil {
		t.Errorf("expected valid update, got: %v", err)
	}

	// Invalid status
	err = svc.UpdateOrderStatus(context.Background(), orderID.Hex(), "StatusNgawur123")
	if err == nil {
		t.Fatalf("expected error on invalid status, got nil")
	}
	if err.Error() != "status pesanan tidak valid" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}
