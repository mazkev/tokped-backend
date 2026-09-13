package handler

import (
	"net/http"

	"tokped-backend/internal/model"
	"tokped-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// GetAll godoc
// @Summary Daftar Katalog Produk
// @Description Mengambil list produk dengan filter kategori, harga, dan pencarian
// @Tags Products
// @Produce json
// @Param category query string false "Filter Kategori"
// @Param search query string false "Kata Kunci Nama"
// @Param min_price query int false "Harga Minimum"
// @Param max_price query int false "Harga Maksimum"
// @Success 200 {object} map[string]interface{}
// @Router /products [get]
func (h *ProductHandler) GetAll(c *gin.Context) {
	var filter model.ProductFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	products, err := h.productService.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  len(products),
		"data":   products,
	})
}

// GetByID godoc
// @Summary Detail Produk
// @Description Mengambil detail 1 produk berdasarkan ID
// @Tags Products
// @Produce json
// @Param id path string true "ID Produk"
// @Success 200 {object} model.Product
// @Failure 404 {object} map[string]string
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	product, err := h.productService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   product,
	})
}

// Create godoc
// @Summary Tambah Produk Baru (Admin)
// @Description Menambahkan produk baru ke katalog (Khusus Admin)
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body model.CreateProductRequest true "Data Produk"
// @Success 201 {object} model.Product
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Produk berhasil ditambahkan",
		"data":    product,
	})
}

// Update godoc
// @Summary Update Produk (Admin)
// @Description Memperbarui data produk/stok (Khusus Admin)
// @Tags Products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID Produk"
// @Param request body model.UpdateProductRequest true "Update Data"
// @Success 200 {object} model.Product
// @Router /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Produk berhasil diperbarui",
		"data":    product,
	})
}

// Delete godoc
// @Summary Hapus Produk (Admin)
// @Description Menghapus produk dari katalog (Khusus Admin)
// @Tags Products
// @Security BearerAuth
// @Param id path string true "ID Produk"
// @Success 200 {object} map[string]string
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.productService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Produk berhasil dihapus",
	})
}
