package main

import (
	"fmt"
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

func TestParseFlagsNoColor(t *testing.T) {
	f, remaining := parseFlags([]string{"--no-color", "get", "pods"})
	if f.colorMode != "never" {
		t.Errorf("--no-color: want colorMode=%q, got %q", "never", f.colorMode)
	}
	if len(remaining) != 2 || remaining[0] != "get" || remaining[1] != "pods" {
		t.Errorf("unexpected remaining args: %v", remaining)
	}
}

func TestParseFlagsColorAlways(t *testing.T) {
	f, _ := parseFlags([]string{"--color=always"})
	if f.colorMode != "always" {
		t.Errorf("--color=always: want colorMode=%q, got %q", "always", f.colorMode)
	}
}

func TestParseFlagsWatchPassthrough(t *testing.T) {
	f, remaining := parseFlags([]string{"--watch", "get", "pods", "-w"})
	if !f.watchMode {
		t.Error("--watch should set watchMode=true")
	}
	want := []string{"get", "pods", "-w"}
	if len(remaining) != len(want) {
		t.Fatalf("remaining: want %v, got %v", want, remaining)
	}
	for i, v := range want {
		if remaining[i] != v {
			t.Errorf("remaining[%d]: want %q, got %q", i, v, remaining[i])
		}
	}
}

func TestParseFlagsUnknownShortFlags(t *testing.T) {
	f, remaining := parseFlags([]string{"-n", "openshift-monitoring", "get", "pods"})
	if f.dryRun || f.showVer || f.watchMode {
		t.Error("no oc-color flags should be set")
	}
	want := []string{"-n", "openshift-monitoring", "get", "pods"}
	if len(remaining) != len(want) {
		t.Fatalf("remaining: want %v, got %v", want, remaining)
	}
	for i, v := range want {
		if remaining[i] != v {
			t.Errorf("remaining[%d]: want %q, got %q", i, v, remaining[i])
		}
	}
}

func TestParseFlagsMixedOcColorAndOc(t *testing.T) {
	f, remaining := parseFlags([]string{"--theme", "nord", "-n", "default", "get", "pods", "-o", "wide"})
	if f.themeName != "nord" {
		t.Errorf("theme: want %q, got %q", "nord", f.themeName)
	}
	want := []string{"-n", "default", "get", "pods", "-o", "wide"}
	if len(remaining) != len(want) {
		t.Fatalf("remaining: want %v, got %v", want, remaining)
	}
	for i, v := range want {
		if remaining[i] != v {
			t.Errorf("remaining[%d]: want %q, got %q", i, v, remaining[i])
		}
	}
}

func TestParseFlagsDoubleDashPassthrough(t *testing.T) {
	f, remaining := parseFlags([]string{"--watch", "--", "-n", "foo", "get", "pods"})
	if !f.watchMode {
		t.Error("--watch should be set")
	}
	want := []string{"--", "-n", "foo", "get", "pods"}
	if len(remaining) != len(want) {
		t.Fatalf("remaining: want %v, got %v", want, remaining)
	}
	for i, v := range want {
		if remaining[i] != v {
			t.Errorf("remaining[%d]: want %q, got %q", i, v, remaining[i])
		}
	}
}

func TestGoInstalled(t *testing.T) {
	found := goInstalled(func(name string) (string, error) {
		return "/usr/local/go/bin/go", nil
	})
	if !found {
		t.Error("goInstalled should return true when lookPath succeeds")
	}

	notFound := goInstalled(func(name string) (string, error) {
		return "", fmt.Errorf("executable file not found in $PATH")
	})
	if notFound {
		t.Error("goInstalled should return false when lookPath fails")
	}
}
