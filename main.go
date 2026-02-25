//go:build windows
// +build windows

//go:generate go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=.github/logo.ico
package main

import (
	"fmt"
	"os"
)

const (
	exitSuccess  = 0
	exitFailed   = 1
	exitBadUsage = 2
)

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(exitBadUsage)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(exitFailed)
	}

	os.Exit(exitSuccess)
}
