package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/crazy-max/IconsRefresh/internal/engine"
)

// This command is a tray-oriented entrypoint that uses engine presets.
// A full native tray UI can call runPreset/runMonitorEvent as menu handlers later.
func main() {
	presetFlag := flag.String("preset", "standard", "tray mode preset: quick|standard|deep")
	dryRun := flag.Bool("dry-run", false, "plan run without deleting files")
	jsonOutput := flag.Bool("json", false, "print run result as JSON")
	flag.Parse()

	preset, err := parseTrayPreset(*presetFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	result, err := runPreset(context.Background(), preset, *dryRun)
	if *jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if !*jsonOutput {
		fmt.Printf("tray preset=%s mode=%s dry-run=%t targets=%d\n", preset, result.Mode, result.DryRun, len(result.Targets))
	}
}

func parseTrayPreset(input string) (engine.Preset, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "quick":
		return engine.PresetTrayQuick, nil
	case "standard":
		return engine.PresetTrayStandard, nil
	case "deep":
		return engine.PresetTrayDeep, nil
	default:
		return "", fmt.Errorf("invalid tray preset %q", input)
	}
}

func runPreset(ctx context.Context, preset engine.Preset, dryRun bool) (engine.RunResult, error) {
	eng := engine.New(engine.Hooks{
		OnMonitorEvent: func(context.Context, engine.MonitorEvent) {
			// hook placeholder for monitor-driven tray refresh automation
		},
	})

	req, err := engine.RequestForPreset(preset, dryRun)
	if err != nil {
		return engine.RunResult{}, err
	}
	return eng.Run(ctx, req)
}

// runMonitorEvent is a future integration point where watcher callbacks can choose presets.
func runMonitorEvent(ctx context.Context, eng *engine.Engine, event engine.MonitorEvent, dryRun bool) (engine.RunResult, error) {
	eng.NotifyMonitorEvent(ctx, event)

	switch event.Kind {
	case engine.MonitorEventResolutionChanged:
		return runPreset(ctx, engine.PresetTrayQuick, dryRun)
	case engine.MonitorEventShellHostRestart:
		return runPreset(ctx, engine.PresetTrayStandard, dryRun)
	default:
		return engine.RunResult{}, fmt.Errorf("unsupported monitor event %q", event.Kind)
	}
}
