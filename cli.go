package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/crazy-max/IconsRefresh/internal/repair"
)

type config struct {
	Mode   repair.Mode
	DryRun bool
	JSON   bool
}

var errUsage = errors.New("usage error")

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

	mode := repair.Mode(strings.ToLower(strings.TrimSpace(positional[0])))
	switch mode {
	case repair.ModeQuick, repair.ModeSoft, repair.ModeStandard, repair.ModeDeep:
		cfg.Mode = mode
	default:
		return config{}, fmt.Errorf("%w: invalid mode %q", errUsage, positional[0])
	}

	return cfg, nil
}

func usage() string {
	return strings.TrimSpace(`Usage: IconsRefresh [--dry-run] [--json] <mode>

Modes:
  quick      Run shell refresh and clean IconCache.db only
  soft       Clean IconCache.db only
  standard   Clean IconCache.db and Explorer iconcache_*.db
  deep       Standard mode + Search AppIconCache cleanup`)
}

func run(cfg config) error {
	targets, err := repair.DiscoverCacheTargets()
	if err != nil {
		return fmt.Errorf("discover cache targets: %w", err)
	}

	selected := repair.TargetsForMode(targets, cfg.Mode)
	if cfg.DryRun {
		return printDryRun(cfg, selected)
	}

	result := repair.DeleteTargetsForMode(cfg.Mode, selected)
	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if result.IE4UInit != nil {
		fmt.Printf("ie4uinit: ran=%t exit=%d warning=%q\n", result.IE4UInit.Ran, result.IE4UInit.ExitCode, result.IE4UInit.Warning)
	}
	for _, p := range result.Paths {
		fmt.Printf("path=%q found=%t deleted=%t skipped=%t", p.Path, p.Found, p.Deleted, p.Skipped)
		if p.Error != "" {
			fmt.Printf(" error=%q", p.Error)
		}
		fmt.Println()
	}

	return nil
}

func printDryRun(cfg config, selected []repair.Target) error {
	if cfg.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(selected)
	}

	fmt.Printf("mode=%s dry-run=true targets=%d\n", cfg.Mode, len(selected))
	for _, t := range selected {
		fmt.Printf("%s\t%s\n", t.Kind, t.Path)
	}
	return nil
}
