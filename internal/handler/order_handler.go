package handler

import (
	"net/http"

	"tokped-backend/internal/apperror"
	"tokped-backend/internal/model"
	"tokped-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder godoc
// @Summary Buat Pesanan Baru (Checkout)
// @Description Shopper membuat pesanan dari keranjang belanja
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.CreateOrderRequest true "Detail Order"
// @Success 201 {object} model.Order
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userName, _ := c.Get("name")

	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID.(string), userName.(string), req)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Pesanan berhasil dibuat",
		"data":    order,
	})
}

// PayOrder godoc
// @Summary Simulasikan Pembayaran Pesanan (Tokopedia Pay)
// @Description Pengguna melunasi pesanan yang berstatus Menunggu Pembayaran
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID Pesanan"
// @Success 200 {object} model.Order
// @Router /orders/{id}/pay [post]
func (h *OrderHandler) PayOrder(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	order, err := h.orderService.PayOrder(c.Request.Context(), id, userID.(string), isAdmin)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pembayaran berhasil diverifikasi! Pesanan Anda sedang diproses penjual.",
		"data":    order,
	})
}

// GetMyOrders godoc
// @Summary Riwayat Pesanan Saya
// @Description Shopper melihat riwayat pesanan miliknya
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} []model.Order
// @Router /orders/my [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")

	orders, err := h.orderService.GetUserOrders(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(orders),
		"data":   orders,
	})
}

// GetAllOrders godoc
// @Summary Semua Pesanan Masuk (Admin)
// @Description Admin melihat semua transaksi masuk
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} []model.Order
// @Router /orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	orders, err := h.orderService.GetAllOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(orders),
		"data":   orders,
	})
}

// GetOrderByID godoc
// @Summary Detail 1 Pesanan
// @Description Mengambil detail pesanan berdasarkan ID
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID Pesanan"
// @Success 200 {object} model.Order
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   order,
	})
}

// UpdateOrderStatus godoc
// @Summary Update Status Pesanan (Admin)
// @Description Admin mengubah status pesanan (Diproses, Dikirim, Selesai)
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID Order"
// @Param request body model.UpdateOrderStatusRequest true "Status Baru"
// @Success 200 {object} map[string]string
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status pesanan berhasil diperbarui",
	})
}

// ApplyVoucher godoc
// @Summary Cek Kode Voucher Diskon
// @Description Validasi kode promo (TOKOPEDIA10, HEMAT20)
// @Tags Vouchers
// @Accept json
// @Produce json
// @Param request body model.ApplyVoucherRequest true "Kode Voucher"
// @Success 200 {object} model.Voucher
// @Router /vouchers/apply [post]
func (h *OrderHandler) ApplyVoucher(c *gin.Context) {
	var req model.ApplyVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	voucher, err := h.orderService.ApplyVoucher(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   voucher,
	})
}

// GetVouchers godoc
// @Summary Daftar Voucher Aktif
// @Description Mendapatkan voucher promo yang tersedia
// @Tags Vouchers
// @Produce json
// @Success 200 {object} []model.Voucher
// @Router /vouchers [get]
func (h *OrderHandler) GetVouchers(c *gin.Context) {
	vouchers, err := h.orderService.GetActiveVouchers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   vouchers,
	})
}