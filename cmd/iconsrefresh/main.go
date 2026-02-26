package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/crazy-max/IconsRefresh/internal/engine"
	"github.com/crazy-max/IconsRefresh/internal/repair"
)

type config struct {
	Preset engine.Preset
	DryRun bool
	JSON   bool
}

var errUsage = errors.New("usage error")

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(2)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (config, error) {
	cfg := config{}
	fs := flag.NewFlagSet("IconsRefresh", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Discover and print targets without deleting")
	fs.BoolVar(&cfg.JSON, "json", false, "Output structured JSON result")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	positional := fs.Args()
	if len(positional) == 0 {
		return config{}, fmt.Errorf("%w: missing required <mode>", errUsage)
	}
	if len(positional) > 1 {
		return config{}, fmt.Errorf("%w: expected one <mode>, got %d", errUsage, len(positional))
	}

	switch strings.ToLower(strings.TrimSpace(positional[0])) {
	case "quick":
		cfg.Preset = engine.PresetCLIQuick
	case "soft":
		cfg.Preset = engine.PresetCLISoft
	case "standard":
		cfg.Preset = engine.PresetCLIStandard
	case "deep":
		cfg.Preset = engine.PresetCLIDeep
	default:
		return config{}, fmt.Errorf("%w: invalid mode %q", errUsage, positional[0])
	}

	return cfg, nil
}

func usage() string {
	return strings.TrimSpace(`Usage: IconsRefresh [--dry-run] [--json] <mode>

Modes:
  quick      Run ie4uinit + shell notify, clean IconCache.db
  soft       Clean IconCache.db only
  standard   Clean IconCache.db and Explorer iconcache_*.db
  deep       Standard mode + Search AppIconCache cleanup`)
}

func run(cfg config) error {
	request, err := engine.RequestForPreset(cfg.Preset, cfg.DryRun)
	if err != nil {
		return err
	}

	eng := engine.New(engine.Hooks{})
	result, err := eng.Run(context.Background(), request)
	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if encodeErr := enc.Encode(result); encodeErr != nil {
			return encodeErr
		}
	}
	if err != nil {
		return err
	}

	if cfg.JSON {
		return resultError(result.Result)
	}

	fmt.Printf("mode=%s dry-run=%t targets=%d\n", result.Mode, result.DryRun, len(result.Targets))
	if result.Result.IE4UInit != nil {
		fmt.Printf("ie4uinit: ran=%t exit=%d warning=%q\n", result.Result.IE4UInit.Ran, result.Result.IE4UInit.ExitCode, result.Result.IE4UInit.Warning)
	}
	if result.Result.ShellNotify != nil {
		fmt.Printf("shell_notify: assoc_changed=%t env_broadcast=%t timeout_ms=%d warning=%q\n",
			result.Result.ShellNotify.AssocChangedSent,
			result.Result.ShellNotify.EnvironmentBroadcast,
			result.Result.ShellNotify.EnvironmentTimeoutMS,
			result.Result.ShellNotify.Warning,
		)
	}
	for _, p := range result.Result.Paths {
		fmt.Printf("path=%q found=%t deleted=%t skipped=%t", p.Path, p.Found, p.Deleted, p.Skipped)
		if p.Error != "" {
			fmt.Printf(" error=%q", p.Error)
		}
		fmt.Println()
	}

	return resultError(result.Result)
}

// isDeletionFailure returns true for errors that occurred on a target that was
// found and validated — i.e., a real deletion failure, not a "not found" skip.
func isDeletionFailure(p repair.PathResult) bool {
	return p.Error != "" && !(p.Skipped && !p.Found)
}

// resultError returns a non-nil error if any deletion failures are present.
func resultError(result repair.Result) error {
	count := 0
	for _, p := range result.Paths {
		if isDeletionFailure(p) {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	return fmt.Errorf("cache refresh completed with %d path error(s)", count)
}
