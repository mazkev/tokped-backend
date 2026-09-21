package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID_GeneratesNewHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		reqID, exists := c.Get("request_id")
		if !exists || reqID == "" {
			t.Errorf("expected request_id in context")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	respHeader := w.Header().Get(HeaderXRequestID)
	if respHeader == "" {
		t.Errorf("expected X-Request-ID in response header")
	}
}

func TestRequestID_PreservesExistingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderXRequestID, "custom-trace-12345")
	r.ServeHTTP(w, req)

	respHeader := w.Header().Get(HeaderXRequestID)
	if respHeader != "custom-trace-12345" {
		t.Errorf("expected custom-trace-12345, got %s", respHeader)
	}
}
