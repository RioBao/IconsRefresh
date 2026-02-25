package windows

import "testing"

func TestRunIE4UInitShow_NotFoundIsWarning(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	result := RunIE4UInitShow()
	if result.Warning == "" {
		t.Fatal("expected warning when ie4uinit.exe is missing")
	}
	if result.Ran {
		t.Fatal("expected command not to run when ie4uinit.exe is missing")
	}
	if result.ExitCode != -1 {
		t.Fatalf("expected exit code -1 when skipped, got %d", result.ExitCode)
	}
}
