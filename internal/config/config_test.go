package config

import (
	"os"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"DATABASE_URL", "JWT_SECRET", "DEEPSEEK_API_KEY", "SILICONFLOW_API_KEY",
		"LISTEN_ADDR", "DEEPSEEK_URL", "SILICONFLOW_URL", "ACCESS_TOKEN_TTL",
		"REFRESH_TOKEN_TTL", "MIGRATIONS_PATH", "AI_PROVIDER", "OPENAI_API_KEY",
		"OPENAI_URL", "ZHIPU_API_KEY", "ZHIPU_URL", "ENABLE_SENTIMENT",
	} {
		os.Unsetenv(key)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()
	os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-at-least-32-chars!!")
	os.Setenv("DEEPSEEK_API_KEY", "sk-test-deep")
	os.Setenv("SILICONFLOW_API_KEY", "sk-test-silicon")
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	clearEnv(t)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv(t)
	os.Setenv("DATABASE_URL", "postgres://x")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
}

func TestLoad_ShortJWTSecret(t *testing.T) {
	clearEnv(t)
	os.Setenv("DATABASE_URL", "postgres://x")
	os.Setenv("JWT_SECRET", "short")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short JWT_SECRET")
	}
}

func TestLoad_MissingDeepSeekKey(t *testing.T) {
	clearEnv(t)
	os.Setenv("DATABASE_URL", "postgres://x")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-at-least-32-chars!!")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DEEPSEEK_API_KEY")
	}
}

func TestLoad_MissingSiliconFlowKey(t *testing.T) {
	clearEnv(t)
	os.Setenv("DATABASE_URL", "postgres://x")
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-at-least-32-chars!!")
	os.Setenv("DEEPSEEK_API_KEY", "sk-test")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing SILICONFLOW_API_KEY")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	clearEnv(t)
	setRequiredEnv(t)
	os.Setenv("LISTEN_ADDR", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q, want :9090", cfg.ListenAddr)
	}
	if cfg.AIProvider != "deepseek" {
		t.Errorf("AIProvider = %q, want deepseek", cfg.AIProvider)
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	clearEnv(t)
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("default ListenAddr = %q, want :8080", cfg.ListenAddr)
	}
	if cfg.DeepSeekURL != "https://api.deepseek.com/v1/chat/completions" {
		t.Errorf("default DeepSeekURL = %q", cfg.DeepSeekURL)
	}
	if cfg.AccessTokenTTL != 15 {
		t.Errorf("default AccessTokenTTL = %d, want 15", cfg.AccessTokenTTL)
	}
	if cfg.RefreshTokenTTL != 168 {
		t.Errorf("default RefreshTokenTTL = %d, want 168", cfg.RefreshTokenTTL)
	}
}
