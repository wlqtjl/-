package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ListenAddr      string
	DatabaseURL     string
	JWTSecret       string
	DeepSeekAPIKey  string
	DeepSeekURL     string
	SiliconFlowKey  string
	SiliconFlowURL  string
	MigrationsPath  string
	AccessTokenTTL  int // minutes
	RefreshTokenTTL int // hours
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:      envOrDefault("LISTEN_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		DeepSeekAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekURL:     envOrDefault("DEEPSEEK_URL", "https://api.deepseek.com/v1/chat/completions"),
		SiliconFlowKey:  os.Getenv("SILICONFLOW_API_KEY"),
		SiliconFlowURL:  envOrDefault("SILICONFLOW_URL", "https://api.siliconflow.cn/v1/audio/speech"),
		MigrationsPath:  envOrDefault("MIGRATIONS_PATH", ""),
		AccessTokenTTL:  envOrDefaultInt("ACCESS_TOKEN_TTL", 15),
		RefreshTokenTTL: envOrDefaultInt("REFRESH_TOKEN_TTL", 168),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.DeepSeekAPIKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY is required")
	}
	if cfg.SiliconFlowKey == "" {
		return nil, fmt.Errorf("SILICONFLOW_API_KEY is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
