package main

import (
	"fmt"
	"log"

	"github.com/daverussell13/rate-limiter-doitpay-project/internal/config"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/rdb"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/rest"
	middleware "github.com/daverussell13/rate-limiter-doitpay-project/internal/rest/midlleware"
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	rateLimiterRdbClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.RateLimiterDb,
	})

	// Uncomment this 2 line if you want to use in memory store
	// fixedWindowMemoryRepo := memory.NewFixedWindowRepository()
	// fixedWindowSvc := service.NewFixedWindowService(fixedWindowMemoryRepo, cfg.RateLimiter.FixedWindow)

	// Uncomment this 2 line if you want to use redis store
	fixedWindowRedisRepo := rdb.NewFixedWindowRepository(rateLimiterRdbClient)
	fixedWindowSvc := service.NewFixedWindowService(fixedWindowRedisRepo, cfg.RateLimiter.FixedWindow)

	pingHdl := rest.NewPingHandler()

	r := gin.Default()

	r.GET("/apikey/ping", middleware.RateLimit(fixedWindowSvc, func(c *gin.Context) string {
		return c.GetHeader("X-API-Key")
	}), pingHdl.Ping)

	r.GET("/ipaddress/ping", middleware.RateLimit(fixedWindowSvc, func(c *gin.Context) string {
		return c.ClientIP()
	}), pingHdl.Ping)

	addr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
