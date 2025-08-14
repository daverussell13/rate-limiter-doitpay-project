package config

import (
	"github.com/daverussell13/rate-limiter-doitpay-project/internal/domain"
	"github.com/spf13/viper"
)

type Config struct {
	Server      Server      `mapstructure:"server"`
	Redis       Redis       `mapstructure:"redis"`
	RateLimiter RateLimiter `mapstructure:"rate-limiter"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Redis struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	FixedWindowDb int    `mapstructure:"fixed-window-db"`
	TokenBucketDb int    `mapstructure:"token-bucket-db"`
}

type RateLimiter struct {
	FixedWindow FixedWindow `mapstructure:"fixed-window"`
	TokenBucket TokenBucket `mapstructure:"token-bucket"`
}

type FixedWindow struct {
	MaxRequests int `mapstructure:"max-requests"`
	TimeFrameMs int `mapstructure:"time-frame-ms"`
}

type TokenBucket struct {
	MaxTokens  float64 `mapstructure:"max-tokens"`
	RefillRate float64 `mapstructure:"refill-rate"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, domain.WrapError(err, domain.ErrNotFound, "config file not found")
		}
		return nil, domain.WrapError(err, domain.ErrUnknown, "failed to read config file")
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, domain.WrapError(err, domain.ErrUnknown, "failed to unmarshal config")
	}
	return &config, nil
}
