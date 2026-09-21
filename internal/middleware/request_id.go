package middleware

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const HeaderXRequestID = "X-Request-ID"

// RequestID memastikan setiap HTTP request memiliki ID unik untuk penelusuran (tracing)
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(HeaderXRequestID)
		if reqID == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			reqID = fmt.Sprintf("req-%x-%d", b, time.Now().UnixMilli()%10000)
		}

		c.Set("request_id", reqID)
		c.Writer.Header().Set(HeaderXRequestID, reqID)
		c.Next()
	}
}
