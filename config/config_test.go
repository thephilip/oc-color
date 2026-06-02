package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShadeEnabledNil(t *testing.T) {
	cfg := Config{}
	if !cfg.ShadeEnabled() {
		t.Error("nil Shade should default to enabled")
	}
}

func TestShadeEnabledFalse(t *testing.T) {
	f := false
	cfg := Config{Shade: &f}
	if cfg.ShadeEnabled() {
		t.Error("Shade=false should return disabled")
	}
}

func TestShadeEnabledTrue(t *testing.T) {
	tr := true
	cfg := Config{Shade: &tr}
	if !cfg.ShadeEnabled() {
		t.Error("Shade=true should return enabled")
	}
}

func TestSaveCreatesFileAndParentDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := Config{Color: "always", Theme: "nord"}
	path, err := Save(cfg)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	expected := filepath.Join(home, ".config", "oc-color", "config.yaml")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file at %s: %v", path, err)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := Config{Color: "always", Theme: "gruvbox"}
	if _, err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if loaded.Theme != "gruvbox" {
		t.Errorf("Theme = %q, want %q", loaded.Theme, "gruvbox")
	}
	if loaded.Color != "always" {
		t.Errorf("Color = %q, want %q", loaded.Color, "always")
	}
}

func TestSaveReturnsResolvedPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := Save(Config{Theme: "nord"})
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if path == "" {
		t.Error("Save() returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Save() returned non-absolute path: %q", path)
	}
}
