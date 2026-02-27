//go:build windows

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/crazy-max/IconsRefresh/internal/engine"
)

func execPreset(preset engine.Preset) runResult {
	eng := engine.New(engine.Hooks{})
	req, err := engine.RequestForPreset(preset, false)
	if err != nil {
		return runResult{logLines: []string{"error: " + err.Error()}, footer: "Error"}
	}
	r, err := eng.Run(context.Background(), req)
	return buildResult(r, err)
}

func buildResult(r engine.RunResult, err error) runResult {
	if err != nil {
		return runResult{logLines: []string{"error: " + err.Error()}, footer: "Error"}
	}

	var lines []string
	var deleted, skipped, failed int

	for _, p := range r.Result.Paths {
		switch {
		case p.Deleted:
			deleted++
			lines = append(lines, "✓  "+filepath.Base(p.Path))
		case p.Skipped:
			skipped++
		case p.Error != "":
			failed++
			lines = append(lines, "✗  "+filepath.Base(p.Path)+": "+p.Error)
		}
	}

	if r.Result.ExplorerRestart != nil && r.Result.ExplorerRestart.Restarted {
		lines = append(lines, "-  Explorer restarted")
	}
	if r.Result.IE4UInit != nil && r.Result.IE4UInit.Ran {
		lines = append(lines, "-  ie4uinit ran")
	}
	if r.Result.ShellNotify != nil && r.Result.ShellNotify.AssocChangedSent {
		lines = append(lines, "-  Shell notified")
	}
	for _, w := range collectWarnings(r) {
		lines = append(lines, "⚠  "+w)
	}

	parts := []string{fmt.Sprintf("%d deleted", deleted)}
	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}

	return runResult{logLines: lines, footer: strings.Join(parts, " • ")}
}

func collectWarnings(r engine.RunResult) []string {
	var ws []string
	if r.Result.ExplorerRestart != nil && r.Result.ExplorerRestart.Warning != "" {
		ws = append(ws, r.Result.ExplorerRestart.Warning)
	}
	if r.Result.ShellNotify != nil && r.Result.ShellNotify.Warning != "" {
		ws = append(ws, r.Result.ShellNotify.Warning)
	}
	if r.Result.IE4UInit != nil && r.Result.IE4UInit.Warning != "" {
		ws = append(ws, r.Result.IE4UInit.Warning)
	}
	return ws
}
