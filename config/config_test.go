package config

import (
	"testing"
)

func TestShadeEnabledNil(t *testing.T) {
	cfg := Config{} // Shade not set → nil
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
