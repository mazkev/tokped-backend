package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"tokped-backend/config"
	"tokped-backend/internal/handler"
	"tokped-backend/internal/middleware"
	"tokped-backend/internal/repository"
	"tokped-backend/internal/service"
)

func main() {
	// 1. Muat Konfigurasi
	cfg := config.LoadConfig()
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	// 2. Hubungkan ke MongoDB
	db := config.ConnectDB(cfg)

	// 3. Inisialisasi Repository & Service untuk Produk & Review
	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productService)

	orderRepo := repository.NewOrderRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	reviewService := service.NewReviewService(reviewRepo, orderRepo, productRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// 4. Inisialisasi Router Gin
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	_ = r.SetTrustedProxies(nil)

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:5173", "http://127.0.0.1:5173", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Product & Review Microservice",
			"status":  "UP",
		})
	})

	// Rute Produk
	productGroup := r.Group("/api/products")
	{
		productGroup.GET("", productHandler.GetAll)
		productGroup.GET("/:id", productHandler.GetByID)

		adminProduct := productGroup.Group("")
		adminProduct.Use(middleware.AuthMiddleware(cfg.JWTSecret), middleware.AdminOnly())
		{
			adminProduct.POST("", productHandler.Create)
			adminProduct.PUT("/:id", productHandler.Update)
			adminProduct.DELETE("/:id", productHandler.Delete)
		}
	}

	// Rute Review
	reviewGroup := r.Group("/api/reviews")
	{
		reviewGroup.GET("/product/:productId", reviewHandler.GetProductReviews)
		reviewGroup.POST("", middleware.AuthMiddleware(cfg.JWTSecret), reviewHandler.CreateReview)
	}

	// Jalankan di port 8082
	port := os.Getenv("PRODUCT_SERVICE_PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("📦 Product & Review Microservice berjalan di port :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Mematikan Product Microservice...")
}
