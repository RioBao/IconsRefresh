package repair

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverCacheTargets(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tempDir)

	mustWriteFile(t, filepath.Join(tempDir, "Microsoft", "Windows", "Explorer", "iconcache_16.db"))
	mustWriteFile(t, filepath.Join(tempDir, "Microsoft", "Windows", "Explorer", "iconcache_64.db"))
	mustWriteFile(t, filepath.Join(tempDir, "Packages", "Microsoft.Windows.Search_123", "LocalState", "AppIconCache"))

	targets, err := DiscoverCacheTargets()
	if err != nil {
		t.Fatalf("DiscoverCacheTargets() error = %v", err)
	}

	if len(targets) != 4 {
		t.Fatalf("expected 4 targets, got %d", len(targets))
	}
}

func TestTargetsForMode(t *testing.T) {
	targets := []Target{
		{Path: "a", Kind: TargetIconCacheDB},
		{Path: "b", Kind: TargetExplorerCacheDB},
		{Path: "c", Kind: TargetSearchAppCache},
	}

	if got := len(TargetsForMode(targets, ModeQuick)); got != 1 {
		t.Fatalf("quick mode expected 1 target, got %d", got)
	}
	if got := len(TargetsForMode(targets, ModeSoft)); got != 1 {
		t.Fatalf("soft mode expected 1 target, got %d", got)
	}
	if got := len(TargetsForMode(targets, ModeStandard)); got != 2 {
		t.Fatalf("standard mode expected 2 targets, got %d", got)
	}
	if got := len(TargetsForMode(targets, ModeDeep)); got != 3 {
		t.Fatalf("deep mode expected 3 targets, got %d", got)
	}
}

func TestShouldRunIE4UInit(t *testing.T) {
	testCases := []struct {
		mode Mode
		want bool
	}{
		{mode: ModeQuick, want: true},
		{mode: ModeSoft, want: false},
		{mode: ModeStandard, want: true},
		{mode: ModeDeep, want: true},
	}

	for _, tc := range testCases {
		if got := ShouldRunIE4UInit(tc.mode); got != tc.want {
			t.Fatalf("ShouldRunIE4UInit(%q)=%v, want %v", tc.mode, got, tc.want)
		}
	}
}

func TestDeleteTargetsForMode(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tempDir)
	t.Setenv("PATH", t.TempDir())

	valid := filepath.Join(tempDir, "IconCache.db")
	mustWriteFile(t, valid)

	quickResult := DeleteTargetsForMode(ModeQuick, []Target{{Path: valid}})
	if quickResult.IE4UInit == nil {
		t.Fatal("expected ie4uinit result for quick mode")
	}
	if len(quickResult.Paths) != 1 || !quickResult.Paths[0].Deleted {
		t.Fatalf("unexpected quick deletion result: %+v", quickResult.Paths)
	}
	if quickResult.ShellNotify == nil {
		t.Fatal("expected shell notify result for quick mode")
	}

	mustWriteFile(t, valid)
	softResult := DeleteTargetsForMode(ModeSoft, []Target{{Path: valid}})
	if softResult.IE4UInit != nil {
		t.Fatal("did not expect ie4uinit result for soft mode")
	}
	if len(softResult.Paths) != 1 || !softResult.Paths[0].Deleted {
		t.Fatalf("unexpected soft deletion result: %+v", softResult.Paths)
	}
	if softResult.ShellNotify == nil {
		t.Fatal("expected shell notify result for soft mode")
	}
}

func TestValidateCandidate(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tempDir)

	iconCache := filepath.Join(tempDir, "IconCache.db")
	mustWriteFile(t, iconCache)

	if err := ValidateCandidate(iconCache); err != nil {
		t.Fatalf("expected valid candidate, got error %v", err)
	}

	outside := filepath.Join(tempDir, "..", "evil.db")
	mustWriteFile(t, outside)
	if err := ValidateCandidate(outside); err == nil {
		t.Fatal("expected error for outside LOCALAPPDATA path")
	}

	invalidName := filepath.Join(tempDir, "Microsoft", "Windows", "Explorer", "thumbcache_16.db")
	mustWriteFile(t, invalidName)
	if err := ValidateCandidate(invalidName); err == nil {
		t.Fatal("expected error for unsupported file name")
	}
}

func TestDeleteTargets(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("LOCALAPPDATA", tempDir)

	valid := filepath.Join(tempDir, "IconCache.db")
	missing := filepath.Join(tempDir, "Microsoft", "Windows", "Explorer", "iconcache_16.db")
	mustWriteFile(t, valid)

	result := DeleteTargets([]Target{{Path: valid}, {Path: missing}})
	if len(result.Paths) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Paths))
	}

	if !result.Paths[0].Found || !result.Paths[0].Deleted || result.Paths[0].Skipped {
		t.Fatalf("unexpected first result: %+v", result.Paths[0])
	}
	if !result.Paths[1].Skipped || result.Paths[1].Error == "" {
		t.Fatalf("unexpected second result: %+v", result.Paths[1])
	}

	if _, err := os.Stat(valid); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be removed", valid)
	}
}

func TestLocalAppDataPathFallback(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")
	got, err := localAppDataPath()
	if err != nil {
		t.Fatalf("localAppDataPath() error = %v", err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "AppData", "Local")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
