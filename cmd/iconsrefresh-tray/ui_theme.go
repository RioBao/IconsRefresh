//go:build windows

package main

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

var (
	colBg      = color.NRGBA{R: 0xF3, G: 0xF3, B: 0xF3, A: 0xFF}
	colSurface = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	colBorder  = color.NRGBA{R: 0xE5, G: 0xE5, B: 0xE5, A: 0xFF}
	colHeader  = color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF}

	colBtnBg     = color.NRGBA{R: 233, G: 238, B: 244, A: 255}
	colBtnFg     = color.NRGBA{R: 32, G: 34, B: 38, A: 255}
	colBtnHover  = color.NRGBA{R: 220, G: 227, B: 235, A: 255}
	colBtnPress  = color.NRGBA{R: 210, G: 218, B: 228, A: 255}
	colBtnBorder = color.NRGBA{R: 200, G: 205, B: 212, A: 255}

	colDisabledBg   = color.NRGBA{R: 0xE8, G: 0xE8, B: 0xE8, A: 0xFF}
	colDisabledText = color.NRGBA{R: 0xAA, G: 0xAA, B: 0xAA, A: 0xFF}

	colText1 = color.NRGBA{R: 0x1C, G: 0x1C, B: 0x1C, A: 0xFF}
	colText2 = color.NRGBA{R: 0x5F, G: 0x5F, B: 0x5F, A: 0xFF}
	colText3 = color.NRGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF}

	colSuccess = color.NRGBA{R: 0x2E, G: 0x7D, B: 0x47, A: 0xFF}
	colFail    = color.NRGBA{R: 0xB0, G: 0x30, B: 0x30, A: 0xFF}
	colWarn    = color.NRGBA{R: 0xB0, G: 0x6A, B: 0x00, A: 0xFF}
	colStep    = color.NRGBA{R: 0x60, G: 0x74, B: 0x88, A: 0xFF}
)

func logLineColor(line string) color.NRGBA {
	switch {
	case strings.HasPrefix(line, "✓"):
		return colSuccess
	case strings.HasPrefix(line, "✗"):
		return colFail
	case strings.HasPrefix(line, "⚠"):
		return colWarn
	case strings.HasPrefix(line, "-"):
		return colStep
	default:
		return colText2
	}
}

func fillRRect(ops *op.Ops, rect image.Rectangle, r int, col color.NRGBA) {
	defer clip.RRect{Rect: rect, SE: r, SW: r, NW: r, NE: r}.Push(ops).Pop()
	paint.Fill(ops, col)
}
