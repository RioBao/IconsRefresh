//go:build windows

package windows

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// ExplorerRestartResult captures the outcome of stopping and restarting explorer.exe.
type ExplorerRestartResult struct {
	Stopped   bool
	Restarted bool
	Warning   string
}

// StopExplorer terminates the current user's explorer.exe via taskkill and waits
// briefly for the OS to release its file handles.
func StopExplorer() ExplorerRestartResult {
	result := ExplorerRestartResult{}

	out, err := exec.Command("taskkill", "/f", "/im", "explorer.exe").CombinedOutput()
	if err != nil {
		result.Warning = fmt.Sprintf("taskkill explorer.exe: %v (%s)", err, bytes.TrimSpace(out))
		return result
	}
	result.Stopped = true

	// Brief pause for the OS to close Explorer's file handles before the caller
	// attempts to delete the now-unlocked cache files.
	time.Sleep(500 * time.Millisecond)
	return result
}

// StartExplorer launches explorer.exe for the current session.
func StartExplorer() ExplorerRestartResult {
	result := ExplorerRestartResult{}

	cmd := exec.Command("explorer.exe")
	if err := cmd.Start(); err != nil {
		result.Warning = fmt.Sprintf("start explorer.exe: %v", err)
		return result
	}
	// Detach: we do not wait for Explorer to exit.
	_ = cmd.Process.Release()
	result.Restarted = true
	return result
}
