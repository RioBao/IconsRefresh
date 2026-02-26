package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/crazy-max/IconsRefresh/internal/repair"
)

// Preset defines mode defaults that callers (CLI, tray, monitor) can reuse.
type Preset string

const (
	PresetCLIQuick     Preset = "cli-quick"
	PresetCLISoft      Preset = "cli-soft"
	PresetCLIStandard  Preset = "cli-standard"
	PresetCLIDeep      Preset = "cli-deep"
	PresetTrayQuick    Preset = "tray-quick"
	PresetTraySoft     Preset = "tray-soft"
	PresetTrayStandard Preset = "tray-standard"
	PresetTrayDeep     Preset = "tray-deep"
)

// RequestForPreset returns a run request populated with mode and trigger defaults.
func RequestForPreset(preset Preset, dryRun bool) (RunRequest, error) {
	now := time.Now().UTC()
	switch preset {
	case PresetCLIQuick:
		return RunRequest{Mode: repair.ModeQuick, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualCLI, Source: "cmd/iconsrefresh", At: now}}, nil
	case PresetCLISoft:
		return RunRequest{Mode: repair.ModeSoft, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualCLI, Source: "cmd/iconsrefresh", At: now}}, nil
	case PresetCLIStandard:
		return RunRequest{Mode: repair.ModeStandard, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualCLI, Source: "cmd/iconsrefresh", At: now}}, nil
	case PresetCLIDeep:
		return RunRequest{Mode: repair.ModeDeep, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualCLI, Source: "cmd/iconsrefresh", At: now}}, nil
	case PresetTrayQuick:
		return RunRequest{Mode: repair.ModeQuick, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualTray, Source: "cmd/iconsrefresh-tray", At: now}}, nil
	case PresetTraySoft:
		return RunRequest{Mode: repair.ModeSoft, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualTray, Source: "cmd/iconsrefresh-tray", At: now}}, nil
	case PresetTrayStandard:
		return RunRequest{Mode: repair.ModeStandard, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualTray, Source: "cmd/iconsrefresh-tray", At: now}}, nil
	case PresetTrayDeep:
		return RunRequest{Mode: repair.ModeDeep, DryRun: dryRun, Trigger: Trigger{Kind: TriggerManualTray, Source: "cmd/iconsrefresh-tray", At: now}}, nil
	default:
		return RunRequest{}, fmt.Errorf("unknown preset %q", preset)
	}
}

// MonitorEventKind models future automation triggers from OS watchers.
type MonitorEventKind string

const (
	MonitorEventResolutionChanged MonitorEventKind = "resolution_changed"
	MonitorEventShellHostRestart  MonitorEventKind = "shell_host_restart"
)

// MonitorEvent is emitted by watcher implementations.
type MonitorEvent struct {
	Kind MonitorEventKind
	At   time.Time
}

// Watcher is a future extension point for monitor automation.
type Watcher interface {
	Name() string
	Start(context.Context, func(MonitorEvent)) error
}

// NotifyMonitorEvent dispatches monitor events to hooks.
func (e *Engine) NotifyMonitorEvent(ctx context.Context, event MonitorEvent) {
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	if e.hooks.OnMonitorEvent != nil {
		e.hooks.OnMonitorEvent(ctx, event)
	}
}
