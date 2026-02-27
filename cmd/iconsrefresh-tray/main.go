//go:build windows

package main

import (
	"fmt"
	"os"

	"gioui.org/app"
	"gioui.org/unit"
)

func main() {
	go func() {
		w := new(app.Window)
		w.Option(app.Title(appTitle), app.Size(unit.Dp(windowWidthDp), unit.Dp(windowHeightDp)))
		if err := eventLoop(w); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(0)
	}()
	app.Main()
}
