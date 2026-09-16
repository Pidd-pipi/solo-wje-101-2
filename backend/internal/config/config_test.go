package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort = %s", cfg.ServerPort)
	}
	if cfg.DBName != "coffeetaste" {
		t.Errorf("DBName = %s", cfg.DBName)
	}
	if cfg.DSN() == "" {
		t.Error("DSN should not be empty")
	}
}
