package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// ReverseProxy meneruskan request yang masuk ke target service URL
func ReverseProxy(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("URL target proxy tidak valid: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Custom error handler jika microservice tujuan offline
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("⚠️ [Gateway Proxy Error] Target %s unreachable: %v", target, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "Layanan sedang tidak dapat dijangkau (502 Bad Gateway)"}`))
	}

	return func(c *gin.Context) {
		c.Request.Host = targetURL.Host
		c.Request.URL.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
