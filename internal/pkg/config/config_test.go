package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultsAndEnvOverride(t *testing.T) {
	// Ensure clean env
	_ = os.Unsetenv("GRPC_SERVER_ADDRESS")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg == nil {
		t.Fatalf("nil config")
	}
	if cfg.GRPC.Address != "127.0.0.1:8081" {
		t.Fatalf("default grpc address unexpected: %s", cfg.GRPC.Address)
	}
	// Override via env
	_ = os.Setenv("GRPC_SERVER_ADDRESS", "1.2.3.4:5555")
	_ = os.Setenv("OPENROUTER_ENABLE", "true")
	_ = os.Setenv("OPENROUTER_MODEL", "openai/gpt-4o-mini")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if cfg.GRPC.Address != "1.2.3.4:5555" {
		t.Fatalf("env override not applied: %s", cfg.GRPC.Address)
	}
	if !cfg.OpenRouter.Enable {
		t.Fatalf("openrouter enable env override not applied")
	}
	if cfg.OpenRouter.Model != "openai/gpt-4o-mini" {
		t.Fatalf("openrouter model env override not applied: %s", cfg.OpenRouter.Model)
	}
}
