package main

import (
	"errors"
	"testing"

	"github.com/crazy-max/IconsRefresh/internal/repair"
)

func TestResultError(t *testing.T) {
	testCases := []struct {
		name   string
		result repair.Result
		want   bool
	}{
		{
			name: "no path errors",
			result: repair.Result{Paths: []repair.PathResult{
				{Path: "a", Deleted: true},
				{Path: "b", Skipped: true, Error: "not found"},
			}},
			want: false,
		},
		{
			name: "has delete error",
			result: repair.Result{Paths: []repair.PathResult{
				{Path: "a", Error: "access denied"},
				{Path: "b", Deleted: true},
			}},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := resultError(tc.result)
			if got := err != nil; got != tc.want {
				t.Fatalf("resultError() error=%v, wantError=%v", err, tc.want)
			}
		})
	}
}

func TestIsDeletionFailure(t *testing.T) {
	testCases := []struct {
		name       string
		pathResult repair.PathResult
		want       bool
	}{
		{
			name:       "no error",
			pathResult: repair.PathResult{Path: "a", Deleted: true},
			want:       false,
		},
		{
			name:       "expected missing target",
			pathResult: repair.PathResult{Path: "a", Skipped: true, Error: "not found"},
			want:       false,
		},
		{
			name:       "real delete failure",
			pathResult: repair.PathResult{Path: "a", Found: true, Error: "access denied"},
			want:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDeletionFailure(tc.pathResult); got != tc.want {
				t.Fatalf("isDeletionFailure()=%v, want %v", got, tc.want)
			}
		})
	}
}

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
