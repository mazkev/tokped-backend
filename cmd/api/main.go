package main

// @title Tokopedia Clone API
// @version 1.0
// @description REST API Backend Tokopedia Clone berbasis Golang, Gin, dan MongoDB.
// @termsOfService http://swagger.io/terms/

// @contact.name Mazkev Tech
// @contact.url https://github.com/mazkev
// @contact.email support@tokopedia.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Format: "Bearer <token>"

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tokped-backend/config"
	_ "tokped-backend/docs"
	"tokped-backend/internal/handler"
	"tokped-backend/internal/middleware"
	"tokped-backend/internal/repository"
	"tokped-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 1. Muat Konfigurasi dari .env
	cfg := config.LoadConfig()

	// 2. Set Gin Mode (Release Mode menghemat alokasi memory & stdout logging di server)
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	// 3. Hubungkan ke MongoDB dengan Connection Pooling
	db := config.ConnectDB(cfg)

	// 4. Inisialisasi Database Indexes untuk performa query instan & integritas data
	config.EnsureIndexes(db)

	// 5. Inisialisasi Auth
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	// 6. Inisialisasi Product
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	// 7. Inisialisasi Order & Voucher
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	// 8. Inisialisasi Review
	reviewRepo := repository.NewReviewRepository(db)
	reviewService := service.NewReviewService(reviewRepo, orderRepo, productRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// 9. Inisialisasi Router Gin
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Nonaktifkan warning proxy jika tidak diperlukan
	_ = r.SetTrustedProxies(nil)

	// 10. Konfigurasi CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:5173", "http://127.0.0.1:5173", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 11. Swagger Documentation Endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Checks
	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Tokopedia Backend Monolith API is running!",
		})
	}
	r.GET("/health", healthHandler)
	r.GET("/api/health", healthHandler)

	// Manual Seeder Trigger
	r.GET("/api/seed", func(c *gin.Context) {
		ctx := c.Request.Context()
		_ = userRepo.SeedAdmin(ctx)
		_ = productRepo.SeedInitialProducts(ctx)
		_ = orderRepo.SeedInitialVouchers(ctx)
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Database berhasil di-seed otomatis!",
		})
	})

	// 12. Rute Autentikasi
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.GetProfile)
	}

	// 13. Rute Produk
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

	// 14. Rute Voucher
	voucherGroup := r.Group("/api/vouchers")
	{
		voucherGroup.GET("", orderHandler.GetVouchers)
		voucherGroup.POST("/apply", orderHandler.ApplyVoucher)
	}

	// 15. Rute Order
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

	// 16. Rute Review
	reviewGroup := r.Group("/api/reviews")
	{
		reviewGroup.GET("/product/:productId", reviewHandler.GetProductReviews)
		reviewGroup.POST("", middleware.AuthMiddleware(cfg.JWTSecret), reviewHandler.CreateReview)
	}

	// 17. Konfigurasi HTTP Server dengan Port 8080 & Graceful Shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("MONOLITH_PORT")
	}
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Jalankan server di goroutine
	go func() {
		log.Printf("🚀 Tokopedia Backend Monolith berjalan di port :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n⚠️ Menerima sinyal terminasi, mematikan server secara graceful...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("❌ Kesalahan saat shutdown: %v\n", err)
	}
	fmt.Println("✅ Server Tokopedia Backend berhasil dimatikan dengan aman.")
}