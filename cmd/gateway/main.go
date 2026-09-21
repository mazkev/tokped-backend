package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// Import package proxy buatan Anda (sesuaikan dengan nama module di go.mod)
	"tokped-backend/internal/proxy"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 1. Centralized CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://tokopedia-react.vercel.app", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Alamat service tujuan
	// monolithURL := getEnv("MONOLITH_SERVICE_URL", "http://localhost:8081")
	authServiceURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	orderServiceURL := getEnv("ORDER_SERVICE_URL", "http://localhost:8081")
	// Produk & Review diarahkan ke Microservice mandiri di port 8082!
	productServiceURL := getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082")

	// 2. Health check Gateway
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "API Gateway is Running"})
	})

	// 3. Routing ke masing-masing Microservice
	api := r.Group("/api")
	{
		// Auth Service (Port 8081)
		api.Any("/auth", proxy.ReverseProxy(authServiceURL))
		api.Any("/auth/*path", proxy.ReverseProxy(authServiceURL))

		// Product & Review Service (Port 8082)
		api.Any("/products", proxy.ReverseProxy(productServiceURL))
		api.Any("/products/*path", proxy.ReverseProxy(productServiceURL))
		api.Any("/reviews", proxy.ReverseProxy(productServiceURL))
		api.Any("/reviews/*path", proxy.ReverseProxy(productServiceURL))

				// Order & Voucher Service (Port 8081)
		api.Any("/orders", proxy.ReverseProxy(orderServiceURL))
		api.Any("/orders/*path", proxy.ReverseProxy(orderServiceURL))
		api.Any("/vouchers", proxy.ReverseProxy(orderServiceURL))
		api.Any("/vouchers/*path", proxy.ReverseProxy(orderServiceURL))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Ubah jadi 8081
	}
	r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
