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
	if cfg.GRPC.Address != "127.0.0.1:8080" {
		t.Fatalf("default grpc address unexpected: %s", cfg.GRPC.Address)
	}
	// Override via env
	_ = os.Setenv("GRPC_SERVER_ADDRESS", "1.2.3.4:5555")
	cfg, err = Load()
	if err != nil { t.Fatalf("reload: %v", err) }
	if cfg.GRPC.Address != "1.2.3.4:5555" {
		t.Fatalf("env override not applied: %s", cfg.GRPC.Address)
	}
}
