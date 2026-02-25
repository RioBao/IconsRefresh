package engine

import (
	"context"
	"testing"

	"github.com/crazy-max/IconsRefresh/internal/repair"
)

func TestRequestForPreset(t *testing.T) {
	req, err := RequestForPreset(PresetTrayStandard, true)
	if err != nil {
		t.Fatalf("RequestForPreset() error = %v", err)
	}
	if req.Mode != repair.ModeStandard {
		t.Fatalf("mode = %q, want %q", req.Mode, repair.ModeStandard)
	}
	if req.Trigger.Kind != TriggerManualTray {
		t.Fatalf("trigger = %q, want %q", req.Trigger.Kind, TriggerManualTray)
	}
}

func TestNotifyMonitorEventHook(t *testing.T) {
	called := false
	eng := New(Hooks{OnMonitorEvent: func(_ context.Context, event MonitorEvent) {
		called = event.Kind == MonitorEventResolutionChanged
	}})

	eng.NotifyMonitorEvent(context.Background(), MonitorEvent{Kind: MonitorEventResolutionChanged})
	if !called {
		t.Fatal("expected monitor hook to be called")
	}
}
