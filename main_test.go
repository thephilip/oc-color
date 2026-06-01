package main

import (
	"testing"
)

func TestParseFlagsNoShade(t *testing.T) {
	f, remaining := parseFlags([]string{"--no-shade", "get", "pods"})
	if !f.noShade {
		t.Error("--no-shade should set noShade=true")
	}
	if len(remaining) != 2 || remaining[0] != "get" || remaining[1] != "pods" {
		t.Errorf("unexpected remaining args: %v", remaining)
	}
}
