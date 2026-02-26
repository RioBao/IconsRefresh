//go:build !windows

package windows

// ShellNotifyResult reports shell refresh API call outcomes.
type ShellNotifyResult struct {
	AssocChangedSent     bool
	EnvironmentBroadcast bool
	EnvironmentTimeoutMS uint32
	Warning              string
}

// NotifyShellRefresh is a non-Windows stub used by tests.
func NotifyShellRefresh(timeoutMS uint32) ShellNotifyResult {
	if timeoutMS == 0 {
		timeoutMS = 5000
	}
	return ShellNotifyResult{
		EnvironmentTimeoutMS: timeoutMS,
		Warning:              "shell refresh notification is only available on Windows",
	}
}
