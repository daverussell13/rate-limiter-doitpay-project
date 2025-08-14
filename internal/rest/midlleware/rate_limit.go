package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Allow(ctx context.Context, clientID string) (bool, error)
}

func RateLimit(rateLimiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := keyFunc(c)
		if clientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing key"})
			c.Abort()
			return
		}

		allowed, err := rateLimiter.Allow(c.Request.Context(), clientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal rate limiter error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
