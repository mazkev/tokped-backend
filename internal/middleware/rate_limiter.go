package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type clientRecord struct {
	tokens     int
	lastRefill time.Time
}

// IPRateLimiter mengimplementasikan token bucket rate limiter per IP address
type IPRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientRecord
	capacity int
	window   time.Duration
}

// NewIPRateLimiter membuat rate limiter baru
func NewIPRateLimiter(capacity int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:  make(map[string]*clientRecord),
		capacity: capacity,
		window:   window,
	}

	// Rutin pembersihan memori setiap 5 menit untuk menghapus IP yang tidak aktif
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			limiter.cleanup(10 * time.Minute)
		}
	}()

	return limiter
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	rec, exists := l.clients[ip]
	if !exists {
		l.clients[ip] = &clientRecord{
			tokens:     l.capacity - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens jika window telah terlewati
	elapsed := now.Sub(rec.lastRefill)
	if elapsed >= l.window {
		rec.tokens = l.capacity
		rec.lastRefill = now
	}

	if rec.tokens > 0 {
		rec.tokens--
		return true
	}

	return false
}

func (l *IPRateLimiter) cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, rec := range l.clients {
		if now.Sub(rec.lastRefill) > maxAge {
			delete(l.clients, ip)
		}
	}
}

// GlobalRateLimiter membatasi request umum: 120 request/menit per IP
func GlobalRateLimiter() gin.HandlerFunc {
	limiter := NewIPRateLimiter(120, time.Minute)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak permintaan. Silakan tunggu beberapa saat lagi.",
			})
			return
		}
		c.Next()
	}
}

// AuthRateLimiter membatasi rute autentikasi sensitif: 10 request/menit per IP
func AuthRateLimiter() gin.HandlerFunc {
	limiter := NewIPRateLimiter(10, time.Minute)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak percobaan autentikasi. Silakan tunggu 1 menit sebelum mencoba lagi.",
			})
			return
		}
		c.Next()
	}
}
