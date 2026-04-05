package config

import (
	"os"
	"testing"
)

func TestLoad_MissingDatabaseURL(t *testing.T) {
	clearEnv(t)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
}

func TestLoad_ShortJWTSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "tooshort")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for short JWT_SECRET")
	}
}

func TestLoad_MissingDeepSeekKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "a]veryLong$ecretKey!thatIsAtLeast32chars")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DEEPSEEK_API_KEY")
	}
}

func TestLoad_MissingSiliconFlowKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "a]veryLong$ecretKey!thatIsAtLeast32chars")
	t.Setenv("DEEPSEEK_API_KEY", "sk-test")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing SILICONFLOW_API_KEY")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "a]veryLong$ecretKey!thatIsAtLeast32chars")
	t.Setenv("DEEPSEEK_API_KEY", "sk-test")
	t.Setenv("SILICONFLOW_API_KEY", "sk-test-sf")
	t.Setenv("LISTEN_ADDR", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q; want %q", cfg.ListenAddr, ":9090")
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("DatabaseURL mismatch")
	}
	if cfg.AccessTokenTTL != 15 {
		t.Errorf("AccessTokenTTL = %d; want 15", cfg.AccessTokenTTL)
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	clearEnv(t)
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "a]veryLong$ecretKey!thatIsAtLeast32chars")
	t.Setenv("DEEPSEEK_API_KEY", "sk-test")
	t.Setenv("SILICONFLOW_API_KEY", "sk-test-sf")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8080" {
		t.Errorf("default ListenAddr = %q; want %q", cfg.ListenAddr, ":8080")
	}
	if cfg.DeepSeekURL != "https://api.deepseek.com/v1/chat/completions" {
		t.Errorf("default DeepSeekURL mismatch")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"LISTEN_ADDR", "DATABASE_URL", "JWT_SECRET",
		"DEEPSEEK_API_KEY", "DEEPSEEK_URL",
		"SILICONFLOW_API_KEY", "SILICONFLOW_URL",
		"MIGRATIONS_PATH", "ACCESS_TOKEN_TTL", "REFRESH_TOKEN_TTL",
	} {
		os.Unsetenv(key)
	}
}
