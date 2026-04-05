package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration.
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

	// Multi-model AI support
	AIProvider    string // "deepseek", "openai", "zhipu", "gemma"
	OpenAIAPIKey  string
	OpenAIURL     string
	ZhipuAPIKey   string
	ZhipuURL      string
	GemmaAPIKey   string
	GemmaURL      string

	// Sentiment analysis
	EnableSentiment bool
}

// Load reads configuration from environment variables.
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

		AIProvider:      envOrDefault("AI_PROVIDER", "deepseek"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		OpenAIURL:       envOrDefault("OPENAI_URL", "https://api.openai.com/v1/chat/completions"),
		ZhipuAPIKey:     os.Getenv("ZHIPU_API_KEY"),
		ZhipuURL:        envOrDefault("ZHIPU_URL", "https://open.bigmodel.cn/api/paas/v4/chat/completions"),
		GemmaAPIKey:     os.Getenv("GEMMA_API_KEY"),
		GemmaURL:        envOrDefault("GEMMA_URL", "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"),
		EnableSentiment: envOrDefault("ENABLE_SENTIMENT", "false") == "true",
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
