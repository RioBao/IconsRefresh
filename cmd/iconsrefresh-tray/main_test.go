//go:build windows

package main

import (
	"errors"
	"testing"

	"github.com/crazy-max/IconsRefresh/internal/engine"
	"github.com/crazy-max/IconsRefresh/internal/repair"
	internalwindows "github.com/crazy-max/IconsRefresh/internal/windows"
)

func TestLogLineColor(t *testing.T) {
	tests := []struct {
		line string
		want [4]uint8
	}{
		{line: "✓ ok", want: [4]uint8{colSuccess.R, colSuccess.G, colSuccess.B, colSuccess.A}},
		{line: "✗ fail", want: [4]uint8{colFail.R, colFail.G, colFail.B, colFail.A}},
		{line: "⚠ warn", want: [4]uint8{colWarn.R, colWarn.G, colWarn.B, colWarn.A}},
		{line: "- step", want: [4]uint8{colStep.R, colStep.G, colStep.B, colStep.A}},
		{line: "other", want: [4]uint8{colText2.R, colText2.G, colText2.B, colText2.A}},
	}

	for _, tt := range tests {
		got := logLineColor(tt.line)
		if [4]uint8{got.R, got.G, got.B, got.A} != tt.want {
			t.Fatalf("logLineColor(%q)=%v want %v", tt.line, got, tt.want)
		}
	}
}

func TestBuildResultSummaryAndLines(t *testing.T) {
	r := engine.RunResult{
		Result: repair.Result{
			Paths: []repair.PathResult{
				{Path: `C:\a\IconCache.db`, Deleted: true},
				{Path: `C:\a\iconcache_16.db`, Skipped: true},
				{Path: `C:\a\iconcache_64.db`, Error: "access denied"},
			},
			ExplorerRestart: &internalwindows.ExplorerRestartResult{
				Restarted: true,
				Warning:   "restart warning",
			},
			IE4UInit: &internalwindows.IE4UInitResult{
				Ran:     true,
				Warning: "ie4uinit warning",
			},
			ShellNotify: &internalwindows.ShellNotifyResult{
				AssocChangedSent: true,
				Warning:          "notify warning",
			},
		},
	}

	got := buildResult(r, nil)
	if got.footer != "1 deleted • 1 skipped • 1 failed" {
		t.Fatalf("footer=%q", got.footer)
	}
	assertContains(t, got.logLines, "✓  IconCache.db")
	assertContains(t, got.logLines, "✗  iconcache_64.db: access denied")
	assertContains(t, got.logLines, "-  Explorer restarted")
	assertContains(t, got.logLines, "-  ie4uinit ran")
	assertContains(t, got.logLines, "-  Shell notified")
	assertContains(t, got.logLines, "⚠  restart warning")
	assertContains(t, got.logLines, "⚠  notify warning")
	assertContains(t, got.logLines, "⚠  ie4uinit warning")
}

func TestBuildResultError(t *testing.T) {
	got := buildResult(engine.RunResult{}, errors.New("boom"))
	if got.footer != "Error" {
		t.Fatalf("footer=%q, want Error", got.footer)
	}
	if len(got.logLines) != 1 || got.logLines[0] != "error: boom" {
		t.Fatalf("logLines=%v", got.logLines)
	}
}

func assertContains(t *testing.T, lines []string, want string) {
	t.Helper()
	for _, line := range lines {
		if line == want {
			return
		}
	}
	t.Fatalf("missing line %q in %v", want, lines)
}
