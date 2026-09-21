package handler

import (
	"net/http"

	"tokped-backend/internal/model"
	"tokped-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// CreateReview godoc
// @Summary Kirim Ulasan Produk
// @Description Shopper memberikan rating dan ulasan pesanan yang telah selesai
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.CreateReviewRequest true "Data Review"
// @Success 201 {object} model.Review
// @Router /reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userName, _ := c.Get("name")

	var req model.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.reviewService.CreateReview(c.Request.Context(), userID.(string), userName.(string), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Terima kasih atas ulasanmu!",
		"data":    review,
	})
}

// GetProductReviews godoc
// @Summary Daftar Ulasan Produk
// @Description Mengambil seluruh ulasan untuk produk tertentu
// @Tags Reviews
// @Produce json
// @Param productId path string true "ID Produk"
// @Success 200 {object} []model.Review
// @Router /reviews/product/{productId} [get]
func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productID := c.Param("productId")
	reviews, err := h.reviewService.GetProductReviews(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(reviews),
		"data":   reviews,
	})
}
