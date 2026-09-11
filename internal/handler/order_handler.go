package handler

import (
	"net/http"

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

// CreateOrder: Shopper membuat pesanan baru (Checkout)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Pesanan berhasil dibuat",
		"data":    order,
	})
}

// GetMyOrders: Shopper melihat riwayat pesanannya
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

// GetAllOrders: Admin melihat seluruh pesanan masuk dari semua shopper
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

// GetOrderByID: Detail 1 pesanan
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderService.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   order,
	})
}

// UpdateOrderStatus: Admin mengubah status pesanan ("Diproses", "Dikirim", "Selesai")
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Status pesanan berhasil diperbarui",
	})
}

// ApplyVoucher: Cek potongan kode diskon
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

// GetVouchers: Ambil daftar voucher yang sedang aktif
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
