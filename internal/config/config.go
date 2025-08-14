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
	RateLimiterDb int    `mapstructure:"rate-limiter-db"`
}

type RateLimiter struct {
	FixedWindow FixedWindow `mapstructure:"fixed-window"`
}

type FixedWindow struct {
	MaxRequests int `mapstructure:"max-requests"`
	TimeFrameMs int `mapstructure:"time-frame-ms"`
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
