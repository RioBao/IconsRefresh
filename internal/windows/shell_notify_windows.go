//go:build windows

package windows

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	shcneAssocChanged = 0x08000000
	shcnfIDList       = 0x0000

	hwndBroadcast   = 0xFFFF
	wmSettingChange = 0x001A
	smtoAbortIfHung = 0x0002
)

// ShellNotifyResult reports shell refresh API call outcomes.
type ShellNotifyResult struct {
	AssocChangedSent     bool
	EnvironmentBroadcast bool
	EnvironmentTimeoutMS uint32
	Warning              string
}

// NotifyShellRefresh asks Explorer/Shell to refresh icon associations and environment-driven updates.
func NotifyShellRefresh(timeoutMS uint32) ShellNotifyResult {
	result := ShellNotifyResult{EnvironmentTimeoutMS: timeoutMS}

	if timeoutMS == 0 {
		timeoutMS = 5000
		result.EnvironmentTimeoutMS = timeoutMS
	}

	if _, _, err := syscall.NewLazyDLL("shell32.dll").NewProc("SHChangeNotify").Call(
		uintptr(shcneAssocChanged),
		uintptr(shcnfIDList),
		0,
		0,
	); err == syscall.Errno(0) {
		result.AssocChangedSent = true
	} else {
		result.Warning = fmt.Sprintf("SHChangeNotify failed: %v", err)
		return result
	}

	env, envErr := syscall.UTF16PtrFromString("Environment")
	if envErr != nil {
		result.Warning = fmt.Sprintf("UTF16 conversion failed: %v", envErr)
		return result
	}

	if _, _, err := syscall.NewLazyDLL("user32.dll").NewProc("SendMessageTimeoutW").Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(smtoAbortIfHung),
		uintptr(timeoutMS),
		0,
	); err == syscall.Errno(0) {
		result.EnvironmentBroadcast = true
	} else {
		result.Warning = fmt.Sprintf("SendMessageTimeoutW failed: %v", err)
	}

	return result
}
