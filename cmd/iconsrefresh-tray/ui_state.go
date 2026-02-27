//go:build windows

package main

import (
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/crazy-max/IconsRefresh/internal/engine"
)

const (
	appTitle       = "IconsRefresh"
	appSubtitle    = "Refresh Windows icon caches"
	windowWidthDp  = 320
	windowHeightDp = 400
)

var modes = [3]struct {
	label       string
	preset      engine.Preset
	description string
}{
	{"Quick", engine.PresetTrayQuick, "Fast cleanup that refreshes icons without restarting Explorer."},
	{"Standard", engine.PresetTrayStandard, "Recommended cleanup that clears more cache files and restarts Explorer."},
	{"Deep", engine.PresetTrayDeep, "Most thorough cleanup: restarts Explorer and also clears Search app icon cache."},
}

type runResult struct {
	logLines []string
	footer   string
}

type uiState struct {
	btns     [3]widget.Clickable
	running  bool
	logLines []string
	footer   string
	logList  widget.List
	resultC  chan runResult
}

func newUIState() *uiState {
	return &uiState{
		logList: widget.List{List: layout.List{Axis: layout.Vertical}},
		resultC: make(chan runResult, 1),
	}
}
