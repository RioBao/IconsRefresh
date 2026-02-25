//go:generate go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=.github/logo.ico
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"github.com/crazy-max/IconsRefresh/internal/repair"
	"golang.org/x/sys/windows"
)

const (
	SHCNE_ASSOCCHANGED = 0x08000000
	SHCNF_IDLIST       = 0x0000

	HWND_BROADCAST   = 0xFFFF
	WM_SETTINGCHANGE = 0x001A
	SMTO_ABORTIFHUNG = 0x0002
)

const (
	exitSuccess  = 0
	exitFailed   = 1
	exitBadUsage = 2
)

type appConfig struct {
	Mode   string
	DryRun bool
	JSON   bool
}

type stepReport struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Duration  int64     `json:"duration_ms"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Details   string    `json:"details,omitempty"`
	Error     string    `json:"error,omitempty"`
}

type runReport struct {
	Mode      string       `json:"mode"`
	DryRun    bool         `json:"dry_run"`
	StartedAt time.Time    `json:"started_at"`
	EndedAt   time.Time    `json:"ended_at"`
	Steps     []stepReport `json:"steps"`
	Succeeded bool         `json:"succeeded"`
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(exitBadUsage)
	}

	report := executePipeline(cfg)
	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "encode json report: %v\n", err)
			os.Exit(exitFailed)
		}
	} else {
		printReport(report)
	}

	if report.Succeeded {
		os.Exit(exitSuccess)
	}
	os.Exit(exitFailed)
}

func parseArgs(args []string) (appConfig, error) {
	cfg := appConfig{}
	fs := flag.NewFlagSet("iconsrefresh", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "print actions without mutating state")
	fs.BoolVar(&cfg.JSON, "json", false, "emit machine-readable report")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}

	rest := fs.Args()
	if len(rest) > 1 {
		return cfg, fmt.Errorf("expected one mode argument, got %q", strings.Join(rest, " "))
	}
	if len(rest) == 0 {
		cfg.Mode = "standard"
	} else {
		cfg.Mode = strings.ToLower(rest[0])
	}

	switch cfg.Mode {
	case "quick", "standard", "deep":
		return cfg, nil
	default:
		return cfg, fmt.Errorf("invalid mode %q (must be quick, standard, or deep)", cfg.Mode)
	}
}

func usage() string {
	return `Usage: IconsRefresh [--dry-run] [--json] <mode>

Modes:
  quick     Run shell notify + ie4uinit.exe -show only.
  standard  quick + Explorer iconcache DB cleanup.
  deep      standard + Search AppIconCache cleanup.

Flags:
  --dry-run Print what would happen without mutation.
  --json    Print machine-readable report.`
}

func executePipeline(cfg appConfig) runReport {
	report := runReport{
		Mode:      cfg.Mode,
		DryRun:    cfg.DryRun,
		StartedAt: time.Now().UTC(),
		Steps:     make([]stepReport, 0, 6),
	}

	runner := func(name string, details string, fn func() error) {
		step := stepReport{Name: name, Status: "ok", Details: details, StartedAt: time.Now().UTC()}
		err := fn()
		step.EndedAt = time.Now().UTC()
		step.Duration = step.EndedAt.Sub(step.StartedAt).Milliseconds()
		if err != nil {
			step.Status = "error"
			step.Error = err.Error()
		}
		report.Steps = append(report.Steps, step)
	}

	runner("shell_notify", "SHChangeNotify + SendMessageTimeoutW", func() error {
		if cfg.DryRun {
			return nil
		}
		shellNotify()
		return nil
	})

	runner("ie4uinit_show", "ie4uinit.exe -show", func() error {
		if cfg.DryRun {
			return nil
		}
		cmd := exec.Command("ie4uinit.exe", "-show")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("run ie4uinit.exe -show: %w (%s)", err, strings.TrimSpace(string(output)))
		}
		return nil
	})

	if cfg.Mode == "standard" || cfg.Mode == "deep" {
		runner("cleanup_targets", "remove discovered icon cache database files", func() error {
			return cleanupTargets(cfg)
		})
	}

	report.EndedAt = time.Now().UTC()
	report.Succeeded = true
	for _, step := range report.Steps {
		if step.Status == "error" {
			report.Succeeded = false
			break
		}
	}

	return report
}

func cleanupTargets(cfg appConfig) error {
	targets, err := repair.DiscoverCacheTargets()
	if err != nil {
		return fmt.Errorf("discover cache targets: %w", err)
	}

	wanted := make(map[repair.TargetKind]struct{}, 2)
	wanted[repair.TargetExplorerCacheDB] = struct{}{}
	if cfg.Mode == "deep" {
		wanted[repair.TargetSearchAppCache] = struct{}{}
	}

	selected := make([]repair.Target, 0, len(targets))
	for _, target := range targets {
		if _, ok := wanted[target.Kind]; ok {
			selected = append(selected, target)
		}
	}

	if cfg.DryRun {
		for _, target := range selected {
			if err := repair.ValidateCandidate(target.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("validate %s: %w", target.Path, err)
			}
		}
		return nil
	}

	result := repair.DeleteTargets(selected)
	var failed []string
	for _, pathResult := range result.Paths {
		if pathResult.Error != "" && pathResult.Error != "not found" {
			failed = append(failed, fmt.Sprintf("%s: %s", pathResult.Path, pathResult.Error))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("cleanup failures: %s", strings.Join(failed, "; "))
	}
	return nil
}

func printReport(report runReport) {
	for _, step := range report.Steps {
		if step.Status == "ok" {
			fmt.Printf("[OK]   %s (%s)\n", step.Name, step.Details)
			continue
		}
		fmt.Printf("[FAIL] %s: %s\n", step.Name, step.Error)
	}

	if report.Succeeded {
		fmt.Printf("Completed mode=%s dry-run=%t\n", report.Mode, report.DryRun)
		return
	}
	fmt.Printf("Failed mode=%s dry-run=%t\n", report.Mode, report.DryRun)
}

func shellNotify() {
	// https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/nf-shlobj_core-shchangenotify
	windows.NewLazyDLL("shell32.dll").NewProc("SHChangeNotify").Call(
		uintptr(SHCNE_ASSOCCHANGED),
		uintptr(SHCNF_IDLIST),
		0, 0)

	// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-sendmessagetimeoutw
	env, _ := windows.UTF16PtrFromString("Environment")
	windows.NewLazyDLL("user32.dll").NewProc("SendMessageTimeoutW").Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000))
}
