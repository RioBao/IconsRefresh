//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"image/png"

	"gioui.org/op/paint"
)

//go:embed icon.png
var uiIconPNG []byte

var (
	uiIconOp  paint.ImageOp
	hasUIIcon bool
)

func init() {
	if len(uiIconPNG) == 0 {
		return
	}
	img, err := png.Decode(bytes.NewReader(uiIconPNG))
	if err != nil {
		return
	}
	uiIconOp = paint.NewImageOp(img)
	hasUIIcon = true
}
