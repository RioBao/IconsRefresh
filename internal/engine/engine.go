package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/crazy-max/IconsRefresh/internal/repair"
)

// TriggerKind describes what initiated an engine run.
type TriggerKind string

const (
	TriggerManualCLI        TriggerKind = "manual_cli"
	TriggerManualTray       TriggerKind = "manual_tray"
	TriggerResolutionChange TriggerKind = "resolution_change"
	TriggerShellRestart     TriggerKind = "shell_restart"
)

// Trigger carries metadata for current and future automation integration.
type Trigger struct {
	Kind   TriggerKind
	Source string
	At     time.Time
}

// RunRequest defines one icon repair execution.
type RunRequest struct {
	Mode    repair.Mode
	DryRun  bool
	Trigger Trigger
}

// RunResult contains both planning and mutation outcomes.
type RunResult struct {
	Mode      repair.Mode     `json:"mode"`
	DryRun    bool            `json:"dry_run"`
	Trigger   Trigger         `json:"trigger"`
	Targets   []repair.Target `json:"targets"`
	Result    repair.Result   `json:"result"`
	StartedAt time.Time       `json:"started_at"`
	EndedAt   time.Time       `json:"ended_at"`
	Error     string          `json:"error,omitempty"`
}

// Hooks enables integration points for monitor automation and telemetry.
//
// Hooks are optional; when set they are called synchronously.
type Hooks struct {
	BeforeRun      func(context.Context, RunRequest)
	AfterRun       func(context.Context, RunRequest, RunResult)
	OnMonitorEvent func(context.Context, MonitorEvent)
}

// Engine orchestrates cache target discovery and repair execution.
type Engine struct {
	hooks Hooks
}

// New creates a repair engine with optional hooks.
func New(hooks Hooks) *Engine {
	return &Engine{hooks: hooks}
}

// Run executes one repair workflow.
func (e *Engine) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	if req.Mode == "" {
		return RunResult{}, errors.New("mode is required")
	}
	if req.Trigger.At.IsZero() {
		req.Trigger.At = time.Now().UTC()
	}

	if e.hooks.BeforeRun != nil {
		e.hooks.BeforeRun(ctx, req)
	}

	out := RunResult{Mode: req.Mode, DryRun: req.DryRun, Trigger: req.Trigger, StartedAt: time.Now().UTC()}

	targets, err := repair.DiscoverCacheTargets()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Error = fmt.Sprintf("discover cache targets: %v", err)
		if e.hooks.AfterRun != nil {
			e.hooks.AfterRun(ctx, req, out)
		}
		return out, fmt.Errorf("discover cache targets: %w", err)
	}

	out.Targets = repair.TargetsForMode(targets, req.Mode)
	if req.DryRun {
		out.EndedAt = time.Now().UTC()
		if e.hooks.AfterRun != nil {
			e.hooks.AfterRun(ctx, req, out)
		}
		return out, nil
	}

	out.Result = repair.DeleteTargetsForMode(req.Mode, out.Targets)
	out.EndedAt = time.Now().UTC()
	if e.hooks.AfterRun != nil {
		e.hooks.AfterRun(ctx, req, out)
	}
	return out, nil
}
