//go:build !windows

package windows

import "testing"

func TestNotifyShellRefresh_NonWindowsStub(t *testing.T) {
	result := NotifyShellRefresh(0)
	if result.EnvironmentTimeoutMS != 5000 {
		t.Fatalf("timeout=%d, want 5000", result.EnvironmentTimeoutMS)
	}
	if result.Warning == "" {
		t.Fatal("expected warning for non-windows stub")
	}
	if result.AssocChangedSent || result.EnvironmentBroadcast {
		t.Fatal("expected no API calls on non-Windows stub")
	}
}
