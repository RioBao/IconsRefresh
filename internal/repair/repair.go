package repair

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	internalwindows "github.com/crazy-max/IconsRefresh/internal/windows"
)

// Mode defines how aggressive cache repair should be.
type Mode string

const (
	ModeQuick    Mode = "quick"
	ModeSoft     Mode = "soft"
	ModeStandard Mode = "standard"
	ModeDeep     Mode = "deep"
)

// TargetKind describes a discovered cache target family.
type TargetKind string

const (
	TargetIconCacheDB     TargetKind = "iconcache_db"
	TargetExplorerCacheDB TargetKind = "explorer_iconcache_db"
	TargetSearchAppCache  TargetKind = "search_app_iconcache"
)

// Target represents one cache file candidate.
type Target struct {
	Path string
	Kind TargetKind
}

// PathResult is a per-path outcome that can be rendered by a CLI.
type PathResult struct {
	Path    string
	Found   bool
	Deleted bool
	Skipped bool
	Error   string
}

// Result contains path-level outcomes for a repair execution.
type Result struct {
	IE4UInit    *internalwindows.IE4UInitResult
	ShellNotify *internalwindows.ShellNotifyResult
	Paths       []PathResult
}

// DiscoverCacheTargets finds known Windows icon cache paths in LOCALAPPDATA.
func DiscoverCacheTargets() ([]Target, error) {
	localAppData, err := localAppDataPath()
	if err != nil {
		return nil, err
	}

	targets := make([]Target, 0, 8)
	seen := make(map[string]struct{})

	add := func(path string, kind TargetKind) {
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		targets = append(targets, Target{Path: clean, Kind: kind})
	}

	add(filepath.Join(localAppData, "IconCache.db"), TargetIconCacheDB)

	explorerMatches, err := filepath.Glob(filepath.Join(localAppData, "Microsoft", "Windows", "Explorer", "iconcache_*.db"))
	if err != nil {
		return nil, fmt.Errorf("glob explorer icon cache: %w", err)
	}
	for _, match := range explorerMatches {
		add(match, TargetExplorerCacheDB)
	}

	searchMatches, err := filepath.Glob(filepath.Join(localAppData, "Packages", "Microsoft.Windows.Search_*", "LocalState", "AppIconCache"))
	if err != nil {
		return nil, fmt.Errorf("glob search app icon cache: %w", err)
	}
	for _, match := range searchMatches {
		add(match, TargetSearchAppCache)
	}

	return targets, nil
}

// TargetsForMode filters discovered targets by repair mode.
func TargetsForMode(targets []Target, mode Mode) []Target {
	filtered := make([]Target, 0, len(targets))
	for _, target := range targets {
		if modeIncludesKind(mode, target.Kind) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

func modeIncludesKind(mode Mode, kind TargetKind) bool {
	switch mode {
	case ModeQuick:
		return kind == TargetIconCacheDB
	case ModeSoft:
		return kind == TargetIconCacheDB
	case ModeStandard:
		return kind == TargetIconCacheDB || kind == TargetExplorerCacheDB
	case ModeDeep:
		return kind == TargetIconCacheDB || kind == TargetExplorerCacheDB || kind == TargetSearchAppCache
	default:
		return false
	}
}

// ShouldRunIE4UInit reports whether the ie4uinit compatibility refresh should run for the mode.
func ShouldRunIE4UInit(mode Mode) bool {
	switch mode {
	case ModeQuick, ModeStandard, ModeDeep:
		return true
	default:
		return false
	}
}

// DeleteTargetsForMode executes mode-specific pre-delete actions and then deletes targets.
func DeleteTargetsForMode(mode Mode, targets []Target) Result {
	result := Result{}
	if ShouldRunIE4UInit(mode) {
		ie4uinitResult := internalwindows.RunIE4UInitShow()
		result.IE4UInit = &ie4uinitResult
	}

	deletion := DeleteTargets(targets)
	result.Paths = deletion.Paths

	notifyResult := internalwindows.NotifyShellRefresh(5000)
	result.ShellNotify = &notifyResult
	return result
}

// ValidateCandidate ensures a path is present and matches one of the expected safe locations.
func ValidateCandidate(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	if strings.Contains(filepath.ToSlash(path), "..") {
		return errors.New("path traversal detected")
	}

	cleaned := filepath.Clean(path)
	absolutePath, err := filepath.Abs(cleaned)
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	localAppData, err := localAppDataPath()
	if err != nil {
		return err
	}
	localAppData, err = filepath.Abs(localAppData)
	if err != nil {
		return fmt.Errorf("get LOCALAPPDATA absolute path: %w", err)
	}

	if !strings.HasPrefix(strings.ToLower(absolutePath), strings.ToLower(localAppData)+string(os.PathSeparator)) && !strings.EqualFold(absolutePath, localAppData) {
		return fmt.Errorf("path %q is outside LOCALAPPDATA", absolutePath)
	}

	if !isAllowedTargetPath(absolutePath, localAppData) {
		return fmt.Errorf("path %q is not a known icon cache target", absolutePath)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.ErrNotExist
		}
		return fmt.Errorf("stat path: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path %q is a directory", absolutePath)
	}

	return nil
}

func isAllowedTargetPath(path, localAppData string) bool {
	path = filepath.Clean(path)
	localAppData = filepath.Clean(localAppData)

	if strings.EqualFold(path, filepath.Join(localAppData, "IconCache.db")) {
		return true
	}

	explorerDir := filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")
	if strings.EqualFold(filepath.Dir(path), explorerDir) {
		base := strings.ToLower(filepath.Base(path))
		if ok, _ := filepath.Match("iconcache_*.db", base); ok {
			return true
		}
	}

	packagesDir := filepath.Join(localAppData, "Packages")
	if !strings.EqualFold(filepath.Base(path), "AppIconCache") {
		return false
	}
	parent := filepath.Dir(path)
	if !strings.EqualFold(filepath.Base(parent), "LocalState") {
		return false
	}
	packageDir := filepath.Dir(parent)
	if !strings.HasPrefix(strings.ToLower(filepath.Base(packageDir)), strings.ToLower("Microsoft.Windows.Search_")) {
		return false
	}
	return strings.EqualFold(filepath.Dir(packageDir), packagesDir)
}

// DeleteTargets validates and attempts deletion for each candidate.
func DeleteTargets(targets []Target) Result {
	result := Result{Paths: make([]PathResult, 0, len(targets))}
	for _, target := range targets {
		entry := PathResult{Path: target.Path}
		if err := ValidateCandidate(target.Path); err != nil {
			entry.Skipped = true
			if errors.Is(err, os.ErrNotExist) {
				entry.Error = "not found"
			} else {
				entry.Error = err.Error()
			}
			result.Paths = append(result.Paths, entry)
			continue
		}

		entry.Found = true
		if err := os.Remove(target.Path); err != nil {
			entry.Error = err.Error()
			result.Paths = append(result.Paths, entry)
			continue
		}
		entry.Deleted = true
		result.Paths = append(result.Paths, entry)
	}
	return result
}

func localAppDataPath() (string, error) {
	if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
		return localAppData, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	if homeDir == "" {
		return "", errors.New("empty home directory")
	}
	return filepath.Join(homeDir, "AppData", "Local"), nil
}
