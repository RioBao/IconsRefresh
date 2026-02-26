//go:build !windows

package windows

// ExplorerRestartResult captures the outcome of stopping and restarting explorer.exe.
type ExplorerRestartResult struct {
	Stopped   bool
	Restarted bool
	Warning   string
}

// StopExplorer is a non-Windows stub.
func StopExplorer() ExplorerRestartResult {
	return ExplorerRestartResult{Warning: "Explorer restart is only available on Windows"}
}

// StartExplorer is a non-Windows stub.
func StartExplorer() ExplorerRestartResult {
	return ExplorerRestartResult{Warning: "Explorer restart is only available on Windows"}
}
