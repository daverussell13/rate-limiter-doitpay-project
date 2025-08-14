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
	fixedWindowRdbClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.FixedWindowDb,
	})
	tokenBucketRdbClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.TokenBucketDb,
	})

	// Uncomment this 2 line if you want to use in memory store
	// fixedWindowMemoryRepo := memory.NewFixedWindowRepository()
	// fixedWindowSvc := service.NewFixedWindowService(fixedWindowMemoryRepo, cfg.RateLimiter.FixedWindow)

	// Uncomment this 2 line if you want to use redis store
	fixedWindowRedisRepo := rdb.NewFixedWindowRepository(fixedWindowRdbClient)
	fixedWindowSvc := service.NewFixedWindowService(fixedWindowRedisRepo, cfg.RateLimiter.FixedWindow)

	// Uncomment this 2 line if you want to use in memory store
	// tokenBucketMemoryRepo := memory.NewTokenBucketRepository()
	// tokenBucketSvc := service.NewTokenBucketService(tokenBucketMemoryRepo, cfg.RateLimiter.TokenBucket)

	// Uncomment this 2 line if you want to use redis store
	tokenBucketRedisRepo := rdb.NewTokenBucketRepository(
		tokenBucketRdbClient,
		cfg.RateLimiter.TokenBucket.MaxTokens,
		cfg.RateLimiter.TokenBucket.RefillRate,
	)
	tokenBucketSvc := service.NewTokenBucketService(tokenBucketRedisRepo, cfg.RateLimiter.TokenBucket)

	pingHdl := rest.NewPingHandler()

	r := gin.Default()

	r.GET("/fw/apikey/ping", middleware.RateLimit(fixedWindowSvc, func(c *gin.Context) string {
		return c.GetHeader("X-API-Key")
	}), pingHdl.Ping)

	r.GET("/fw/ipaddress/ping", middleware.RateLimit(fixedWindowSvc, func(c *gin.Context) string {
		return c.ClientIP()
	}), pingHdl.Ping)

	r.GET("/tb/apikey/ping", middleware.RateLimit(tokenBucketSvc, func(c *gin.Context) string {
		return c.GetHeader("X-API-Key")
	}), pingHdl.Ping)

	r.GET("/tb/ipaddress/ping", middleware.RateLimit(tokenBucketSvc, func(c *gin.Context) string {
		return c.ClientIP()
	}), pingHdl.Ping)

	addr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
