package main

import (
	"errors"
	"testing"
)

func TestErrInterrupted(t *testing.T) {
	if !errors.Is(errInterrupted, errInterrupted) {
		t.Fatal("errInterrupted sentinel identity broken")
	}
	if errInterrupted.Error() != "interrupted" {
		t.Fatalf("expected message %q, got %q", "interrupted", errInterrupted.Error())
	}
}
