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

// NotifyShellRefresh asks Explorer to refresh icon associations and broadcasts an
// environment change. timeoutMS is the SendMessageTimeoutW deadline; pass 0 for 5000ms.
func NotifyShellRefresh(timeoutMS uint32) ShellNotifyResult {
	result := ShellNotifyResult{EnvironmentTimeoutMS: timeoutMS}
	if timeoutMS == 0 {
		timeoutMS = 5000
		result.EnvironmentTimeoutMS = timeoutMS
	}

	// https://docs.microsoft.com/en-us/windows/desktop/api/shlobj_core/nf-shlobj_core-shchangenotify
	if _, _, err := syscall.NewLazyDLL("shell32.dll").NewProc("SHChangeNotify").Call(
		uintptr(shcneAssocChanged),
		uintptr(shcnfIDList),
		0,
		0,
	); err != syscall.Errno(0) {
		result.Warning = fmt.Sprintf("SHChangeNotify failed: %v", err)
		return result
	}
	result.AssocChangedSent = true

	env, envErr := syscall.UTF16PtrFromString("Environment")
	if envErr != nil {
		result.Warning = fmt.Sprintf("UTF16 conversion failed: %v", envErr)
		return result
	}

	// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-sendmessagetimeoutw
	if _, _, err := syscall.NewLazyDLL("user32.dll").NewProc("SendMessageTimeoutW").Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(smtoAbortIfHung),
		uintptr(timeoutMS),
		0,
	); err != syscall.Errno(0) {
		result.Warning = fmt.Sprintf("SendMessageTimeoutW failed: %v", err)
		return result
	}
	result.EnvironmentBroadcast = true

	return result
}
