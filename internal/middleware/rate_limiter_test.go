package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimiter_Allow(t *testing.T) {
	limiter := NewIPRateLimiter(2, 500*time.Millisecond)

	ip := "192.168.1.100"
	// Request 1: boleh
	if !limiter.Allow(ip) {
		t.Errorf("expected 1st request to be allowed")
	}
	// Request 2: boleh
	if !limiter.Allow(ip) {
		t.Errorf("expected 2nd request to be allowed")
	}
	// Request 3: ditolak (melebihi kapasitas 2)
	if limiter.Allow(ip) {
		t.Errorf("expected 3rd request to be denied")
	}

	// Tunggu window berlalu
	time.Sleep(550 * time.Millisecond)

	// Request 4: boleh lagi setelah window reset
	if !limiter.Allow(ip) {
		t.Errorf("expected request after window reset to be allowed")
	}
}

func TestAuthRateLimiter_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	testLimiter := NewIPRateLimiter(2, time.Second)
	testMiddleware := func(c *gin.Context) {
		if !testLimiter.Allow(c.ClientIP()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak permintaan.",
			})
			return
		}
		c.Next()
	}

	r.GET("/api/auth/login", testMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request 1
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/auth/login", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w1.Code)
	}

	// Request 2
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/auth/login", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	// Request 3 (Harus 429 Too Many Requests)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/auth/login", nil)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w3.Code)
	}
	if w3.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After header: 60")
	}
}
