package windows

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

const (
	ie4uinitExecutable = "ie4uinit.exe"
	ie4uinitShowArg    = "-show"
)

// IE4UInitResult captures command execution details for ie4uinit.exe -show.
type IE4UInitResult struct {
	Command  string
	Args     []string
	Stdout   string
	Stderr   string
	ExitCode int
	Warning  string
	Ran      bool
}

// RunIE4UInitShow locates and invokes ie4uinit.exe -show.
//
// Command-not-found is treated as a non-fatal compatibility warning.
func RunIE4UInitShow() IE4UInitResult {
	result := IE4UInitResult{
		Command:  ie4uinitExecutable,
		Args:     []string{ie4uinitShowArg},
		ExitCode: -1,
	}

	resolvedPath, err := exec.LookPath(ie4uinitExecutable)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			result.Warning = fmt.Sprintf("%s not found in PATH; skipping compatibility step", ie4uinitExecutable)
			return result
		}
		result.Warning = fmt.Sprintf("unable to resolve %s: %v", ie4uinitExecutable, err)
		return result
	}

	result.Command = resolvedPath

	cmd := exec.Command(resolvedPath, ie4uinitShowArg)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	result.Ran = true
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			result.Warning = fmt.Sprintf("%s not found in PATH; skipping compatibility step", ie4uinitExecutable)
			result.Ran = false
			result.ExitCode = -1
			return result
		}
		if _, ok := err.(*exec.ExitError); ok {
			return result
		}
		result.Warning = fmt.Sprintf("unable to execute %s: %v", ie4uinitExecutable, err)
	}

	return result
}
