package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// UploadImage godoc
// @Summary Unggah Foto Produk (Admin)
// @Description Mengunggah file gambar produk langsung ke storage server (Khusus Admin)
// @Tags Products
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "File Gambar (JPG, PNG, WEBP, maks 5MB)"
// @Success 200 {object} map[string]interface{}
// @Router /products/upload [post]
func (h *ProductHandler) UploadImage(c *gin.Context) {
	// Batasi ukuran upload maksimal 5MB
	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file terlalu besar, maksimal 5MB"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File gambar tidak ditemukan"})
		return
	}

	// Validasi ekstensi file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !validExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung. Gunakan JPG, PNG, atau WEBP."})
		return
	}

	// Buat direktori uploads jika belum ada
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan direktori penyimpanan"})
		return
	}

	// Generate nama file unik
	filename := fmt.Sprintf("prod_%d_%04d%s", time.Now().UnixNano(), rand.Intn(10000), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file gambar ke server"})
		return
	}

	// Buat URL yang bisa diakses publik
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	fullURL := fmt.Sprintf("%s://%s/uploads/%s", scheme, host, filename)

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Gambar berhasil diunggah",
		"filename": filename,
		"imageUrl": fullURL,
		"path":     "/uploads/" + filename,
	})
}