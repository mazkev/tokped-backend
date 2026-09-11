package main

import (
	"fmt"
	"net/http"

	"tokped-backend/config"
	"tokped-backend/internal/handler"
	"tokped-backend/internal/middleware"
	"tokped-backend/internal/repository"
	"tokped-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Muat Konfigurasi dari .env
	cfg := config.LoadConfig()

	// 2. Hubungkan ke MongoDB
	db := config.ConnectDB(cfg)

	// 3. Inisialisasi Auth
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	// 4. Inisialisasi Product
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	// 5. Inisialisasi Order & Voucher
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	// 6. Inisialisasi Review
	reviewRepo := repository.NewReviewRepository(db)
	reviewService := service.NewReviewService(reviewRepo, orderRepo, productRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// 7. Inisialisasi Router Gin
	r := gin.Default()

	// 8. Konfigurasi CORS (siap untuk domain Vercel & testing lokal)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:5173", "http://127.0.0.1:5173", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 9. Health Check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Tokopedia Backend API is running!",
		})
	})

	// 10. Rute Autentikasi
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.GetProfile)
	}

	// 11. Rute Produk
	productGroup := r.Group("/api/products")
	{
		productGroup.GET("", productHandler.GetAll)
		productGroup.GET("/:id", productHandler.GetByID)

		// Admin Only: Manajemen Katalog
		adminProduct := productGroup.Group("")
		adminProduct.Use(middleware.AuthMiddleware(cfg.JWTSecret), middleware.AdminOnly())
		{
			adminProduct.POST("", productHandler.Create)
			adminProduct.PUT("/:id", productHandler.Update)
			adminProduct.DELETE("/:id", productHandler.Delete)
		}
	}

	// 12. Rute Voucher
	voucherGroup := r.Group("/api/vouchers")
	{
		voucherGroup.GET("", orderHandler.GetVouchers)
		voucherGroup.POST("/apply", orderHandler.ApplyVoucher)
	}

	// 13. Rute Order
	orderGroup := r.Group("/api/orders")
	orderGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		orderGroup.POST("", orderHandler.CreateOrder)       // Shopper checkout
		orderGroup.GET("/my", orderHandler.GetMyOrders)     // Riwayat order shopper
		orderGroup.GET("/:id", orderHandler.GetOrderByID)   // Detail order

		// Admin Only: Manajemen Status Pesanan
		adminOrder := orderGroup.Group("")
		adminOrder.Use(middleware.AdminOnly())
		{
			adminOrder.GET("", orderHandler.GetAllOrders)                   // Semua pesanan masuk
			adminOrder.PATCH("/:id/status", orderHandler.UpdateOrderStatus) // Ubah status pesanan
		}
	}

	// 14. Rute Review
	reviewGroup := r.Group("/api/reviews")
	{
		reviewGroup.GET("/product/:productId", reviewHandler.GetProductReviews)
		reviewGroup.POST("", middleware.AuthMiddleware(cfg.JWTSecret), reviewHandler.CreateReview)
	}

	// 15. Jalankan Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("🚀 Tokopedia Backend API berjalan di http://localhost%s\n", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		fmt.Printf("❌ Gagal menjalankan server: %v\n", err)
	}
}
