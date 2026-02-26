package main

import (
	"errors"
	"testing"

	"github.com/crazy-max/IconsRefresh/internal/engine"
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
	if cfg.Preset != engine.PresetCLIStandard {
		t.Fatalf("preset = %q, want %q", cfg.Preset, engine.PresetCLIStandard)
	}
	if !cfg.DryRun || !cfg.JSON {
		t.Fatalf("expected dry-run and json flags enabled: %+v", cfg)
	}
}
