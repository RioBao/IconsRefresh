package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/crazy-max/IconsRefresh/internal/repair"
)

func TestParseArgs_RequiresMode(t *testing.T) {
	_, err := parseArgs(nil)
	if err == nil {
		t.Fatal("expected error when mode is missing")
	}
	if !errors.Is(err, errUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestParseArgs_InvalidMode(t *testing.T) {
	_, err := parseArgs([]string{"oops"})
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !errors.Is(err, errUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestParseArgs_ParsesValidMode(t *testing.T) {
	cfg, err := parseArgs([]string{"--dry-run", "--json", "standard"})
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}
	if cfg.Mode != repair.ModeStandard {
		t.Fatalf("mode = %q, want %q", cfg.Mode, repair.ModeStandard)
	}
	if !cfg.DryRun || !cfg.JSON {
		t.Fatalf("expected dry-run and json flags enabled: %+v", cfg)
	}
}

func TestDeleteFailureError_IgnoresSkippedAndSuccess(t *testing.T) {
	err := deleteFailureError(repair.Result{Paths: []repair.PathResult{
		{Path: "a", Deleted: true},
		{Path: "b", Skipped: true, Error: "not found"},
	}})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDeleteFailureError_ReturnsErrorForDeletionFailures(t *testing.T) {
	err := deleteFailureError(repair.Result{Paths: []repair.PathResult{
		{Path: "a", Error: "access denied"},
		{Path: "b", Deleted: true},
		{Path: "c", Error: "file in use"},
	}})
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "failed to delete 2 target(s)") || !strings.Contains(got, "a: access denied") || !strings.Contains(got, "c: file in use") {
		t.Fatalf("unexpected error message: %q", got)
	}
}
